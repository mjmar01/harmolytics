package cache

import (
	"github.com/mjmar01/harmolytics/pkg/hmybebop"
	"github.com/mjmar01/harmolytics/pkg/types"
	"github.com/syndtr/goleveldb/leveldb/util"
	"sync"
)

var mPrefix = []byte{0x02}
var mMemory = map[string]*types.Method{}
var mMutex = sync.RWMutex{}

func (c *Cache) GetMethod(sig string) (m *types.Method, ok bool) {
	mMutex.RLock()
	m, ok = mMemory[sig]
	mMutex.RUnlock()
	if ok {
		return
	}
	v, err := c.levelDB.Get(methodKey(sig), nil)
	if err != nil {
		return nil, false
	}
	m, err = hmybebop.DecodeMethod(v)
	if err != nil {
		return nil, false
	}
	mMutex.Lock()
	mMemory[m.Signature] = m
	mMutex.Unlock()
	return m, true
}

func (c *Cache) SetMethod(m *types.Method) {
	mMutex.Lock()
	mMemory[m.Signature] = m
	mMutex.Unlock()
	v, err := hmybebop.EncodeMethod(m)
	if err != nil {
		return
	}
	c.levelDB.Put(methodKey(m.Signature), v, nil)
}

func (c *Cache) loadMMemory() {
	iter := c.levelDB.NewIterator(util.BytesPrefix(mPrefix), nil)
	wg := sync.WaitGroup{}
	for iter.Next() {
		src := iter.Value()
		cp := make([]byte, len(src))
		copy(cp, src)
		go func(in []byte) {
			wg.Add(1)
			mPtr, _ := hmybebop.DecodeMethod(in)
			mMutex.Lock()
			mMemory[mPtr.Signature] = mPtr
			mMutex.Unlock()
			wg.Done()
		}(cp)
	}
	wg.Wait()
	iter.Release()
}

func methodKey(sig string) []byte {
	return append(mPrefix, []byte(sig)...)
}
