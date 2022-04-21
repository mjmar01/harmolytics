package hmyload

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/pkg/hmybebop"
	"github.com/mjmar01/harmolytics/pkg/types"
	"github.com/syndtr/goleveldb/leveldb/util"
	"strings"
	"sync"
	"time"
)

const (
	txPrefix    = "tx:"
	tokenPrefix = "tk:"
)

var lock sync.RWMutex
var load sync.Once
var inMemory map[string]*types.Transaction

func (l *Loader) saveTransaction(tx *types.Transaction) (err error) {
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

func (l *Loader) checkTransaction(hash string) (tx *types.Transaction, ok bool) {
	load.Do(func() {
		t1 := time.Now()
		loadIntoMemory(l)
		fmt.Println(time.Since(t1))
	})
	lock.RLock()
	txPtr, inCache := inMemory[hash]
	lock.RUnlock()
	return txPtr, inCache
}

func loadIntoMemory(l *Loader) {
	lock.Lock()
	inMemory = map[string]*types.Transaction{}
	lock.Unlock()
	iter := l.cache.NewIterator(util.BytesPrefix([]byte(txPrefix)), nil)
	wg := sync.WaitGroup{}
	for iter.Next() {
		src := iter.Value()
		cp := make([]byte, len(src))
		copy(cp, src)
		go func(key string, in []byte) {
			wg.Add(1)
			txPtr, _ := hmybebop.DecodeTransaction(in)
			lock.Lock()
			inMemory[key] = txPtr
			lock.Unlock()
			wg.Done()
		}(strings.TrimPrefix(string(iter.Key()), txPrefix), cp)
	}
	wg.Wait()
	iter.Release()
}
