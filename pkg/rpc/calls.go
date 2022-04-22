package rpc

import (
	"encoding/json"
	"github.com/go-errors/errors"
	"strings"
	"time"
)

// Call executes an RPC and returns the result as a generic interface
func (r *RPC) Call(method string, params ...interface{}) (result interface{}, err error) {
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
func (r *RPC) BatchCall(bodies []Body) (results []interface{}, err error) {
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
func (r *RPC) RawCall(method string, params ...interface{}) (result []byte, err error) {
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
func (r *RPC) RawBatchCall(bodies []Body) (results [][]byte, err error) {
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
