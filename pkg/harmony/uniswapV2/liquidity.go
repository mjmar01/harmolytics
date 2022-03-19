package uniswapV2

import (
	hexEncoding "encoding/hex"
	"github.com/mjmar01/harmolytics/pkg/harmony"
	"github.com/mjmar01/harmolytics/pkg/harmony/address"
	"github.com/mjmar01/harmolytics/pkg/harmony/hex"
	"github.com/mjmar01/harmolytics/pkg/harmony/transaction"
	"github.com/mjmar01/harmolytics/pkg/rpc"
	"sort"
)

const (
	addLiquidity       = "e8e33700"
	addLiquidityEth    = "f305d719"
	removeLiquidity    = "baa2abde"
	removeLiquidityEth = "02751cec"
	wone               = "one1eanyppa9hvpr0g966e6zs5hvdjxkngn6jtulua"
)

func GetLiquidityRatios(lps []harmony.HistoricLiquidityRatio) (rs []harmony.HistoricLiquidityRatio, err error) {
	var addrs []harmony.Address
	var tokens []harmony.Token
	var blockNumsInput []uint64
	for i := 0; i < len(lps); i++ {
		addrs = append(addrs, lps[i].LP.LpToken.Address)
		addrs = append(addrs, lps[i].LP.LpToken.Address)
		tokens = append(tokens, lps[i].LP.TokenA)
		tokens = append(tokens, lps[i].LP.TokenB)
		blockNumsInput = append(blockNumsInput, lps[i].BlockNum)
		blockNumsInput = append(blockNumsInput, lps[i].BlockNum)
	}
	balances, err := rpc.GetBalances(addrs, tokens, blockNumsInput)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(lps)*2; i += 2 {
		rs = append(rs, harmony.HistoricLiquidityRatio{
			LP:       lps[i/2].LP,
			BlockNum: lps[i/2].BlockNum,
			ReserveA: balances[i],
			ReserveB: balances[i+1],
		})
	}
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

// DecodeLiquidity receives a swap transaction and returns the list of involved liquidity pools
func DecodeLiquidity(tx harmony.Transaction) (lps []harmony.LiquidityPool, err error) {
	// Get involved tokens
	pathOffset := getPathOffset(tx.Method.Signature)
	path, err := hex.DecodeArray(tx.Input[8:], pathOffset)
	if err != nil {
		return
	}
	var pathTokens []harmony.Token
	for i := range path {
		addr, err := hex.DecodeAddress(hexEncoding.EncodeToString(path[i]), 0)
		if err != nil {
			return nil, err
		}
		pathTokens = append(pathTokens, harmony.Token{Address: addr})
	}
	// Get token transfers
	ttxs, err := transaction.DecodeTokenTransaction(tx)
	if err != nil {
		return
	}
	// Crosscheck path and ttxs
path:
	for i := 0; i < len(path)-1; i++ {
		for _, ttx := range ttxs {
			if ttx.Token.Address.OneAddress == pathTokens[i].Address.OneAddress {
				for _, ttx2 := range ttxs {
					if ttx2.Sender.OneAddress == ttx.Receiver.OneAddress &&
						ttx2.Token.Address.OneAddress == pathTokens[i+1].Address.OneAddress {
						sortedTokens := []harmony.Token{
							{Address: pathTokens[i].Address},
							{Address: pathTokens[i+1].Address},
						}
						sort.Slice(sortedTokens, func(i, j int) bool {
							return sortedTokens[i].Address.OneAddress < sortedTokens[j].Address.OneAddress
						})
						lps = append(lps, harmony.LiquidityPool{
							TokenA:  sortedTokens[0],
							TokenB:  sortedTokens[1],
							LpToken: harmony.Token{Address: ttx.Receiver},
						})
						continue path
					}
				}
			}
		}
	}
	return
}
