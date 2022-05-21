package test

import (
	"bytes"
	"encoding/gob"
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/pkg/cache"
	"github.com/mjmar01/harmolytics/pkg/hmybebop"
	"github.com/mjmar01/harmolytics/pkg/rpc"
	"github.com/mjmar01/harmolytics/pkg/types"
	"math/big"
)

const (
	url = "wss://ws.s0.t.hmny.io"
)

var tx = &types.Transaction{
	TxHash:    "0x98e9e65c1920a49f68cc523c8a5f5103d922fb8d250a859671b5a959aba9e2b1",
	EthTxHash: "0xaed17cb61f9c446112f6c50163c65d1c42d022183c6de9652495fee35a208a4f",
	Sender:    types.NewAddress("0x492064d08a3426fc15b7009301eb56bb285b6d08"),
	Receiver:  types.NewAddress("0xbda99c8695986b45a0dd3979cc6f3974d9753d30"),
	BlockNum:  24191303,
	Timestamp: 24191303,
	Value:     new(big.Int).SetInt64(0),
	Method:    types.Method{},
	Input:     "0xa69df4b5",
	Logs: []types.TransactionLog{
		{
			TxHash:   "0x98e9e65c1920a49f68cc523c8a5f5103d922fb8d250a859671b5a959aba9e2b1",
			LogIndex: 0,
			Address:  types.NewAddress("0xbda99c8695986b45a0dd3979cc6f3974d9753d30"),
			Topics: []string{
				"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
				"0x000000000000000000000000bda99c8695986b45a0dd3979cc6f3974d9753d30",
				"0x000000000000000000000000492064d08a3426fc15b7009301eb56bb285b6d08",
			},
			Data: "0x000000000000000000000000000000000000000000000038f7ab9a0605354e75",
		},
	},
	Status:    0,
	GasAmount: 61818,
	GasPrice:  new(big.Int).SetInt64(33000000000),
	ShardID:   0,
	ToShardID: 0,
}
var txBebop, txGob []byte

var centralCache *cache.Cache
var defaultRPC *rpc.RPC

var dump interface{}

func init() {
	// Pre encode for benchmarking
	buff := new(bytes.Buffer)
	enc := gob.NewEncoder(buff)
	enc.Encode(tx)
	txGob = buff.Bytes()
	txBebop, _ = hmybebop.EncodeTransaction(tx)

	// Open central cache
	var err error
	centralCache, err = cache.NewCache(&cache.Opts{PreLoadTransactions: false})
	if err != nil {
		panic(err.(*errors.Error).ErrorStack())
	}

	// Open benchmark RPC
	defaultRPC, err = rpc.NewRPC(url, nil)
	if err != nil {
		panic(err.(*errors.Error).ErrorStack())
	}
}
