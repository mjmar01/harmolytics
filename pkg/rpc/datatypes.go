package rpc

import (
	"github.com/gorilla/websocket"
	"time"
)

//<editor-fold desc="External types">
const (
	AllTx      = "ALL"
	SentTx     = "SENT"
	ReceivedTx = "RECEIVED"
)

type Rpc struct {
	PeerId  string
	timeout time.Duration
	ws      *websocket.Conn
	queryId int
}

type Opts struct {
	Timeout time.Duration
}

func (o *Opts) defaults() *Opts {
	if o == nil {
		o = new(Opts)
	}
	if o.Timeout == 0 {
		o.Timeout = time.Minute * 2
	}
	return o
}

type Body struct {
	RpcVersion string        `json:"jsonrpc"`
	Id         int           `json:"id"`
	Method     string        `json:"method"`
	Params     []interface{} `json:"params"`
}

//</editor-fold>

//<editor-fold desc="Internal types">
// Generic RPC results
type rpcReplyG struct {
	RpcVersion string      `json:"jsonrpc"`
	Id         int         `json:"id"`
	Result     interface{} `json:"result"`
}

// goFunc returns
type goRpcs struct {
	err error
	rpc *Rpc
}

//</editor-fold>
