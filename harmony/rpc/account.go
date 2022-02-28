package rpc

import (
	"encoding/json"
	"github.com/go-errors/errors"
	"harmolytics/harmony"
	addressPkg "harmolytics/harmony/address"
	"math/big"
)

const (
	transactionCount   = "hmyv2_getTransactionsCount"
	transactionHistory = "hmyv2_getTransactionsHistory"
)

// GetTransactionCount returns the number of transactions for a given address and type.
// (Types: rpc.AllTx, rpc.SentTx, rpc.ReceivedTx)
func GetTransactionCount(address harmony.Address, txType string) (c int, err error) {
	// Get transaction count
	params := []interface{}{address.OneAddress, txType}
	result, err := rpcCall(transactionCount, params)
	if err != nil {
		return
	}
	// Read result (which is float64 for some unknown reason)
	c = int(result.(float64))
	return
}

// GetTransactionHistory calls the hmyv2_getTransactionsHistory method and parses the result into a list of harmony.Transaction.
// This does not fill the harmony.Transaction with all values
func GetTransactionHistory(address string, pageIndex, pageSize int, txType string) (txs []harmony.Transaction, err error) {
	// Get raw transaction history
	params := struct {
		Address   string `json:"address"`
		PageIndex int    `json:"pageIndex"`
		PageSize  int    `json:"pageSize"`
		FullTx    bool   `json:"fullTx"`
		TxType    string `json:"txType"`
	}{
		Address:   address,
		PageIndex: pageIndex,
		PageSize:  pageSize,
		FullTx:    true,
		TxType:    txType,
	}
	result, err := rawRpcCall(transactionHistory, []interface{}{params})
	if err != nil {
		return
	}

	// Convert transactionHistoryJson to []harmony.Transaction
	var transactions transactionHistoryJson
	err = json.Unmarshal(result, &transactions)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	for _, txJson := range transactions.Result.Transactions {
		// Convert data formats
		sender, err := addressPkg.New(txJson.Sender)
		if err != nil {
			return nil, err
		}
		receiver, err := addressPkg.New(txJson.Receiver)
		if err != nil {
			return nil, err
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
		// Add harmony.Transaction
		txs = append(txs, harmony.Transaction{
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
		})
	}
	return
}
