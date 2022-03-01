// Package token handles reading token information from a harmony RPC API
package token

import (
	"harmolytics/harmony"
	"harmolytics/harmony/address"
	"harmolytics/harmony/rpc"
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
