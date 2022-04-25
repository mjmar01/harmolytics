package cache

import (
	"github.com/mjmar01/harmolytics/pkg/hmybebop"
	"github.com/mjmar01/harmolytics/pkg/types"
	"github.com/syndtr/goleveldb/leveldb/util"
	"sync"
)

var txPrefix = []byte{0x01}
var txMemoryByEthHash = map[string]*types.Transaction{}
var txMemoryByHash = map[string]*types.Transaction{}
var txMutex = sync.RWMutex{}
var loadedTx = false

func (c *Cache) GetTransaction(hash string) (tx *types.Transaction, ok bool) {
	txMutex.RLock()
	tx, ok = txMemoryByEthHash[hash]
	if !ok {
		tx, ok = txMemoryByHash[hash]
	}
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
	txMemoryByEthHash[tx.EthTxHash] = tx
	txMemoryByHash[tx.TxHash] = tx
	txMutex.Unlock()
	return tx, true
}

func (c *Cache) GetTransactionByFilter(include func(m *types.Transaction) bool) (txs []*types.Transaction) {
	if !loadedTx {
		c.loadTxMemory()
	}
	for _, tx := range txMemoryByHash {
		in := include(tx)
		if in {
			txs = append(txs, tx)
		}
	}
	return
}

func (c *Cache) SetTransaction(tx *types.Transaction) {
	txMutex.Lock()
	txMemoryByEthHash[tx.EthTxHash] = tx
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
			txMemoryByEthHash[txPtr.EthTxHash] = txPtr
			txMemoryByHash[txPtr.TxHash] = txPtr
			txMutex.Unlock()
			wg.Done()
		}(cp)
	}
	wg.Wait()
	iter.Release()
	loadedTx = true
}

func transactionKey(hash string) []byte {
	return append(txPrefix, []byte(hash)...)
}
