// Package transaction handles reading transaction data from a harmony RPC API
package transaction

import (
	"fmt"
	"harmolytics/harmony"
	"harmolytics/harmony/rpc"
	"sync"
)

// GetTransactionsByWallet returns a list of all successful Transaction for a given address.Address
func GetTransactionsByWallet(addr harmony.Address) (txs []harmony.Transaction, err error) {
	// Get total number of transactions
	txCount, err := rpc.GetTransactionCount(addr, rpc.AllTx)
	if err != nil {
		return
	}
	// Split into groups of 1000
	for i := 0; i < txCount; i += 1000 {
		// Get upto 1000 transactions
		transactions, err := rpc.GetTransactionHistory(addr.OneAddress, i/1000, 1000, rpc.AllTx)
		if err != nil {
			return nil, err
		}
		// Fill remaining transaction info using go routines
		wg := sync.WaitGroup{}
		wg.Add(len(transactions))
		ch := make(chan harmony.Transaction, len(transactions))
		for _, transaction := range transactions {
			go func(tx harmony.Transaction) {
				// Retrieve transaction receipt
				txStatus, logs, err := rpc.GetTransactionReceipt(tx.TxHash)
				if err != nil {
					// TODO thread safe logging in go routines
					fmt.Printf("Error occured while trying to get transaction receipt. %s. %s\n", err.Error(), tx.TxHash)
					ch <- harmony.Transaction{}
					wg.Done()
					return
				}
				if txStatus == harmony.TxFailed {
					tx.TxHash = ""
					ch <- tx
					wg.Done()
					return
				}

				tx.Logs = logs
				// TODO Retrieve method information
				// Proper way would be to get method info here but this is also wasteful.
				// Either implement thread safe caching or do it later. More than signature info isn't needed for now.
				// Currently this is being dealt with using the load methods command.
				ch <- tx
				wg.Done()
			}(transaction)
		}
		wg.Wait()
		// Filter out failed transactions
		for i := 0; i < len(transactions); i++ {
			tx := <-ch
			if tx.TxHash != "" {
				txs = append(txs, tx)
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
