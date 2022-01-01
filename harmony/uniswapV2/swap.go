package uniswapV2

import (
	hexEncoding "encoding/hex"
	"harmolytics/harmony"
	"harmolytics/harmony/hex"
	"math/big"
	"sort"
)

const (
	swapEvent    = "0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822"
	swapEth      = "fb3bdb41"
	swapExEth    = "7ff36ab5"
	swapExEthSup = "b6f9de95"
)

func DecodeSwap(tx harmony.Transaction) (s harmony.Swap, err error) {
	// Get involved tokens
	pathOffset := getPathOffset(tx.Method.Signature)
	path, err := hex.DecodeArray(tx.Input[8:], pathOffset)
	if err != nil {
		return
	}
	pathLeft := len(path) - 1
	inputAddr, err := hex.DecodeAddress(hexEncoding.EncodeToString(path[0]), 0)
	if err != nil {
		return
	}
	outputAddr, err := hex.DecodeAddress(hexEncoding.EncodeToString(path[len(path)-1]), 0)
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
				aIn0, err := hex.DecodeInt(txLog.Data, 0)
				if err != nil {
					return harmony.Swap{}, err
				}
				aIn1, err := hex.DecodeInt(txLog.Data, 1)
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
				aOut0, err := hex.DecodeInt(txLog.Data, 2)
				if err != nil {
					return harmony.Swap{}, err
				}
				aOut1, err := hex.DecodeInt(txLog.Data, 3)
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

func getPathOffset(methodSig string) int {
	switch methodSig {
	case swapEth, swapExEth, swapExEthSup:
		return 1
	default:
		return 2
	}
}
