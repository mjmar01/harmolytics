package cache

import (
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"path/filepath"
)

var txPrefix = []byte{0x01}

//<editor-fold desc="External types">

type Cache struct {
	levelDB *leveldb.DB
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
