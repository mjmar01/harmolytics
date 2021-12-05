// Package token handles reading token information from a harmony RPC API
package token

import (
	"harmolytics/harmony"
	"harmolytics/harmony/hex"
	"harmolytics/harmony/rpc"
)

const (
	getNameMethod     = "0x06fdde03"
	getSymbolMethod   = "0x95d89b41"
	getDecimalsMethod = "0x313ce567"
)

// GetToken takes a harmony.Address and returns a harmony.Token or an empty Token if the address isn't an HRC-20 token
func GetToken(a harmony.Address) (t harmony.Token, err error) {
	// Check if this address is an NFT
	rawDecimals, err := rpc.SimpleCall(a.EthAddress.String(), getDecimalsMethod)
	if err != nil {
		return
	}
	if rawDecimals == "0x" {
		// This isn't a token. It's an NFT!
		return harmony.Token{}, nil
	}
	// Get token data from contract
	rawName, err := rpc.SimpleCall(a.EthAddress.String(), getNameMethod)
	if err != nil {
		return
	}
	rawSymbol, err := rpc.SimpleCall(a.EthAddress.String(), getSymbolMethod)
	if err != nil {
		return
	}
	// Read return values
	decimals, err := hex.ReadInt(rawDecimals, 0)
	if err != nil {
		return
	}
	name, err := hex.ReadString(rawName, 0)
	if err != nil {
		return
	}
	symbol, err := hex.ReadString(rawSymbol, 0)
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
