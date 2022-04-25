package hmyload

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/pkg/rpc"
	"github.com/mjmar01/harmolytics/pkg/types"
	"math"
	"math/big"
	"strings"
)

const (
	transactionReceiptMethod = "hmyv2_getTransactionReceipt"
	transactionByHashMethod  = "hmyv2_getTransactionByHash"
	transactionCountMethod   = "hmyv2_getTransactionsCount"
	transactionHistoryMethod = "hmyv2_getTransactionsHistory"
)

// GetTransactionsByWallet returns a list of all successful Transaction for a given types.Address
func (l *Loader) GetTransactionsByWallet(addr types.Address) (txs []types.Transaction, err error) {
	// Get total number of transactions and prepare slice with a bit overhead
	c, err := l.defaultConn.Call(transactionCountMethod, addr.OneAddress, rpc.AllTx)
	if err != nil {
		return
	}
	txCount := int(c.(float64) * 1.01)
	// Split into groups and let pages overlap a bit
	pageSize, overlap := 50000, 50
	// Get histories
	uniqueHashes := map[string]bool{}
	for i := 0; i < txCount; i += pageSize {
		res, err := l.defaultConn.RawCall(transactionHistoryMethod, map[string]interface{}{
			"address":   addr.OneAddress,
			"pageIndex": i / pageSize,
			"pageSize":  pageSize + overlap,
			"fullTx":    false,
			"txType":    rpc.AllTx,
			"order":     "ASC",
		})
		if err != nil {
			return nil, err
		}
		// Sort hashes into a map
		hashes, err := readTxHistory(res)
		if err != nil {
			return nil, err
		}
		for _, hash := range hashes {
			uniqueHashes[hash] = true
		}
	}
	// Get transactions by hash
	hashes := make([]string, len(uniqueHashes))
	i := 0
	for hash := range uniqueHashes {
		hashes[i] = hash
		i++
	}
	// Split into pages to avoid node stress
	txs = make([]types.Transaction, len(hashes))
	uniqueMethods := map[string]bool{}
	for i := 0; i < len(hashes); i += 5000 {
		size := int(math.Min(float64(len(hashes)-i), 5000))
		txsPart, err := l.GetFullTransactions(hashes[i : i+size]...)
		if err != nil {
			return nil, err
		}
		for j, tx := range txsPart {
			txs[i+j] = tx
			if tx.Method.Signature != "" {
				uniqueMethods[tx.Method.Signature] = true
			}
		}
	}
	for sig := range uniqueMethods {
		_, ok := l.cache.GetMethod(sig)
		if !ok {
			m, err := l.GetMethod(sig)
			if err != nil {
				return nil, err
			}
			m.Signature = sig
			l.cache.SetMethod(&m)
		}
	}
	for _, tx := range txs {
		m, ok := l.cache.GetMethod(tx.Method.Signature)
		if ok {
			tx.Method = *m
		}
	}
	return
}

// GetFullTransactions returns transactions and their receipts for every given hash. Does not include method information (yet)
func (l *Loader) GetFullTransactions(hashes ...string) (txs []types.Transaction, err error) {
	// Prepare requests
	txs = make([]types.Transaction, len(hashes))
	txByEthHash, txByHash := map[string]*types.Transaction{}, map[string]*types.Transaction{}
	bodiesByConn, idx, foundInCache := make([][]rpc.Body, l.uniqueConnCount), 0, 0
	for _, hash := range hashes {
		if pTx, ok := l.cache.GetTransaction(hash); ok {
			// Cache hit
			txByHash[pTx.TxHash] = pTx
			txByEthHash[pTx.EthTxHash] = pTx
			foundInCache++
		} else {
			// Cache miss
			b := l.uniqueConns[idx].NewBody(transactionByHashMethod, hash)
			bodiesByConn[idx] = append(bodiesByConn[idx], b)
			b = l.uniqueConns[idx].NewBody(transactionReceiptMethod, hash)
			bodiesByConn[idx] = append(bodiesByConn[idx], b)
			idx++
			if idx == l.uniqueConnCount {
				idx = 0
			}
		}
	}
	// Do requests across unique nodes
	ch := make(chan goTx, len(hashes)-foundInCache)
	for i, conn := range l.uniqueConns {
		go func(rpc *rpc.RPC, bodies []rpc.Body) {
			ress, err := rpc.RawBatchCall(bodies)
			if err != nil {
				ch <- goTx{err: err}
				return
			}
			// Read each result into a transaction
			for i := 0; i < len(ress); i += 2 {
				tx, err := readTxInfoFromResponse(ress[i])
				if err != nil {
					ch <- goTx{err: err}
					return
				}
				tx.Status, tx.Logs, err = readTxReceiptFromResponse(ress[i+1])
				if err != nil {
					ch <- goTx{err: err}
					return
				}
				m, ok := l.cache.GetMethod(tx.Method.Signature)
				if ok {
					tx.Method = *m
				}
				ch <- goTx{
					err: nil,
					tx:  tx,
				}
			}
		}(conn, bodiesByConn[i])
	}
	// Read output
	for i := foundInCache; i < len(hashes); i++ {
		out := <-ch
		if out.err != nil {
			return txs, out.err
		}
		txByHash[out.tx.TxHash] = out.tx
		txByEthHash[out.tx.EthTxHash] = out.tx
		l.cache.SetTransaction(out.tx)
	}
	for i, hash := range hashes {
		txPtr, ok := txByEthHash[hash]
		if !ok {
			txPtr, ok = txByHash[hash]
			if !ok {
				return nil, errors.Errorf("Given hash (%s) was not found in list of results", hash)
			}
		}
		txs[i] = *txPtr
	}
	return
}

func readTxInfoFromResponse(data []byte) (tx *types.Transaction, err error) {
	// Read JSON into transaction
	var t transactionInfoJson
	err = json.Unmarshal(data, &t)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	value := new(big.Int)
	value.SetString(t.Value.String(), 10)
	gasPrice := new(big.Int)
	gasPrice.SetString(t.GasPrice.String(), 10)
	var method types.Method
	if len(t.Input) >= 10 {
		method.Signature = t.Input[2:10]
	} else {
		method.Signature = ""
	}
	tx = &types.Transaction{
		TxHash:    t.TxHash,
		EthTxHash: t.EthTxHash,
		Sender:    types.NewAddress(t.Sender),
		Receiver:  types.NewAddress(t.Receiver),
		BlockNum:  t.BlockNum,
		Timestamp: t.Timestamp,
		Value:     value,
		Method:    method,
		Input:     t.Input,
		GasAmount: uint32(t.GasAmount),
		GasPrice:  gasPrice,
		ShardID:   t.ShardID,
		ToShardID: t.ToShardID,
	}
	return
}

func readTxReceiptFromResponse(data []byte) (s int, ls []types.TransactionLog, err error) {
	// Read JSON to get tx status and logs
	var t transactionReceiptJson
	err = json.Unmarshal(data, &t)
	if err != nil {
		return 0, nil, errors.Wrap(err, 0)
	}
	s = t.Status
	for _, l := range t.Logs {
		// Convert data formats
		index, err := hexutil.DecodeUint64(l.LogIndex)
		if err != nil {
			return 0, nil, errors.Wrap(err, 0)
		}
		// Add logs
		ls = append(ls, types.TransactionLog{
			TxHash:   t.TxHash,
			LogIndex: int(index),
			Address:  types.NewAddress(l.Address),
			Topics:   l.Topics,
			Data:     l.Data,
		})
	}
	return
}

func readTxHistory(data []byte) (txs []string, err error) {
	// Cutoff wrapper
	data = data[17 : len(data)-2]
	// Split into single txs
	txs = strings.Split(strings.ReplaceAll(string(data), "\"", ""), ",")
	return
}
