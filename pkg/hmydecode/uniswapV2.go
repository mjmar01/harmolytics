package hmydecode

import (
	"encoding/hex"
	"github.com/mjmar01/harmolytics/internal/helper"
	"github.com/mjmar01/harmolytics/pkg/hmysolidityio"
	"github.com/mjmar01/harmolytics/pkg/types"
	"math/big"
	"sort"
)

const (
	wone = "one1eanyppa9hvpr0g966e6zs5hvdjxkngn6jtulua"

	swapEvent                                             = "0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822"
	swapETHForExactTokens                                 = "fb3bdb41"
	swapExactETHForTokens                                 = "7ff36ab5"
	swapExactETHForTokensSupportingFeeOnTransferTokens    = "b6f9de95"
	swapExactTokensForETH                                 = "18cbafe5"
	swapExactTokensForETHSupportingFeeOnTransferTokens    = "791ac947"
	swapExactTokensForTokens                              = "38ed1739"
	swapExactTokensForTokensSupportingFeeOnTransferTokens = "5c11d795"
	swapTokensForExactETH                                 = "4a25d94a"
	swapTokensForExactTokens                              = "8803dbee"

	addLiquidity       = "e8e33700"
	addLiquidityEth    = "f305d719"
	removeLiquidity    = "baa2abde"
	removeLiquidityEth = "02751cec"
)

func DecodeSwap(tx types.Transaction) (s types.Swap, ok bool, err error) {
	if !checkSwap(tx) {
		return types.Swap{}, false, nil
	}
	// Get involved tokens
	pathOffset := getPathOffset(tx.Method.Signature)
	path, err := hmysolidityio.DecodeArray(tx.Input, pathOffset)
	if err != nil {
		return
	}
	pathLeft := len(path) - 1
	inputAddr, err := hmysolidityio.DecodeAddress(hex.EncodeToString(path[0]), 0)
	if err != nil {
		return
	}
	outputAddr, err := hmysolidityio.DecodeAddress(hex.EncodeToString(path[len(path)-1]), 0)
	if err != nil {
		return
	}
	// Fill part of the swap
	lpPath, _, err := DecodeLiquidity(tx)
	if err != nil {
		return
	}
	s = types.Swap{
		TxHash:   tx.TxHash,
		InToken:  types.Token{Address: inputAddr},
		OutToken: types.Token{Address: outputAddr},
		Path:     lpPath,
	}
	// Read through logs for swap events
	sort.Slice(tx.Logs, func(i, j int) bool {
		return tx.Logs[i].LogIndex < tx.Logs[j].LogIndex
	})
	for _, txLog := range tx.Logs {
		if txLog.Topics[0] == swapEvent {
			// If it's the first swap of the path read input amount
			if pathLeft == len(path)-1 {
				inputAmount := big.NewInt(0)
				aIn0, err := hmysolidityio.DecodeInt(txLog.Data, 0)
				if err != nil {
					return types.Swap{}, false, err
				}
				aIn1, err := hmysolidityio.DecodeInt(txLog.Data, 1)
				if err != nil {
					return types.Swap{}, false, err
				}
				inputAmount = inputAmount.Or(aIn0, aIn1)
				s.InAmount = inputAmount
			}
			pathLeft--
			// If it's the last swap of the path read output amount
			if pathLeft == 0 {
				outputAmount := big.NewInt(0)
				aOut0, err := hmysolidityio.DecodeInt(txLog.Data, 2)
				if err != nil {
					return types.Swap{}, false, err
				}
				aOut1, err := hmysolidityio.DecodeInt(txLog.Data, 3)
				if err != nil {
					return types.Swap{}, false, err
				}
				outputAmount = outputAmount.Or(aOut0, aOut1)
				s.OutAmount = outputAmount
				return s, true, nil
			}
		}
	}
	ok = true
	return
}

// DecodeLiquidity receives a swap transaction and returns the list of involved liquidity pools
func DecodeLiquidity(tx types.Transaction) (lps []types.LiquidityPool, ok bool, err error) {
	if !checkSwap(tx) {
		return nil, false, nil
	}
	// Get involved tokens
	pathOffset := getPathOffset(tx.Method.Signature)
	path, err := hmysolidityio.DecodeArray(tx.Input, pathOffset)
	if err != nil {
		return
	}
	var pathTokens []types.Token
	for i := range path {
		addr, err := hmysolidityio.DecodeAddress(hex.EncodeToString(path[i]), 0)
		if err != nil {
			return nil, false, err
		}
		pathTokens = append(pathTokens, types.Token{Address: addr})
	}
	// Get token transfers
	ttxs, err := DecodeTokenTransaction(tx)
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
						sortedTokens := []types.Token{
							{Address: pathTokens[i].Address},
							{Address: pathTokens[i+1].Address},
						}
						sort.Slice(sortedTokens, func(i, j int) bool {
							return sortedTokens[i].Address.OneAddress < sortedTokens[j].Address.OneAddress
						})
						lps = append(lps, types.LiquidityPool{
							TokenA:  sortedTokens[0],
							TokenB:  sortedTokens[1],
							LpToken: types.Token{Address: ttx.Receiver},
						})
						continue path
					}
				}
			}
		}
	}
	ok = true
	return
}

func AnalyzeFees(swap *types.Swap, reserves []types.HistoricLiquidityRatio, tx types.Transaction) (err error) {
	side := getVariableSide(tx.Method.Signature)
	pathOffset := getPathOffset(tx.Method.Signature)
	tokenPath, err := hmysolidityio.DecodeArray(tx.Input[8:], pathOffset)
	switch side {
	case 1:
		swap.FeeToken = swap.InToken.Address.OneAddress
		feeAmount := swap.OutAmount
		for i := len(tokenPath) - 1; i > 0; i-- {
			var reserveIn, reserveOut *big.Int
			s := hex.EncodeToString(tokenPath[i])[24:]
			a := hex.EncodeToString(reserves[i-1].LP.TokenA.Address.Bytes)
			b := hex.EncodeToString(reserves[i-1].LP.TokenB.Address.Bytes)
			if s == a {
				reserveIn = reserves[i-1].ReserveB
				reserveOut = reserves[i-1].ReserveA
			} else if s == b {
				reserveIn = reserves[i-1].ReserveA
				reserveOut = reserves[i-1].ReserveB
			}
			feeAmount = getAmountIn(feeAmount, reserveIn, reserveOut)
		}
		feeAmount.Sub(swap.InAmount, feeAmount)
		swap.FeeAmount = feeAmount
	case 2:
		swap.FeeToken = swap.OutToken.Address.OneAddress
		feeAmount := swap.InAmount
		for i := 0; i < len(tokenPath)-1; i++ {
			var reserveIn, reserveOut *big.Int
			s := hex.EncodeToString(tokenPath[i])[24:]
			a := hex.EncodeToString(reserves[i].LP.TokenA.Address.Bytes)
			b := hex.EncodeToString(reserves[i].LP.TokenB.Address.Bytes)
			if s == a {
				reserveIn = reserves[i].ReserveA
				reserveOut = reserves[i].ReserveB
			} else if s == b {
				reserveIn = reserves[i].ReserveB
				reserveOut = reserves[i].ReserveA
			}
			feeAmount = getAmountOut(feeAmount, reserveIn, reserveOut)
		}
		feeAmount.Sub(feeAmount, swap.OutAmount)
		swap.FeeAmount = feeAmount
	}
	return
}

func checkSwap(tx types.Transaction) bool {
	return helper.StringInSlice(tx.Method.Signature, []string{
		swapETHForExactTokens,
		swapExactETHForTokens,
		swapExactETHForTokensSupportingFeeOnTransferTokens,
		swapExactTokensForETH,
		swapExactTokensForETHSupportingFeeOnTransferTokens,
		swapExactTokensForTokens,
		swapExactTokensForTokensSupportingFeeOnTransferTokens,
		swapTokensForExactETH,
		swapTokensForExactTokens,
	})
}

func getAmountIn(amountOut, reserveIn, reserveOut *big.Int) *big.Int {
	numerator := new(big.Int)
	numerator.Mul(reserveIn, amountOut)
	denominator := new(big.Int)
	denominator.Sub(reserveOut, amountOut)
	result := new(big.Int)
	result.Div(numerator, denominator)
	result.Add(result, big.NewInt(1))
	return result
}

func getAmountOut(amountIn, reserveIn, reserveOut *big.Int) *big.Int {
	numerator := new(big.Int)
	numerator.Mul(amountIn, reserveOut)
	denominator := new(big.Int)
	denominator.Add(reserveIn, amountIn)
	result := new(big.Int)
	result.Div(numerator, denominator)
	return result
}

func getPathOffset(methodSig string) int {
	switch methodSig {
	case swapETHForExactTokens, swapExactETHForTokens, swapExactETHForTokensSupportingFeeOnTransferTokens:
		return 1
	default:
		return 2
	}
}

func getVariableSide(methodSig string) int {
	switch methodSig {
	case swapETHForExactTokens, swapTokensForExactETH, swapTokensForExactTokens:
		return 1
	default:
		return 2
	}
}
