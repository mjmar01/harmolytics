package hmyload

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/pkg/harmony"
	"github.com/mjmar01/harmolytics/pkg/rpc"
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

// GetTransactionsByWallet returns a list of all successful Transaction for a given harmony.Address
func (l *Loader) GetTransactionsByWallet(addr harmony.Address) (txs []harmony.Transaction, err error) {
	// Get total number of transactions and prepare slice with a bit overhead
	c, err := l.defaultConn.Call(transactionCountMethod, addr.OneAddress, rpc.AllTx)
	if err != nil {
		return
	}
	txCount := int(c.(float64) * 1.01)
	// Split into groups and let pages overlap a bit
	pageSize, overlap := 50000, 50
	// Get histories
	txMap := map[string]harmony.Transaction{}
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
			txMap[hash] = harmony.Transaction{}
		}
	}
	// Get transactions by hash
	hashes := make([]string, len(txMap))
	i := 0
	for hash := range txMap {
		hashes[i] = hash
		i++
	}
	// Split into pages to avoid node stress
	for i := 0; i < len(hashes); i += 5000 {
		size := int(math.Min(float64(len(hashes)-i), 5000))
		txsPart, err := l.GetFullTransactions(hashes[i : i+size]...)
		if err != nil {
			return nil, err
		}
		txs = append(txs, txsPart...)
	}
	return
}

// GetFullTransactions returns transactions and their receipts for every given hash. Does not include method information (yet)
func (l *Loader) GetFullTransactions(hashes ...string) (txs []harmony.Transaction, err error) {
	// Prepare requests
	txs = make([]harmony.Transaction, len(hashes))
	bodiesByConn, idx := make([][]rpc.Body, l.uniqueConnCount), 0
	for _, hash := range hashes {
		b := l.uniqueConns[idx].NewBody(transactionByHashMethod, hash)
		bodiesByConn[idx] = append(bodiesByConn[idx], b)
		b = l.uniqueConns[idx].NewBody(transactionReceiptMethod, hash)
		bodiesByConn[idx] = append(bodiesByConn[idx], b)
		idx++
		if idx == l.uniqueConnCount {
			idx = 0
		}
	}
	// Do requests across unique nodes
	ch := make(chan goTx, len(hashes))
	for i, conn := range l.uniqueConns {
		go func(rpc *rpc.Rpc, bodies []rpc.Body) {
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
				//TODO Get method information with caching
				ch <- goTx{
					err: nil,
					tx:  tx,
				}
			}
		}(conn, bodiesByConn[i])
	}
	// Read output
	for i := 0; i < len(hashes); i++ {
		out := <-ch
		if out.err != nil {
			if err != nil {
				return txs, out.err
			}
		}
		txs[i] = out.tx
	}
	return
}

func readTxInfoFromResponse(data []byte) (tx harmony.Transaction, err error) {
	// Read JSON into harmony.Transaction
	var t transactionInfoJson
	err = json.Unmarshal(data, &t)
	if err != nil {
		return harmony.Transaction{}, errors.Wrap(err, 0)
	}
	value := new(big.Int)
	value.SetString(t.Value.String(), 10)
	gasPrice := new(big.Int)
	gasPrice.SetString(t.GasPrice.String(), 10)
	var method harmony.Method
	if len(t.Input) > 10 {
		method.Signature = t.Input[2:10]
	} else {
		method.Signature = ""
	}
	tx = harmony.Transaction{
		TxHash:    t.TxHash,
		Sender:    harmony.NewAddress(t.Sender),
		Receiver:  harmony.NewAddress(t.Receiver),
		BlockNum:  t.BlockNum,
		Timestamp: t.Timestamp,
		Value:     value,
		Method:    method,
		Input:     t.Input,
		GasAmount: t.GasAmount,
		GasPrice:  gasPrice,
		ShardID:   t.ShardID,
		ToShardID: t.ToShardID,
	}
	return
}

func readTxReceiptFromResponse(data []byte) (s int, ls []harmony.TransactionLog, err error) {
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
		ls = append(ls, harmony.TransactionLog{
			TxHash:   t.TxHash,
			LogIndex: int(index),
			Address:  harmony.NewAddress(l.Address),
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
