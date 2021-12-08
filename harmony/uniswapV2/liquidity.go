package uniswapV2

import (
	"harmolytics/harmony"
	"harmolytics/harmony/address"
	"harmolytics/harmony/hex"
	"harmolytics/harmony/transaction"
)

const (
	addLiquidity       = "e8e33700"
	addLiquidityEth    = "f305d719"
	removeLiquidity    = "baa2abde"
	removeLiquidityEth = "02751cec"
	wone               = "one1eanyppa9hvpr0g966e6zs5hvdjxkngn6jtulua"
)

func DecodeLiquidity(tx harmony.Transaction) (la harmony.LiquidityAction, err error) {
	la.TxHash = tx.TxHash
	switch tx.Method.Signature {
	case addLiquidityEth, removeLiquidityEth:
		addrA, _ := address.New(wone)
		la.TokenA = harmony.Token{Address: addrA}
		addrB, err := hex.ReadAddress(tx.Input[10:], 0)
		if err != nil {
			return la, err
		}
		la.TokenB = harmony.Token{Address: addrB}
	case addLiquidity, removeLiquidity:
		addrA, err := hex.ReadAddress(tx.Input[10:], 0)
		if err != nil {
			return la, err
		}
		la.TokenA = harmony.Token{Address: addrA}
		addrB, err := hex.ReadAddress(tx.Input[10:], 1)
		if err != nil {
			return la, err
		}
		la.TokenB = harmony.Token{Address: addrB}
	}

	tTxs, err := transaction.DecodeTokenTransaction(tx)
	if err != nil {
		return
	}
	switch tx.Method.Signature {
	case addLiquidityEth, addLiquidity:
		la.Direction = harmony.AddLiquidity
		for _, tTx := range tTxs {
			if tTx.Sender.OneAddress == tx.Sender.OneAddress || tTx.Sender.OneAddress == tx.Receiver.OneAddress {
				if tTx.Token.Address.OneAddress == la.TokenA.Address.OneAddress {
					la.AmountA = tTx.Amount
				} else if tTx.Token.Address.OneAddress == la.TokenB.Address.OneAddress {
					la.AmountB = tTx.Amount
				}
			} else if tTx.Receiver.OneAddress == tx.Sender.OneAddress {
				la.LpToken = tTx.Token
				la.LpAmount = tTx.Amount
			}
		}
	case removeLiquidity, removeLiquidityEth:
		la.Direction = harmony.RemoveLiquidity
		for _, tTx := range tTxs {
			if tTx.Receiver.OneAddress == tx.Sender.OneAddress || tTx.Receiver.OneAddress == tx.Receiver.OneAddress {
				if tTx.Token.Address.OneAddress == la.TokenA.Address.OneAddress {
					la.AmountA = tTx.Amount
				} else if tTx.Token.Address.OneAddress == la.TokenB.Address.OneAddress {
					la.AmountB = tTx.Amount
				}
			} else if tTx.Sender.OneAddress == tx.Sender.OneAddress {
				la.LpToken = tTx.Token
				la.LpAmount = tTx.Amount
			}
		}
	}
	return
}
