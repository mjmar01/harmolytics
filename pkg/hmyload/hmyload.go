package hmyload

import (
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/pkg/rpc"
	"github.com/syndtr/goleveldb/leveldb"
	"path/filepath"
)

// NewLoader creates a struct to load blockchain data
func NewLoader(url string, opts *Opts) (l *Loader, err error) {
	opts, err = opts.defaults()
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	// Create default RPC
	l = new(Loader)
	l.defaultConn, err = rpc.NewRpc(url, &rpc.Opts{Timeout: opts.RpcTimeout})
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	// Create additional RPCs or assign default as only addition
	if opts.AdditionalConnections != 1 {
		l.optionalConns, err = rpc.NewRpcs(url, opts.AdditionalConnections, nil)
		if err != nil {
			return
		}
		l.connCount = opts.AdditionalConnections
	} else {
		l.optionalConns = []*rpc.Rpc{l.defaultConn}
		l.connCount = 1
	}
	// Open cache
	err = l.openCache(opts.CacheDir)
	if err != nil {
		return
	}
	// Fill Loader metadata
	l.connByPeer = make(map[string][]*rpc.Rpc)
	for _, conn := range l.optionalConns {
		l.connByPeer[conn.PeerId] = append(l.connByPeer[conn.PeerId], conn)
	}
	l.uniqueConns = make([]*rpc.Rpc, len(l.connByPeer))
	idx := 0
	for _, rpcs := range l.connByPeer {
		l.uniqueConns[idx] = rpcs[0]
		idx++
	}
	l.uniqueConnCount = len(l.uniqueConns)
	return
}

func (l *Loader) openCache(path string) (err error) {
	openCachesInit.Do(func() {
		openCaches = map[string]*leveldb.DB{}
		openCacheTracker = map[string]int{}
	})
	openCacheLock.Lock()
	path, err = filepath.Abs(path)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	cache, exists := openCaches[path]
	if exists {
		openCacheTracker[path]++
		l.cache = cache
		l.cachePath = path
		openCacheLock.Unlock()
		return
	}
	newCache, err := leveldb.OpenFile(path, nil)
	if err != nil {
		openCacheLock.Unlock()
		return errors.Wrap(err, 0)
	}
	openCaches[path] = newCache
	openCacheTracker[path] = 1

	l.cache = newCache
	l.cachePath = path
	openCacheLock.Unlock()
	return
}

func (l *Loader) Close() {
	l.defaultConn.Close()
	for _, conn := range l.optionalConns {
		conn.Close()
	}
	openCacheLock.Lock()
	used, _ := openCacheTracker[l.cachePath]
	if used == 1 {
		delete(openCaches, l.cachePath)
		delete(openCacheTracker, l.cachePath)
		l.cache.Close()
	} else {
		openCacheTracker[l.cachePath]--
	}
	openCacheLock.Unlock()
}
