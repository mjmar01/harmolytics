package hmyload

import (
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/pkg/harmony"
	"github.com/mjmar01/harmolytics/pkg/hmybebop"
	"github.com/syndtr/goleveldb/leveldb/util"
	"strings"
	"sync"
)

const (
	txPrefix    = "tx:"
	tokenPrefix = "tk:"
)

var lock sync.RWMutex
var load sync.Once
var inMemory map[string][]byte

func (l *Loader) saveTransaction(tx harmony.Transaction) (err error) {
	data, err := hmybebop.EncodeTransaction(tx)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	key := append([]byte(txPrefix), []byte(tx.EthTxHash)...)
	err = l.cache.Put(key, data, nil)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	return
}

func (l *Loader) checkTransaction(hash string) (tx harmony.Transaction, ok bool) {
	load.Do(func() {
		lock.Lock()
		inMemory = map[string][]byte{}
		iter := l.cache.NewIterator(util.BytesPrefix([]byte(txPrefix)), nil)
		for iter.Next() {
			src := iter.Value()
			cp := make([]byte, len(src))
			copy(cp, src)
			inMemory[strings.TrimPrefix(string(iter.Key()), txPrefix)] = cp
		}
		iter.Release()
		lock.Unlock()
	})
	lock.RLock()
	rawTx, inCache := inMemory[hash]
	lock.RUnlock()
	if !inCache {
		return harmony.Transaction{}, false
	}
	tx, err := hmybebop.DecodeTransaction(rawTx)
	if err != nil {
		return harmony.Transaction{}, false
	}
	ok = true
	return
}
