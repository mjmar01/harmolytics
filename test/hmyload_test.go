package test

import (
	"github.com/mjmar01/harmolytics/pkg/harmony"
	"github.com/mjmar01/harmolytics/pkg/hmyload"
	"testing"
)

func TestHistory(t *testing.T) {
	t.Parallel()
	l, err := hmyload.NewLoader(url, &hmyload.Opts{AdditionalConnections: 10})
	defer l.Close()
	if err != nil {
		t.Error(err)
	}
	txs, err := l.GetTransactionsByWallet(harmony.NewAddress("one15vlc8yqstm9algcf6e94dxqx6y04jcsqjuc3gt"))
	if err != nil {
		t.Error(err)
	}
	if len(txs) < 3 {
		t.Errorf("Response is missing transactions got %d expected 3", len(txs))
	}
	if txs[0].TxHash == "" {
		t.Errorf("Response contains empty transaction")
	}
}

func TestGetFullTransaction(t *testing.T) {
	t.Parallel()
	l, err := hmyload.NewLoader(url, nil)
	defer l.Close()
	if err != nil {
		t.Error(err)
	}
	txs, err := l.GetFullTransactions("0xf916accb28b218085da083f2df398d66f65ce175e32a38ea232debf708b2cc84", "0xf916accb28b218085da083f2df398d66f65ce175e32a38ea232debf708b2cc84")
	if err != nil {
		t.Error(err)
	}
	if txs[0].Status != 1 {
		t.Errorf("Result did not contain correct Status: %v", txs[0].Status)
	}
	if txs[0].BlockNum != 24658150 {
		t.Errorf("Result did not contain correct BlockNum: %v", txs[0].BlockNum)
	}
	if len(txs[0].Logs) != 6 {
		t.Errorf("Result did contian incorrect number of Logs: %v", len(txs[0].Logs))
	}
	if txs[0].TxHash != txs[1].TxHash {
		t.Errorf("Result is not consistent")
	}
}
