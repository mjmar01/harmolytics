package uniswapV2

import (
	"encoding/hex"
	"github.com/mjmar01/harmolytics/pkg/harmony"
	"github.com/mjmar01/harmolytics/pkg/solidityio"
	"math/big"
	"sort"
)

const (
	swapEvent    = "0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822"
	swapEthExTk  = "fb3bdb41"
	swapExEthTk  = "7ff36ab5"
	swapExEthSup = "b6f9de95"
	swapTkExEth  = "4a25d94a"
	swapTkExTk   = "8803dbee"
)

func DecodeSwap(tx harmony.Transaction) (s harmony.Swap, err error) {
	// Get involved tokens
	pathOffset := getPathOffset(tx.Method.Signature)
	path, err := solidityio.DecodeArray(tx.Input[8:], pathOffset)
	if err != nil {
		return
	}
	pathLeft := len(path) - 1
	inputAddr, err := solidityio.DecodeAddress(hex.EncodeToString(path[0]), 0)
	if err != nil {
		return
	}
	outputAddr, err := solidityio.DecodeAddress(hex.EncodeToString(path[len(path)-1]), 0)
	if err != nil {
		return
	}
	// Fill part of the swap
	lpPath, err := DecodeLiquidity(tx)
	if err != nil {
		return
	}
	s = harmony.Swap{
		TxHash:   tx.TxHash,
		InToken:  harmony.Token{Address: inputAddr},
		OutToken: harmony.Token{Address: outputAddr},
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
				aIn0, err := solidityio.DecodeInt(txLog.Data, 0)
				if err != nil {
					return harmony.Swap{}, err
				}
				aIn1, err := solidityio.DecodeInt(txLog.Data, 1)
				if err != nil {
					return harmony.Swap{}, err
				}
				inputAmount = inputAmount.Or(aIn0, aIn1)
				s.InAmount = inputAmount
			}
			pathLeft--
			// If it's the last swap of the path read output amount
			if pathLeft == 0 {
				outputAmount := big.NewInt(0)
				aOut0, err := solidityio.DecodeInt(txLog.Data, 2)
				if err != nil {
					return harmony.Swap{}, err
				}
				aOut1, err := solidityio.DecodeInt(txLog.Data, 3)
				if err != nil {
					return harmony.Swap{}, err
				}
				outputAmount = outputAmount.Or(aOut0, aOut1)
				s.OutAmount = outputAmount
				return s, err
			}
		}
	}
	return
}

func AnalyzeFees(swap *harmony.Swap, reserves []harmony.HistoricLiquidityRatio, tx harmony.Transaction) (err error) {
	side := getVariableSide(tx.Method.Signature)
	pathOffset := getPathOffset(tx.Method.Signature)
	tokenPath, err := solidityio.DecodeArray(tx.Input[8:], pathOffset)
	if side == 1 {
		swap.FeeToken = swap.InToken.Address.OneAddress
		feeAmount := swap.OutAmount
		for i := len(tokenPath) - 1; i > 0; i-- {
			var reserveIn, reserveOut *big.Int
			s := hex.EncodeToString(tokenPath[i])[24:]
			a := hex.EncodeToString(reserves[i-1].LP.TokenA.Address.EthAddress.Bytes())
			b := hex.EncodeToString(reserves[i-1].LP.TokenB.Address.EthAddress.Bytes())
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
	}
	if side == 2 {
		swap.FeeToken = swap.OutToken.Address.OneAddress
		feeAmount := swap.InAmount
		for i := 0; i < len(tokenPath)-1; i++ {
			var reserveIn, reserveOut *big.Int
			s := hex.EncodeToString(tokenPath[i])[24:]
			a := hex.EncodeToString(reserves[i].LP.TokenA.Address.EthAddress.Bytes())
			b := hex.EncodeToString(reserves[i].LP.TokenB.Address.EthAddress.Bytes())
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
	case swapEthExTk, swapExEthTk, swapExEthSup:
		return 1
	default:
		return 2
	}
}

func getVariableSide(methodSig string) int {
	switch methodSig {
	case swapEthExTk, swapTkExEth, swapTkExTk:
		return 1
	default:
		return 2
	}
}
