package test

import (
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/pkg/hmydecode"
	"github.com/mjmar01/harmolytics/pkg/hmyload"
	"testing"
)

const (
	swapTx = "0x22c07f6246dd502bd8b8b7accb1cee30b27455da2ea3cd2093668d262625b809"
)

func TestDecodeSwap(t *testing.T) {
	t.Parallel()
	ldr, _ := hmyload.NewLoader(url, nil)
	txs, _ := ldr.GetFullTransactions(swapTx)
	tx := txs[0]

	swp, ok, err := hmydecode.DecodeSwap(tx)
	if err != nil {
		t.Error(err.(*errors.Error).ErrorStack())
	}
	if !ok {
		t.Errorf("DecodeSwap incorrectly returned !ok")
	}
	if len(swp.Path) != 2 {
		t.Errorf("DecodeSwap returned incorrect path length: %d", len(swp.Path))
	}
	if swp.OutToken.Address.OneAddress != "one16azr8vv8eu96nx9dn035s6ujn3mgz5s4nnea3w" {
		t.Errorf("DecodeSwap returned incorrect OutToken: %s", swp.OutToken.Address.OneAddress)
	}
	if swp.OutAmount.String() != "544843435652907496939" {
		t.Errorf("DecodeSwap returned incorrect OutAmount: %s", swp.OutAmount.String())
	}
}

func TestDecodeTokenTransfers(t *testing.T) {
	t.Parallel()
	ldr, err := hmyload.NewLoader(url, nil)
	if err != nil {
		t.Error(err.(*errors.Error).ErrorStack())
	}
	txs, _ := ldr.GetFullTransactions(swapTx)
	tx := txs[0]

	tkTxs, err := hmydecode.DecodeTokenTransaction(tx)
	if err != nil {
		t.Error(err.(*errors.Error).ErrorStack())
	}
	if len(tkTxs) != 3 {
		t.Errorf("DecodeTokenTransaction returned incorrect number of transfers: %d", len(tkTxs))
	}
	if tkTxs[0].Token.Address.OneAddress != "one1eanyppa9hvpr0g966e6zs5hvdjxkngn6jtulua" {
		t.Errorf("DecodeTokenTransaction returned incorrect Token address: %s", tkTxs[0].Token.Address.OneAddress)
	}
	if tkTxs[0].Amount.String() != "5000000000000000000000" {
		t.Errorf("DecodeTokenTransaction returned incorrect Token amount: %s", tkTxs[0].Amount.String())
	}
}
