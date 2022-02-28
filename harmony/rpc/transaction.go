package rpc

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-errors/errors"
	"github.com/gorilla/websocket"
	"harmolytics/harmony"
	"harmolytics/harmony/address"
	"math/big"
)

const (
	transactionReceipt = "hmyv2_getTransactionReceipt"
	transactionByHash  = "hmyv2_getTransactionByHash"
)

// GetTransaction uses the getTransactionByHash endpoint to retrieve information for a single transaction.
// This does not fill the harmony.Transaction with all values
func GetTransaction(hash string) (tx harmony.Transaction, err error) {
	// Get transaction info
	params := []interface{}{hash}
	result, err := rawRpcCall(transactionByHash, params)
	if err != nil {
		return
	}
	// read result
	var wTxJson wrappedTransactionJson
	err = json.Unmarshal(result, &wTxJson)
	if err != nil {
		return harmony.Transaction{}, errors.Wrap(err, 0)
	}
	txJson := wTxJson.Result
	// Convert transactionJson to harmony.Transaction
	sender, err := address.New(txJson.Sender)
	if err != nil {
		return
	}
	receiver, err := address.New(txJson.Receiver)
	if err != nil {
		return
	}
	value := new(big.Int)
	value.SetString(txJson.Value.String(), 10)
	gasPrice := new(big.Int)
	gasPrice.SetString(txJson.GasPrice.String(), 10)
	var method harmony.Method
	if len(txJson.Input) > 10 {
		method.Signature = txJson.Input[2:10]
	} else {
		method.Signature = ""
	}
	// Fill transaction
	tx = harmony.Transaction{
		TxHash:    txJson.TxHash,
		Sender:    sender,
		Receiver:  receiver,
		BlockNum:  txJson.BlockNum,
		Timestamp: txJson.Timestamp,
		Value:     value,
		Method:    method,
		Input:     txJson.Input[2:],
		Logs:      nil,
		GasAmount: txJson.GasAmount,
		GasPrice:  gasPrice,
		ShardID:   txJson.ShardID,
		ToShardID: txJson.ToShardID,
	}
	return
}

// GetTransactionReceipt uses the getTransactionReceipt endpoint to get all logs for a given transaction hash and its status.
// (Status: 0 -> not OK, 1 -> OK)
func GetTransactionReceipt(hash string) (status int, logs []harmony.TransactionLog, err error) {
	// Get transaction receipt
	params := []interface{}{hash}
	result, err := rawRpcCall(transactionReceipt, params)
	if err != nil {
		return
	}
	// Read result
	var txReceipt transactionReceiptJson
	err = json.Unmarshal(result, &txReceipt)
	if err != nil {
		return 0, nil, errors.Wrap(err, 0)
	}
	status = txReceipt.Result.Status
	// Convert transactionReceiptJson to []harmony.TransactionLog
	for _, l := range txReceipt.Result.Logs {
		// Convert data formats
		index, err := hexutil.DecodeUint64(l.LogIndex)
		if err != nil {
			return 0, nil, errors.Wrap(err, 0)
		}
		addr, err := address.New(l.Address)
		if err != nil {
			return 0, nil, err
		}
		// Add log
		logs = append(logs, harmony.TransactionLog{
			TxHash:   hash,
			LogIndex: int(index),
			Address:  addr,
			Topics:   l.Topics,
			Data:     l.Data,
		})
	}
	return
}

// GetTransactionReceipts is functionally identical to GetTransactionReceipt except, it is optimised for a large number
// of requests
func GetTransactionReceipts(hashs []string) (txs []harmony.Transaction, err error) {
	for _, hash := range hashs {
		body, err := json.Marshal(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      queryId,
			"method":  transactionReceipt,
			"params":  []interface{}{hash},
		})
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		queryId++
		err = conn.WriteMessage(websocket.TextMessage, body)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
	}
	// Get transaction receipt
	for i := 0; i < len(hashs); i++ {
		var logs []harmony.TransactionLog
		var tx harmony.Transaction
		_, result, err := conn.ReadMessage()
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		// Read result
		var txReceipt transactionReceiptJson
		err = json.Unmarshal(result, &txReceipt)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		if txReceipt.Result.Status == harmony.TxFailed {
			continue
		}
		// Convert transactionReceiptJson to []harmony.TransactionLog
		for _, l := range txReceipt.Result.Logs {
			// Convert data formats
			index, err := hexutil.DecodeUint64(l.LogIndex)
			if err != nil {
				return nil, errors.Wrap(err, 0)
			}
			addr, err := address.New(l.Address)
			if err != nil {
				return nil, err
			}
			// Add log
			logs = append(logs, harmony.TransactionLog{
				TxHash:   txReceipt.Result.TxHash,
				LogIndex: int(index),
				Address:  addr,
				Topics:   l.Topics,
				Data:     l.Data,
			})
		}
		tx.Logs = logs
		tx.TxHash = txReceipt.Result.TxHash

		txs = append(txs, tx)
	}
	return
}
