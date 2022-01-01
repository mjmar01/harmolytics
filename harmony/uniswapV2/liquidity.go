package uniswapV2

import (
	"harmolytics/harmony"
	"harmolytics/harmony/address"
	"harmolytics/harmony/hex"
	"harmolytics/harmony/token"
	"harmolytics/harmony/transaction"
	"math/big"
)

const (
	addLiquidity       = "e8e33700"
	addLiquidityEth    = "f305d719"
	removeLiquidity    = "baa2abde"
	removeLiquidityEth = "02751cec"
	wone               = "one1eanyppa9hvpr0g966e6zs5hvdjxkngn6jtulua"
)

// GetLiquidityRatio returns the ratio TokenB/TokenA as in: 1 TokenA = r TokenB
func GetLiquidityRatio(lp harmony.LiquidityPool, blockNum int) (r *big.Rat, err error) {
	AmountA, err := token.GetBalanceOf(lp.LpToken.Address, lp.TokenA, blockNum)
	if err != nil {
		return
	}
	AmountB, err := token.GetBalanceOf(lp.LpToken.Address, lp.TokenB, blockNum)
	if err != nil {
		return
	}
	r.SetFrac(AmountB, AmountA)
	return
}

func DecodeLiquidityAction(tx harmony.Transaction) (la harmony.LiquidityAction, err error) {
	la.TxHash = tx.TxHash
	switch tx.Method.Signature {
	case addLiquidityEth, removeLiquidityEth:
		addrA, _ := address.New(wone)
		la.LP.TokenA = harmony.Token{Address: addrA}
		addrB, err := hex.DecodeAddress(tx.Input[8:], 0)
		if err != nil {
			return la, err
		}
		la.LP.TokenB = harmony.Token{Address: addrB}
	case addLiquidity, removeLiquidity:
		addrA, err := hex.DecodeAddress(tx.Input[8:], 0)
		if err != nil {
			return la, err
		}
		la.LP.TokenA = harmony.Token{Address: addrA}
		addrB, err := hex.DecodeAddress(tx.Input[8:], 1)
		if err != nil {
			return la, err
		}
		la.LP.TokenB = harmony.Token{Address: addrB}
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
				if tTx.Token.Address.OneAddress == la.LP.TokenA.Address.OneAddress {
					la.AmountA = tTx.Amount
				} else if tTx.Token.Address.OneAddress == la.LP.TokenB.Address.OneAddress {
					la.AmountB = tTx.Amount
				}
			} else if tTx.Receiver.OneAddress == tx.Sender.OneAddress {
				la.LP.LpToken = tTx.Token
				la.AmountLP = tTx.Amount
			}
		}
	case removeLiquidity, removeLiquidityEth:
		la.Direction = harmony.RemoveLiquidity
		for _, tTx := range tTxs {
			if tTx.Receiver.OneAddress == tx.Sender.OneAddress || tTx.Receiver.OneAddress == tx.Receiver.OneAddress {
				if tTx.Token.Address.OneAddress == la.LP.TokenA.Address.OneAddress {
					la.AmountA = tTx.Amount
				} else if tTx.Token.Address.OneAddress == la.LP.TokenB.Address.OneAddress {
					la.AmountB = tTx.Amount
				}
			} else if tTx.Sender.OneAddress == tx.Sender.OneAddress {
				la.LP.LpToken = tTx.Token
				la.AmountLP = tTx.Amount
			}
		}
	}
	return
}
