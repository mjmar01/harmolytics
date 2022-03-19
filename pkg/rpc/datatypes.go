package rpc

import "encoding/json"

//<editor-fold desc="External types">
const (
	AllTx      = "ALL"
	SentTx     = "SENT"
	ReceivedTx = "RECEIVED"
)

//</editor-fold>

//<editor-fold desc="Internal types">
type transactionJson struct {
	TxHash    string      `json:"hash"`
	Sender    string      `json:"from"`
	Timestamp uint64      `json:"timestamp"`
	GasAmount uint64      `json:"gas"`
	GasPrice  json.Number `json:"gasPrice"`
	Input     string      `json:"input"`
	Receiver  string      `json:"to"`
	Value     json.Number `json:"value"`
	ShardID   uint        `json:"shardID"`
	ToShardID uint        `json:"toShardID"`
	BlockNum  uint64      `json:"blockNumber"`
}

type wrappedTransactionJson struct {
	Result transactionJson `json:"result"`
}

type transactionHistoryJson struct {
	Result struct {
		Transactions []transactionJson `json:"transactions"`
	} `json:"result"`
}

type transactionLogJson struct {
	Topics   []string `json:"topics"`
	Data     string   `json:"data"`
	Address  string   `json:"address"`
	LogIndex string   `json:"logIndex"`
}
type transactionReceiptJson struct {
	Result struct {
		Logs   []transactionLogJson `json:"logs"`
		Status int                  `json:"status"`
		From   string               `json:"from"`
		TxHash string               `json:"transactionHash"`
	} `json:"result"`
}

//</editor-fold>
