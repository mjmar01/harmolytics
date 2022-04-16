// Package rpc exports slightly adapted versions of harmony RPC endpoints as functions for convenience.
package rpc

import (
	"encoding/json"
	"github.com/go-errors/errors"
	"github.com/gorilla/websocket"
	"strings"
	"sync"
	"time"
)

const (
	NodeMetadataMethod = "hmyv2_getNodeMetadata"
)

func NewRpc(url string, opts *Opts) (r *Rpc, err error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	r = &Rpc{
		ws:      conn,
		queryId: 1,
	}
	opts = opts.defaults()
	r.timeout = opts.Timeout
	metaData, err := r.Call(NodeMetadataMethod)
	r.peerId = metaData.(map[string]interface{})["peerid"].(string)
	return
}

func NewRpcs(url string, count int, opts *Opts) (rs []*Rpc, err error) {
	rs, wg, ch := make([]*Rpc, count), new(sync.WaitGroup), make(chan goRpcs, count)
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			r, err := NewRpc(url, opts)
			ch <- goRpcs{
				err: err,
				rpc: r,
			}
			wg.Done()
		}()
	}
	wg.Wait()
	for i := 0; i < count; i++ {
		out := <-ch
		if out.err != nil {
			return nil, errors.Wrap(out.err, 0)
		}
		rs[i] = out.rpc
	}
	return
}

func (r *Rpc) Close() {
	r.ws.Close()
}

func (r *Rpc) NewBody(method string, params ...interface{}) (b Body) {
	if params == nil {
		params = []interface{}{}
	}
	b = Body{
		RpcVersion: "2.0",
		Id:         r.queryId,
		Method:     method,
		Params:     params,
	}
	r.queryId++
	return
}

func (r *Rpc) Call(method string, params ...interface{}) (result interface{}, err error) {
	body := r.NewBody(method, params...)
	err = r.ws.WriteJSON(body)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	r.ws.SetReadDeadline(time.Now().Add(r.timeout))
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
		r.ws.SetReadDeadline(time.Now().Add(r.timeout))
		var rst rpcReplyG
		err = r.ws.ReadJSON(&rst)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		results[idx[rst.Id]] = rst.Result
	}
	return
}

func (r *Rpc) RawCall(method string, params ...interface{}) (result []byte, err error) {
	body := r.NewBody(method, params...)
	err = r.ws.WriteJSON(body)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	r.ws.SetReadDeadline(time.Now().Add(r.timeout))
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
		r.ws.SetReadDeadline(time.Now().Add(r.timeout))
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
