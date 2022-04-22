package hmyload

import (
	"encoding/json"
	"github.com/mjmar01/harmolytics/pkg/cache"
	"github.com/mjmar01/harmolytics/pkg/rpc"
	"github.com/mjmar01/harmolytics/pkg/types"
	"time"
)

//<editor-fold desc="External types">

// Loader struct used to load blockchain data
type Loader struct {
	defaultConn     *rpc.RPC
	optionalConns   []*rpc.RPC
	connCount       int
	uniqueConnCount int
	connByPeer      map[string][]*rpc.RPC
	uniqueConns     []*rpc.RPC
	cache           *cache.Cache
	sharedCache     bool
}

// Opts contains optional parameters for the NewLoader function
type Opts struct {
	// Loader settings
	AdditionalConnections int
	// RPC settings
	RpcTimeout time.Duration
	// Cache settings
	CacheDir                 string
	ExistingCache            *cache.Cache
	PreLoadCacheTransactions bool
}

func defaults(in *Opts) (out *Opts) {
	if in == nil {
		out = new(Opts)
	} else {
		out = in
	}
	if out.AdditionalConnections == 0 {
		out.AdditionalConnections = 1
	}
	return
}

//</editor-fold>

//<editor-fold desc="Internal types">
type transactionInfoJson struct {
	TxHash    string      `json:"hash"`
	EthTxHash string      `json:"ethHash"`
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
	tx  *types.Transaction
}

type goTk struct {
	err error
	tk  types.Token
}

//</editor-fold>
