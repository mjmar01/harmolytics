package cache

import (
	"github.com/go-errors/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

func NewCache(opts *Opts) (newCache *Cache, err error) {
	newCache = new(Cache)
	opts = defaults(opts)

	newCache.levelDB, err = leveldb.OpenFile(opts.CacheDir, nil)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	if opts.PreLoadTransactions {
		newCache.loadTxMemory()
	}
	return
}

func (c *Cache) Close() {
	c.levelDB.Close()
}
