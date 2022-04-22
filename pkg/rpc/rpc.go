// Package rpc handles communication with harmony nodes over the RPC protocol using websockets
package rpc

import (
	"github.com/go-errors/errors"
	"github.com/gorilla/websocket"
	"net/http"
)

const (
	NodeMetadataMethod = "hmyv2_getNodeMetadata"
)

// NewRpc creates a new RPC struct
func NewRpc(url string, opts *Opts) (r *RPC, err error) {
	opts = defaults(opts)
	r = new(RPC)
	var rsp *http.Response
	r.ws, rsp, err = websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		if rsp.StatusCode == 502 {
			return nil, errors.Errorf("failed to open websocket due to temporary connection issues. This shouldn't last longer than ~5min")
		}
		if rsp.Status != "" {
			return nil, errors.Errorf("failed to open websocket. Received status: %s", rsp.Status)
		}
		return nil, errors.Wrap(err, 0)
	}

	r.queryId = 1
	r.timeout = opts.Timeout

	metaData, err := r.Call(NodeMetadataMethod)
	r.PeerId = metaData.(map[string]interface{})["peerid"].(string)
	return
}

// NewRpcs is used to generated multiple RPC structs using go routines
func NewRpcs(url string, count int, opts *Opts) (rs []*RPC, err error) {
	rs, ch := make([]*RPC, count), make(chan goRpcs, count)
	for i := 0; i < count; i++ {
		go func() {
			r, err := NewRpc(url, opts)
			ch <- goRpcs{
				err: err,
				rpc: r,
			}
		}()
	}
	for i := 0; i < count; i++ {
		out := <-ch
		if out.err != nil {
			return nil, errors.Wrap(out.err, 0)
		}
		rs[i] = out.rpc
	}
	return
}

func (r *RPC) Close() {
	r.ws.Close()
}
