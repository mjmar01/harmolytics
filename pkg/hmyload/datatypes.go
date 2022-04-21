package hmyload

import (
	"encoding/json"
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/pkg/rpc"
	"github.com/mjmar01/harmolytics/pkg/types"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"sync"
	"time"
)

//<editor-fold desc="External types">

// Loader struct used to load blockchain data
type Loader struct {
	defaultConn     *rpc.Rpc
	optionalConns   []*rpc.Rpc
	connCount       int
	uniqueConnCount int
	connByPeer      map[string][]*rpc.Rpc
	uniqueConns     []*rpc.Rpc
	cache           *leveldb.DB
	cachePath       string
}

var openCachesInit sync.Once
var openCaches map[string]*leveldb.DB
var openCacheTracker map[string]int
var openCacheLock sync.Mutex

// Opts contains optional parameters for the NewLoader function
type Opts struct {
	AdditionalConnections int
	RpcTimeout            time.Duration
	CacheDir              string
}

func (o *Opts) defaults() (out *Opts, err error) {
	if o != nil {
		out = o
	} else {
		out = new(Opts)
	}
	if out.AdditionalConnections == 0 {
		out.AdditionalConnections = 1
	}
	if out.RpcTimeout == 0 {
		out.RpcTimeout = time.Minute * 2
	}
	if out.CacheDir == "" {
		var dir string
		dir, err = os.UserCacheDir()
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		out.CacheDir = dir + "/harmony-tk"
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
