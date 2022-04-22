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

// RPC struct to interact with RPC endpoints
type RPC struct {
	PeerId  string
	timeout time.Duration
	ws      *websocket.Conn
	queryId int
}

// Opts contains optional parameters for the NewRpc function
type Opts struct {
	Timeout time.Duration
}

func defaults(in *Opts) (out *Opts) {
	if in == nil {
		out = new(Opts)
	} else {
		out = in
	}
	if out.Timeout == 0 {
		out.Timeout = time.Minute * 2
	}
	return
}

// Body represents an RPC calls body input
type Body struct {
	RpcVersion string        `json:"jsonrpc"`
	Id         int           `json:"id"`
	Method     string        `json:"method"`
	Params     []interface{} `json:"params"`
}

//</editor-fold>

//<editor-fold desc="Internal types">
type rpcReplyG struct {
	RpcVersion string      `json:"jsonrpc"`
	Id         int         `json:"id"`
	Result     interface{} `json:"result"`
}

// goFunc returns
type goRpcs struct {
	err error
	rpc *RPC
}

//</editor-fold>
