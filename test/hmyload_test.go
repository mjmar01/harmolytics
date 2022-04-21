package test

import (
	"github.com/mjmar01/harmolytics/pkg/hmyload"
	"github.com/mjmar01/harmolytics/pkg/types"
	"testing"
)

func TestGetHistory(t *testing.T) {
	t.Parallel()
	l, err := hmyload.NewLoader(url, &hmyload.Opts{AdditionalConnections: 10})
	defer l.Close()
	if err != nil {
		t.Error(err)
	}
	txs, err := l.GetTransactionsByWallet(types.NewAddress("one15vlc8yqstm9algcf6e94dxqx6y04jcsqjuc3gt"))
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

func TestGetTokens(t *testing.T) {
	t.Parallel()
	l, _ := hmyload.NewLoader(url, nil)
	defer l.Close()
	tks, err := l.GetTokens(
		types.NewAddress("one1eanyppa9hvpr0g966e6zs5hvdjxkngn6jtulua"),
		types.NewAddress("one1t8auuy8kl30ujqt2u229273r2eshvhzpu59sz6"),
		types.NewAddress("one1eanyppa9hvpr0g966e6zs5hvdjxkngn6jtulua"),
	)
	if err != nil {
		t.Error(err)
	}
	if len(tks) != 3 {
		t.Errorf("Result did contain incorrect amount of tokens: %d", len(tks))
	}
	if tks[0].Symbol != "WONE" {
		t.Errorf("Result did contain incorrect token Symbol: %s", tks[0].Symbol)
	}
	if tks[0].Name != tks[2].Name {
		t.Errorf("Result is in incorrect order")
	}
}
