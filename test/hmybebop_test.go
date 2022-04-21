package test

import (
	"bytes"
	"encoding/gob"
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/pkg/hmybebop"
	"github.com/mjmar01/harmolytics/pkg/types"
	"testing"
)

func TestTransactionBebop(t *testing.T) {
	t.Parallel()
	data, err := hmybebop.EncodeTransaction(tx)
	if err != nil {
		t.Fatal(err.(*errors.Error).ErrorStack())
	}
	out, err := hmybebop.DecodeTransaction(data)
	if err != nil {
		t.Fatal(err.(*errors.Error).ErrorStack())
	}

	if tx.TxHash != out.TxHash {
		t.Errorf("Processed transaction does not have same TxHash: %s|%s", tx.TxHash, out.TxHash)
	}
	if tx.Sender.OneAddress != out.Sender.OneAddress {
		t.Errorf("Processed transaction does not have same sender: %s|%s", tx.Sender.OneAddress, out.Sender.OneAddress)
	}
	if tx.Value.String() != out.Value.String() {
		t.Errorf("Processed transaction does not have same value: %s|%s", tx.Value.String(), out.Value.String())
	}
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
