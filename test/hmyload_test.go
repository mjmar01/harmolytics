package test

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/pkg/hmyload"
	"github.com/mjmar01/harmolytics/pkg/rpc"
	"github.com/mjmar01/harmolytics/pkg/types"
	"strings"
	"testing"
	"time"
)

func TestGetHistory(t *testing.T) {
	t.Parallel()
	l, err := hmyload.NewLoader(url, &hmyload.Opts{AdditionalConnections: 10, ExistingCache: centralCache})
	defer l.Close()
	if err != nil {
		t.Fatal(err.(*errors.Error).ErrorStack())
	}
	txs, err := l.GetTransactionsByWallet(types.NewAddress("0x42813a05ec9c7e17af2d1499f9b0a591b7619abf"))
	if err != nil {
		t.Fatal(err.(*errors.Error).ErrorStack())
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
	l, err := hmyload.NewLoader(url, &hmyload.Opts{ExistingCache: centralCache})
	defer l.Close()
	if err != nil {
		t.Fatal(err.(*errors.Error).ErrorStack())
	}
	txs, err := l.GetFullTransactions("0xf916accb28b218085da083f2df398d66f65ce175e32a38ea232debf708b2cc84", "0xf916accb28b218085da083f2df398d66f65ce175e32a38ea232debf708b2cc84")
	if err != nil {
		t.Fatal(err.(*errors.Error).ErrorStack())
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
	l, err := hmyload.NewLoader(url, &hmyload.Opts{ExistingCache: centralCache})
	defer l.Close()
	if err != nil {
		t.Fatal(err.(*errors.Error).ErrorStack())
	}
	tks, err := l.GetTokens(
		types.NewAddress("one1eanyppa9hvpr0g966e6zs5hvdjxkngn6jtulua"),
		types.NewAddress("one1t8auuy8kl30ujqt2u229273r2eshvhzpu59sz6"),
		types.NewAddress("one1eanyppa9hvpr0g966e6zs5hvdjxkngn6jtulua"),
	)
	if err != nil {
		t.Fatal(err.(*errors.Error).ErrorStack())
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

func TestAltHistory(t *testing.T) {
	t1 := time.Now()
	var testBodies []rpc.Body
	for i := 0; i < 80; i++ {
		testBodies = append(testBodies, defaultRPC.NewBody(
			"hmyv2_getTransactionsHistory",
			map[string]interface{}{
				"address":   "0x42813a05ec9c7e17af2d1499f9b0a591b7619abf",
				"pageIndex": i,
				"pageSize":  5000,
				"fullTx":    false,
				"txType":    "ALL",
			},
		))
	}
	fmt.Printf("Allocate bodies: %s\n", time.Since(t1))
	t1 = time.Now()
	ress, err := defaultRPC.BatchCall(testBodies)
	fmt.Println(err)
	sum := 0
	for _, res := range ress {
		sum += strings.Count(string(res), "\",\"")
	}
	fmt.Printf("Batch Call: %s\n", time.Since(t1))
	testBodies = []rpc.Body{}
	t1 = time.Now()
	for i := 0; i < sum; i++ {
		testBodies = append(testBodies, defaultRPC.NewBody("hmyv2_getTransactionReceipt", "0x771d2da16e07d81c63f2e7cf22418e5e98b5b57438de8005f5b144cfbe6867ba"))
	}
	fmt.Printf("Allocate bodies: %s\n", time.Since(t1))

	t1 = time.Now()
	for i := 0; i < 20000; i += 5000 {
		ress, err = defaultRPC.RawBatchCall(testBodies[i : i+5000])
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("Batch Call: %s\n", time.Since(t1))
}
