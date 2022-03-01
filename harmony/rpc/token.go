package rpc

import (
	"encoding/json"
	"github.com/go-errors/errors"
	"harmolytics/harmony"
	"harmolytics/harmony/hex"
	"math/big"
)

const (
	getNameMethod     = "0x06fdde03"
	getSymbolMethod   = "0x95d89b41"
	getDecimalsMethod = "0x313ce567"
	getBalanceMethod  = "0x70a08231"
)

func GetTokens(addrs []harmony.Address) (ts []harmony.Token, err error) {
	var reply rpcReplyS

	// Get valid tokens
	start := queryId
	for _, addr := range addrs {
		body := newRpcBody(contractCall)
		body.Params = params(addr, getDecimalsMethod, 0)
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
	for i := range ts {
		body := newRpcBody(contractCall)
		body.Params = params(ts[i].Address, getNameMethod, 0)
		err = conn.WriteJSON(body)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		body = newRpcBody(contractCall)
		body.Params = params(ts[i].Address, getSymbolMethod, 0)
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
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
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

func GetBalances(addrs []harmony.Address, tokens []harmony.Token, blockNums []uint64) (bs []*big.Int, err error) {
	bs = make([]*big.Int, len(addrs))
	var reply rpcReplyS
	start := queryId
	for i := 0; i < len(addrs); i++ {
		body := newRpcBody(contractCall)
		body.Params = params(tokens[i].Address, getBalanceMethod+hex.EncodeAddress(addrs[i]), blockNums[i])
		err = historicConn.WriteJSON(body)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
	}
	for i := 0; i < len(addrs); i++ {
		_, ret, err := historicConn.ReadMessage()
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		err = json.Unmarshal(ret, &reply)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		idx := reply.Id - start
		b, err := hex.DecodeInt(reply.Result, 0)
		if err != nil {
			return nil, err
		}
		bs[idx] = b
	}
	return
}

func params(addr harmony.Address, data string, blockNum uint64) (ret []interface{}) {
	ret = append(ret, map[string]string{
		"to":   addr.EthAddress.Hex(),
		"data": data,
	})
	if blockNum == 0 {
		ret = append(ret, "latest")
	} else {
		ret = append(ret, blockNum)
	}
	return ret
}
