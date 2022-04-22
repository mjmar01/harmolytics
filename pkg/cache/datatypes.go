package cache

import (
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"path/filepath"
	"sync"
)

//<editor-fold desc="External types">

type Cache struct {
	levelDB    *leveldb.DB
	closeLock  int
	closeMutex sync.Mutex
}

type Opts struct {
	CacheDir            string
	PreLoadTransactions bool
}

func defaults(in *Opts) (out *Opts) {
	if in == nil {
		out = new(Opts)
	} else {
		out = in
	}
	if out.CacheDir == "" {
		dir, _ := os.UserCacheDir()
		dir = filepath.Join(dir, "harmony-tk")
		out.CacheDir = dir
	}
	return
}

//</editor-fold>
