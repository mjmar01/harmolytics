// Package rpc exports slightly adapted versions of harmony RPC endpoints as functions for convenience.
package rpc

import (
	"encoding/json"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/gorilla/websocket"
)

var (
	conn         *websocket.Conn
	historicConn *websocket.Conn
	queryId      = 1
)

type rpcBody struct {
	RpcVersion string        `json:"jsonrpc"`
	Id         int           `json:"id"`
	Method     string        `json:"method"`
	Params     []interface{} `json:"params"`
}

// When the result explicitly is a string
type rpcReplyS struct {
	RpcVersion string `json:"jsonrpc"`
	Id         int    `json:"id"`
	Result     string `json:"result"`
}

// For generic results
type rpcReplyG struct {
	RpcVersion string      `json:"jsonrpc"`
	Id         int         `json:"id"`
	Result     interface{} `json:"result"`
}

func InitRpc(url, historicUrl string) (err error) {
	conn, _, err = websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	historicConn, _, err = websocket.DefaultDialer.Dial(historicUrl, nil)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	return
}
func CloseRpc() {
	conn.Close()
	historicConn.Close()
	if queryId > 1 {
		fmt.Printf("\n%d", queryId-1)
	}
}

func newRpcBody(method string) rpcBody {
	body := rpcBody{
		RpcVersion: "2.0",
		Id:         queryId,
		Method:     method,
	}
	queryId++
	return body
}

func rpcCall(method string, params interface{}) (result interface{}, err error) {
	body, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      queryId,
		"method":  method,
		"params":  params,
	})
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	queryId++
	err = conn.WriteMessage(websocket.TextMessage, body)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	_, ret, err := conn.ReadMessage()
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	var rst struct {
		Result interface{} `json:"result"`
	}
	err = json.Unmarshal(ret, &rst)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	result = rst.Result
	return
}

func rawRpcCall(method string, params interface{}) (result []byte, err error) {
	body, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      queryId,
		"method":  method,
		"params":  params,
	})
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	queryId++
	err = conn.WriteMessage(websocket.TextMessage, body)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	_, ret, err := conn.ReadMessage()
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	result = ret
	return
}

func historicRpcCall(method string, params interface{}) (result interface{}, err error) {
	body, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      queryId,
		"method":  method,
		"params":  params,
	})
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	queryId++
	err = historicConn.WriteMessage(websocket.TextMessage, body)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	_, ret, err := historicConn.ReadMessage()
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	var rst struct {
		Result interface{} `json:"result"`
	}
	err = json.Unmarshal(ret, &rst)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	result = rst.Result
	return
}
