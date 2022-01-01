// Package token handles reading token information from a harmony RPC API
package token

import (
	"harmolytics/harmony"
	"harmolytics/harmony/hex"
	"harmolytics/harmony/rpc"
	"math/big"
)

const (
	getNameMethod     = "0x06fdde03"
	getSymbolMethod   = "0x95d89b41"
	getDecimalsMethod = "0x313ce567"
	getBalanceMethod  = "0x70a08231"
)

// GetToken takes a harmony.Address and returns a harmony.Token or an empty Token if the address isn't an HRC-20 token
func GetToken(a harmony.Address) (t harmony.Token, err error) {
	// Check if this address is an NFT
	rawDecimals, err := rpc.SimpleCall(a.EthAddress.Hex(), getDecimalsMethod)
	if err != nil {
		return
	}
	if rawDecimals == "0x" {
		// This isn't a token. It's an NFT!
		return harmony.Token{}, nil
	}
	// Get token data from contract
	rawName, err := rpc.SimpleCall(a.EthAddress.Hex(), getNameMethod)
	if err != nil {
		return
	}
	rawSymbol, err := rpc.SimpleCall(a.EthAddress.Hex(), getSymbolMethod)
	if err != nil {
		return
	}
	// Read return values
	decimals, err := hex.DecodeInt(rawDecimals, 0)
	if err != nil {
		return
	}
	name, err := hex.DecodeString(rawName, 0)
	if err != nil {
		return
	}
	symbol, err := hex.DecodeString(rawSymbol, 0)
	if err != nil {
		return
	}
	// Fill token
	t = harmony.Token{
		Address:  a,
		Name:     name,
		Symbol:   symbol,
		Decimals: int(decimals.Int64()),
	}
	return
}

// GetBalanceOf returns the amount harmony.Address a holds of harmony.Token t at the given block. 0 to use the latest block.
func GetBalanceOf(a harmony.Address, t harmony.Token, blockNum int) (balance *big.Int, err error) {
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
