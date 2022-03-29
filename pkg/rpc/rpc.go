// Package rpc exports slightly adapted versions of harmony RPC endpoints as functions for convenience.
package rpc

import (
	"encoding/json"
	"github.com/go-errors/errors"
	"github.com/gorilla/websocket"
	"strings"
	"sync"
)

func NewRpc(url string) (r *Rpc, err error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	r = &Rpc{
		ws:      conn,
		queryId: 1,
	}
	return
}

func NewRpcs(url string, count int) (rs []*Rpc, err error) {
	rs = make([]*Rpc, count)
	// Do one normal to check for error
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	rs[0] = &Rpc{
		ws:      conn,
		queryId: 1,
	}
	// Do the others parallel
	wg := new(sync.WaitGroup)
	wg.Add(count - 1)
	ch := make(chan *Rpc, count)
	for i := 1; i < count; i++ {
		go func() {
			conn, _, _ := websocket.DefaultDialer.Dial(url, nil)
			ch <- &Rpc{
				ws:      conn,
				queryId: 1,
			}
			wg.Done()
		}()
	}
	wg.Wait()
	for i := 1; i < count; i++ {
		rs[i] = <-ch
	}
	return
}

func (r *Rpc) Close() {
	r.ws.Close()
}

func (r *Rpc) NewBody(method string, params []interface{}) (b Body) {
	b = Body{
		RpcVersion: "2.0",
		Id:         r.queryId,
		Method:     method,
		Params:     params,
	}
	r.queryId++
	return
}

func (r *Rpc) Call(method string, params []interface{}) (result interface{}, err error) {
	body := r.NewBody(method, params)
	err = r.ws.WriteJSON(body)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	var rst rpcReplyG
	err = r.ws.ReadJSON(&rst)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	result = rst.Result
	return
}

func (r *Rpc) BatchCall(bodies []Body) (results []interface{}, err error) {
	results = make([]interface{}, len(bodies))
	idx := make(map[int]int, len(bodies))
	for i, body := range bodies {
		idx[body.Id] = i
		err = r.ws.WriteJSON(body)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
	}
	for i := 0; i < len(bodies); i++ {
		var rst rpcReplyG
		err = r.ws.ReadJSON(&rst)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		results[idx[rst.Id]] = rst.Result
	}
	return
}

func (r *Rpc) RawCall(method string, params []interface{}) (result []byte, err error) {
	body := r.NewBody(method, params)
	err = r.ws.WriteJSON(body)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	_, rs, err := r.ws.ReadMessage()
	if err != nil {
		return result, errors.Wrap(err, 0)
	}
	result = []byte(strings.TrimSuffix(strings.SplitAfterN(string(rs), ":", 4)[3], "}\n"))
	return
}

func (r *Rpc) RawBatchCall(bodies []Body) (results [][]byte, err error) {
	results = make([][]byte, len(bodies))
	idx := make(map[int]int, len(bodies))
	for i, body := range bodies {
		idx[body.Id] = i
		err = r.ws.WriteJSON(body)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
	}
	for i := 0; i < len(bodies); i++ {
		var rst rpcReplyG
		_, rs, err := r.ws.ReadMessage()
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		err = json.Unmarshal(rs, &rst)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		results[idx[rst.Id]] = []byte(strings.TrimSuffix(strings.SplitAfterN(string(rs), ":", 4)[3], "}\n"))
	}
	return
}
