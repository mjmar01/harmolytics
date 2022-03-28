package rpc

import (
	"github.com/gorilla/websocket"
)

//<editor-fold desc="External types">
const (
	AllTx      = "ALL"
	SentTx     = "SENT"
	ReceivedTx = "RECEIVED"
)

type Rpc struct {
	ws      *websocket.Conn
	queryId int
}

type Body struct {
	RpcVersion string        `json:"jsonrpc"`
	Id         int           `json:"id"`
	Method     string        `json:"method"`
	Params     []interface{} `json:"params"`
}

//</editor-fold>

//<editor-fold desc="Internal types">
// For generic results
type rpcReplyG struct {
	RpcVersion string      `json:"jsonrpc"`
	Id         int         `json:"id"`
	Result     interface{} `json:"result"`
}

//</editor-fold>
