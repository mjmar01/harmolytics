package hmyload

import (
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/pkg/rpc"
)

// NewLoader creates a struct to load blockchain data
func NewLoader(url string, opts *Opts) (l Loader, err error) {
	opts = opts.defaults()
	// Create default RPC
	l.defaultConn, err = rpc.NewRpc(url, &rpc.Opts{Timeout: opts.RpcTimeout})
	if err != nil {
		return Loader{}, errors.Wrap(err, 0)
	}
	// Create additional RPCs or assign default as only addition
	if opts.AdditionalConnections != 1 {
		l.optionalConns, err = rpc.NewRpcs(url, opts.AdditionalConnections, nil)
		if err != nil {
			return
		}
		l.connCount = opts.AdditionalConnections
	} else {
		l.optionalConns = []rpc.Rpc{l.defaultConn}
		l.connCount = 1
	}
	// Fill Loader metadata
	l.connByPeer = make(map[string][]rpc.Rpc)
	for _, conn := range l.optionalConns {
		l.connByPeer[conn.PeerId] = append(l.connByPeer[conn.PeerId], conn)
	}
	l.uniqueConns = make([]rpc.Rpc, len(l.connByPeer))
	idx := 0
	for _, rpcs := range l.connByPeer {
		l.uniqueConns[idx] = rpcs[0]
		idx++
	}
	l.uniqueConnCount = len(l.uniqueConns)
	return
}

func (l *Loader) Close() {
	l.defaultConn.Close()
	for _, conn := range l.optionalConns {
		conn.Close()
	}
}
