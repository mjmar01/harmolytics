package cache

import (
	"github.com/mjmar01/harmolytics/pkg/hmybebop"
	"github.com/mjmar01/harmolytics/pkg/types"
	"github.com/syndtr/goleveldb/leveldb/util"
	"sync"
)

func (c *Cache) GetTransaction(hash string) (tx *types.Transaction, ok bool) {
	txMutex.RLock()
	tx, ok = txMemory[hash]
	txMutex.RUnlock()
	if ok {
		return
	}
	v, err := c.levelDB.Get(transactionKey(hash), nil)
	if err != nil {
		return nil, false
	}
	tx, err = hmybebop.DecodeTransaction(v)
	if err != nil {
		return nil, false
	}
	txMutex.Lock()
	txMemory[hash] = tx
	txMutex.Unlock()
	return tx, true
}

func (c *Cache) SetTransaction(tx *types.Transaction) {
	txMutex.Lock()
	txMemory[tx.EthTxHash] = tx
	txMutex.Unlock()
	v, err := hmybebop.EncodeTransaction(tx)
	if err != nil {
		return
	}
	c.levelDB.Put(transactionKey(tx.EthTxHash), v, nil)
}

func (c *Cache) loadTxMemory() {
	iter := c.levelDB.NewIterator(util.BytesPrefix(txPrefix), nil)
	wg := sync.WaitGroup{}
	for iter.Next() {
		src := iter.Value()
		cp := make([]byte, len(src))
		copy(cp, src)
		go func(in []byte) {
			wg.Add(1)
			txPtr, _ := hmybebop.DecodeTransaction(in)
			txMutex.Lock()
			txMemory[txPtr.EthTxHash] = txPtr
			txMutex.Unlock()
			wg.Done()
		}(cp)
	}
	wg.Wait()
	iter.Release()
}

func transactionKey(hash string) []byte {
	return append(txPrefix, []byte(hash)...)
}
