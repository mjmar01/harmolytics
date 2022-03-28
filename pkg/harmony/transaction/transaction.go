// Package transaction handles reading transaction data from a harmony RPC API
package transaction

import (
	"github.com/mjmar01/harmolytics/pkg/harmony"
	"github.com/mjmar01/harmolytics/pkg/rpc"
)

// GetTransactionsByWallet returns a list of all successful Transaction for a given address.Address
func GetTransactionsByWallet(addr harmony.Address) (txs []harmony.Transaction, err error) {
	// Get total number of transactions
	txCount, err := rpc.GetTransactionCount(addr, rpc.AllTx)
	if err != nil {
		return
	}
	// Split into groups
	pageSize := 10000
	for i := 0; i < txCount; i += pageSize {
		// Get upto 1000 transactions
		transactions, err := rpc.GetTransactionHistory(addr.OneAddress, i/pageSize, pageSize, rpc.AllTx)
		if err != nil {
			return nil, err
		}
		var hashs []string
		for _, transaction := range transactions {
			hashs = append(hashs, transaction.TxHash)
		}
		receipts, err := rpc.GetTransactionReceipts(hashs)
		if err != nil {
			return nil, err
		}
		// Fill remaining transaction info
		for _, tx := range transactions {
			for i2, receipt := range receipts {
				if tx.TxHash == receipt.TxHash {
					tx.Logs = receipt.Logs
					txs = append(txs, tx)
					receipts = append(receipts[:i2], receipts[i2+1:]...)
					break
				}
			}
		}
	}
	return
}

func GetFullTransaction(hash string) (tx harmony.Transaction, err error) {
	// Get basic transaction information
	tx, err = rpc.GetTransaction(hash)
	if err != nil {
		return
	}
	// Load transaction receipt
	txStatus, logs, err := rpc.GetTransactionReceipt(tx.TxHash)
	if err != nil {
		return
	}
	if txStatus == harmony.TxFailed {
		tx = harmony.Transaction{}
		return
	}
	tx.Logs = logs
	// TODO Retrieve method information
	// You could retrieve method information here, but depending on how this function is used it could be wasteful.
	// Again either implement caching or ignore for now.
	// Currently this is being dealt with using the load methods command.
	return
}
