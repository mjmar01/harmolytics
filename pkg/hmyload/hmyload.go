package hmyload

import (
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/pkg/cache"
	"github.com/mjmar01/harmolytics/pkg/rpc"
)

// NewLoader creates a struct to load blockchain data
func NewLoader(url string, opts *Opts) (l *Loader, err error) {
	opts = defaults(opts)
	l = new(Loader)

	// Create RPCs
	rs, err := rpc.NewRPCs(url, opts.AdditionalConnections, &rpc.Opts{Timeout: opts.RpcTimeout})
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	l.optionalConns = rs
	l.defaultConn = rs[0]

	// Open cache
	if opts.ExistingCache != nil {
		l.sharedCache = true
		l.cache = opts.ExistingCache
	} else {
		l.sharedCache = false
		l.cache, err = cache.NewCache(&cache.Opts{
			CacheDir:            opts.CacheDir,
			PreLoadTransactions: opts.PreLoadCacheTransactions,
		})
	}

	// Fill Loader metadata
	l.connByPeer = make(map[string][]*rpc.RPC)
	for _, conn := range l.optionalConns {
		l.connByPeer[conn.PeerId] = append(l.connByPeer[conn.PeerId], conn)
	}
	l.uniqueConns = make([]*rpc.RPC, len(l.connByPeer))
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
	if !l.sharedCache {
		l.cache.Close()
	}
}
