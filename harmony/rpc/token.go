package rpc

import (
	"encoding/json"
	"github.com/go-errors/errors"
	"harmolytics/harmony"
	"harmolytics/harmony/hex"
)

const (
	getNameMethod     = "0x06fdde03"
	getSymbolMethod   = "0x95d89b41"
	getDecimalsMethod = "0x313ce567"
)

func GetTokens(addrs []harmony.Address) (ts []harmony.Token, err error) {
	var reply rpcReplyS

	// Get valid tokens
	start := queryId
	for _, addr := range addrs {
		body := newRpcBody(contractCall)
		body.Params = params(addr, getDecimalsMethod)
		err = conn.WriteJSON(body)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
	}

	for i := 0; i < len(addrs); i++ {
		_, ret, err := conn.ReadMessage()
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		err = json.Unmarshal(ret, &reply)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		if reply.Result == "0x" {
			continue
		}
		decimals, err := hex.DecodeInt(reply.Result, 0)
		if err != nil {
			return nil, err
		}
		idx := reply.Id - start
		ts = append(ts, harmony.Token{
			Address:  addrs[idx],
			Decimals: int(decimals.Int64()),
		})
	}

	// Get the rest
	start = queryId
	for i, _ := range ts {
		body := newRpcBody(contractCall)
		body.Params = params(ts[i].Address, getNameMethod)
		err = conn.WriteJSON(body)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		body = newRpcBody(contractCall)
		body.Params = params(ts[i].Address, getSymbolMethod)
		err = conn.WriteJSON(body)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
	}

	for i := 0; i < len(ts)*2; i++ {
		_, ret, err := conn.ReadMessage()
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		err = json.Unmarshal(ret, &reply)
		idx := reply.Id - start
		if idx&1 == 0 {
			name, err := hex.DecodeString(reply.Result, 0)
			if err != nil {
				return nil, err
			}
			ts[idx/2].Name = name
		} else {
			symbol, err := hex.DecodeString(reply.Result, 0)
			if err != nil {
				return nil, err
			}
			ts[idx/2].Symbol = symbol
		}
	}
	return
}

func params(addr harmony.Address, method string) []interface{} {
	return []interface{}{
		map[string]string{
			"to":   addr.EthAddress.Hex(),
			"data": method,
		},
		"latest",
	}
}
