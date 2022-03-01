// Package token handles reading token information from a harmony RPC API
package token

import (
	"harmolytics/harmony"
	"harmolytics/harmony/address"
	"harmolytics/harmony/hex"
	"harmolytics/harmony/rpc"
	"math/big"
)

const (
	getBalanceMethod = "0x70a08231"
)

func GetTokens(addrs []string) (ts []harmony.Token, err error) {
	var inputs []harmony.Address
	for _, addr := range addrs {
		a, err := address.New(addr)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, a)
	}
	ts, err = rpc.GetTokens(inputs)
	return
}

// GetBalanceOf returns the amount harmony.Address a holds of harmony.Token t at the given block. 0 to use the latest block.
func GetBalanceOf(a harmony.Address, t harmony.Token, blockNum uint64) (balance *big.Int, err error) {
	var rawBalance string
	input := getBalanceMethod + hex.EncodeAddress(a)
	if blockNum == 0 {
		rawBalance, err = rpc.SimpleCall(t.Address.EthAddress.Hex(), input)
		if err != nil {
			return
		}
	} else {
		rawBalance, err = rpc.HistoricCall(t.Address.EthAddress.Hex(), input, blockNum)
		if err != nil {
			return
		}
	}
	balance, err = hex.DecodeInt(rawBalance, 0)
	if err != nil {
		return
	}
	return
}
