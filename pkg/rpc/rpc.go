// Package rpc handles communication with harmony nodes over the RPC protocol using websockets
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

// NewRpc creates a new Rpc struct
func NewRpc(url string, opts *Opts) (r Rpc, err error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return Rpc{}, errors.Wrap(err, 0)
	}
	r = Rpc{
		ws:      conn,
		queryId: 1,
	}
	opts = opts.defaults()
	r.timeout = opts.Timeout
	metaData, err := r.Call(NodeMetadataMethod)
	r.PeerId = metaData.(map[string]interface{})["peerid"].(string)
	return
}

// NewRpcs is used to generated multiple Rpc structs using go routines
func NewRpcs(url string, count int, opts *Opts) (rs []Rpc, err error) {
	rs, wg, ch := make([]Rpc, count), new(sync.WaitGroup), make(chan goRpcs, count)
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

// NewBody prepares a body with correct syntax and incremental IDs
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

// Call executes an RPC and returns the result as a generic interface
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

// BatchCall executes one RPC per given Body and returns the result as a slice of generic interfaces
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

// RawCall executes an RPC and returns the raw JSON result
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

// RawBatchCall executes one RPC per given Body and returns the result as a slice of raw JSON results
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
