package hmyload

import (
	"encoding/json"
	"github.com/mjmar01/harmolytics/pkg/harmony"
	"github.com/mjmar01/harmolytics/pkg/rpc"
	"time"
)

//<editor-fold desc="External types">

// Loader struct used to load blockchain data
type Loader struct {
	defaultConn     rpc.Rpc
	optionalConns   []rpc.Rpc
	connCount       int
	uniqueConnCount int
	connByPeer      map[string][]rpc.Rpc
	uniqueConns     []rpc.Rpc
}

// Opts contains optional parameters for the NewLoader function
type Opts struct {
	AdditionalConnections int
	RpcTimeout            time.Duration
}

func (o *Opts) defaults() *Opts {
	if o == nil {
		o = new(Opts)
	}
	if o.AdditionalConnections == 0 {
		o.AdditionalConnections = 1
	}
	if o.RpcTimeout == 0 {
		o.RpcTimeout = time.Minute * 2
	}
	return o
}

//</editor-fold>

//<editor-fold desc="Internal types">
type transactionInfoJson struct {
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

type transactionReceiptJson struct {
	Logs   []transactionLogJson `json:"logs"`
	Status int                  `json:"status"`
	TxHash string               `json:"transactionHash"`
}

type transactionLogJson struct {
	Topics   []string `json:"topics"`
	Data     string   `json:"data"`
	Address  string   `json:"address"`
	LogIndex string   `json:"logIndex"`
}

// goFunc returns
type goTx struct {
	err error
	tx  harmony.Transaction
}

type goTk struct {
	err error
	tk  harmony.Token
}

//</editor-fold>
