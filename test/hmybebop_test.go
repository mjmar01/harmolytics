package test

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/mjmar01/harmolytics/pkg/hmybebop"
	"github.com/mjmar01/harmolytics/pkg/types"
	"testing"
)

func TestTransactionBebop(t *testing.T) {
	t.Parallel()
	data, _ := hmybebop.EncodeTransaction(tx)
	out, _ := hmybebop.DecodeTransaction(data)
	fmt.Println(out.TxHash)
}

func BenchmarkBebopEncode(b *testing.B) {
	var data []byte
	for i := 0; i < b.N; i++ {
		data, _ = hmybebop.EncodeTransaction(tx)
	}
	dump = data
}

func BenchmarkGobEncode(b *testing.B) {
	var data []byte
	for i := 0; i < b.N; i++ {
		buff := new(bytes.Buffer)
		enc := gob.NewEncoder(buff)
		enc.Encode(tx)
		data = buff.Bytes()
	}
	dump = data
}

func BenchmarkBebopDecode(b *testing.B) {
	var tmp *types.Transaction
	for i := 0; i < b.N; i++ {
		tmp, _ = hmybebop.DecodeTransaction(txBebop)
	}
	dump = tmp
}

func BenchmarkGobDecode(b *testing.B) {
	var tmp types.Transaction
	for i := 0; i < b.N; i++ {
		dec := gob.NewDecoder(bytes.NewReader(txGob))
		dec.Decode(&tmp)
	}
	dump = tmp
}
