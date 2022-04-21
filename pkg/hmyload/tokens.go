package hmyload

import (
	"github.com/mjmar01/harmolytics/pkg/hmysolidityio"
	"github.com/mjmar01/harmolytics/pkg/rpc"
	"github.com/mjmar01/harmolytics/pkg/types"
)

const (
	callMethod     = "hmyv2_call"
	nameMethod     = "0x06fdde03"
	symbolMethod   = "0x95d89b41"
	decimalsMethod = "0x313ce567"
	balanceMethod  = "0x70a08231"
)

func (l *Loader) GetTokens(addrs ...types.Address) (tks []types.Token, err error) {
	// Prepare requests across unique peers
	tks = make([]types.Token, len(addrs))
	bodiesByConn, idx, addrsByConn := make([][]rpc.Body, l.uniqueConnCount), 0, make([][]types.Address, l.uniqueConnCount)
	for _, addr := range addrs {
		addrsByConn[idx] = append(addrsByConn[idx], addr)
		b := l.uniqueConns[idx].NewBody(callMethod, map[string]string{"to": addr.HexAddress, "data": nameMethod}, "latest")
		bodiesByConn[idx] = append(bodiesByConn[idx], b)
		b = l.uniqueConns[idx].NewBody(callMethod, map[string]string{"to": addr.HexAddress, "data": symbolMethod}, "latest")
		bodiesByConn[idx] = append(bodiesByConn[idx], b)
		b = l.uniqueConns[idx].NewBody(callMethod, map[string]string{"to": addr.HexAddress, "data": decimalsMethod}, "latest")
		bodiesByConn[idx] = append(bodiesByConn[idx], b)
		idx++
		if idx == l.uniqueConnCount {
			idx = 0
		}
	}
	// Do requests
	ch := make(chan goTk, len(addrs))
	for i, conn := range l.uniqueConns {
		go func(rpc *rpc.Rpc, bodies []rpc.Body, addrs []types.Address) {
			ress, err := rpc.BatchCall(bodies)
			if err != nil {
				ch <- goTk{err: err}
				return
			}
			// Read each result into a token
			for i := 0; i < len(ress); i += 3 {
				var tk types.Token
				tk.Address = addrs[i/3]
				tk.Name, err = hmysolidityio.DecodeString(ress[i].(string), 0)
				if err != nil {
					ch <- goTk{err: err}
					return
				}
				tk.Symbol, err = hmysolidityio.DecodeString(ress[i+1].(string), 0)
				if err != nil {
					ch <- goTk{err: err}
					return
				}
				rawDecimals, err := hmysolidityio.DecodeInt(ress[i+2].(string), 0)
				if err != nil {
					ch <- goTk{err: err}
					return
				}
				tk.Decimals = int(rawDecimals.Int64())
				ch <- goTk{
					err: nil,
					tk:  tk,
				}
			}
		}(conn, bodiesByConn[i], addrsByConn[i])
	}
	// Read Output
	tkMap := make(map[string]types.Token, len(addrs))
	for i := 0; i < len(addrs); i++ {
		out := <-ch
		if out.err != nil {
			return nil, out.err
		}
		tkMap[out.tk.Address.OneAddress] = out.tk
	}
	for i, addr := range addrs {
		tks[i] = tkMap[addr.OneAddress]
	}
	return
}
