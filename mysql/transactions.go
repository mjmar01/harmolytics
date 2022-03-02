package mysql

import (
	"bytes"
	"database/sql"
	_ "embed"
	"fmt"
	"github.com/go-errors/errors"
	"harmolytics/harmony"
	"harmolytics/harmony/address"
	"math/big"
	"strings"
	"text/template"
)

const (
	transactionLogsHashQuery    = "SELECT hash, topics, address, data, log_index FROM harmolytics_profile_%s.transaction_logs WHERE hash = '%s'"
	transactionLogsTypeQuery    = "SELECT hash, topics, address, data, log_index FROM harmolytics_profile_%s.transaction_logs WHERE topics LIKE '%s%%'"
	transactionsMethodNameQuery = "SELECT hash, sender, receiver, input, method_signature, unixtime, block_num, gas_amount, gas_price, value, shard_id, to_shard_id FROM harmolytics_profile_%s.transactions WHERE method_signature IN (SELECT signature FROM harmolytics_default.methods WHERE name LIKE '%s') ORDER BY block_num ASC"
	transactionsLogIdQuery      = "SELECT hash, sender, receiver, input, method_signature, unixtime, block_num, gas_amount, gas_price, value, shard_id, to_shard_id FROM harmolytics_profile_%s.transactions WHERE hash IN (SELECT hash FROM harmolytics_profile_%s.transaction_logs WHERE topics LIKE '%s%%')"
	transactionHashQuery        = "SELECT hash, sender, receiver, input, method_signature, unixtime, block_num, gas_amount, gas_price, value, shard_id, to_shard_id FROM harmolytics_profile_%s.transactions WHERE hash = '%s'"
)

//go:embed queries/fill_transactions.tmpl
var transactionsQ string

//go:embed queries/fill_tokenTransfers.tmpl
var tokenTransfersQ string

// GetTransactionsByMethodName gets all harmony.Transaction with the specified method name, provided the name can be found in the methods table
func GetTransactionsByMethodName(name string) (txs []harmony.Transaction, err error) {
	rows, err := db.Query(fmt.Sprintf(transactionsMethodNameQuery, profile, name))
	defer rows.Close()
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	txs, err = getTransactionsFromResult(rows)
	if err != nil {
		return
	}
	return
}

// GetTransactionsByLogId gets all harmony.Transaction that contain a log of the specified type
func GetTransactionsByLogId(id string) (txs []harmony.Transaction, err error) {
	rows, err := db.Query(fmt.Sprintf(transactionsLogIdQuery, profile, profile, id))
	defer rows.Close()
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	txs, err = getTransactionsFromResult(rows)
	if err != nil {
		return
	}
	return
}

// GetTransactionByHash gets the harmony.Transaction with a given hash
func GetTransactionByHash(hash string) (tx harmony.Transaction, err error) {
	rows, err := db.Query(fmt.Sprintf(transactionHashQuery, profile, hash))
	defer rows.Close()
	if err != nil {
		return harmony.Transaction{}, errors.Wrap(err, 0)
	}
	txs, err := getTransactionsFromResult(rows)
	if err != nil {
		return
	}
	tx = txs[0]
	return
}

// GetTransactionLogsByHash returns a list of transaction.Log for a given transaction hash.
func GetTransactionLogsByHash(txHash string) (logs []harmony.TransactionLog, err error) {
	rows, err := db.Query(fmt.Sprintf(transactionLogsHashQuery, profile, txHash))
	defer rows.Close()
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	for rows.Next() {
		var l harmony.TransactionLog
		var topics, addr string
		err = rows.Scan(&l.TxHash, &topics, &addr, &l.Data, &l.LogIndex)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		l.Topics = strings.Split(topics, ":")
		l.Address, err = address.New(addr)
		if err != nil {
			return
		}
		logs = append(logs, l)
	}
	return
}

// GetTransactionLogByType gets all harmony.TransactionLog of a given type across all transactions
func GetTransactionLogByType(id string) (logs []harmony.TransactionLog) {
	rows, err := db.Query(fmt.Sprintf(transactionLogsTypeQuery, profile, id))
	defer rows.Close()
	if err != nil {
		return nil
	}
	for rows.Next() {
		var l harmony.TransactionLog
		var topics, addr string
		err := rows.Scan(&l.TxHash, &topics, &addr, &l.Data, &l.LogIndex)
		if err != nil {
			return nil
		}
		l.Topics = strings.Split(topics, ":")
		l.Address, err = address.New(addr)
		if err != nil {
			return nil
		}
		logs = append(logs, l)
	}
	return logs
}

// SetTransactions takes a list of transaction.Transaction and saves those to the tables transactions and transaction_logs
func SetTransactions(transactions []harmony.Transaction) (err error) {
	data := struct {
		Profile      string
		Transactions []harmony.Transaction
	}{
		Profile:      profile,
		Transactions: transactions,
	}
	var buf bytes.Buffer
	t, err := template.New("fillTransaction").Parse(transactionsQ)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = t.Execute(&buf, data)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = RunTemplate(buf.String())
	return
}

func SetTokenTransfers(transfers []harmony.TokenTransaction) (err error) {
	data := struct {
		Profile   string
		Transfers []harmony.TokenTransaction
	}{
		Profile:   profile,
		Transfers: transfers,
	}
	var buf bytes.Buffer
	t, err := template.New("fillTokenTransfers").Parse(tokenTransfersQ)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = t.Execute(&buf, data)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = RunTemplate(buf.String())
	return
}

func getTransactionsFromResult(rows *sql.Rows) (txs []harmony.Transaction, err error) {
	for rows.Next() {
		var tx harmony.Transaction
		var s, r, m, v, g string
		err = rows.Scan(&tx.TxHash, &s, &r, &tx.Input, &m, &tx.Timestamp, &tx.BlockNum, &tx.GasAmount, &g, &v, &tx.ShardID, &tx.ToShardID)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		tx.Sender, err = address.New(s)
		if err != nil {
			return
		}
		tx.Receiver, err = address.New(r)
		if err != nil {
			return
		}
		if m != "" {
			tx.Method, err = GetMethodBySignature(m)
		}
		if err != nil {
			return
		}
		tx.Logs, err = GetTransactionLogsByHash(tx.TxHash)
		if err != nil {
			return
		}
		tx.Value, tx.GasPrice = new(big.Int), new(big.Int)
		tx.Value.SetString(v, 10)
		tx.GasPrice.SetString(g, 10)
		txs = append(txs, tx)
	}
	return
}
