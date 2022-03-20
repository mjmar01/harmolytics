package mysql

import (
	_ "database/sql"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/internal/log"
	"github.com/mjmar01/harmolytics/pkg/harmony"
	"github.com/mjmar01/harmolytics/pkg/harmony/address"
	"math/big"
	"strings"
)

const (
	swapsQuery         = "SELECT hash, input_token, output_token, input_amount, output_amount, path FROM harmolytics_profile_%s.swaps"
	ratiosQuery        = "SELECT liquidity_pool, block_num, reserve_a, reserve_b FROM harmolytics_historic.liquidity_ratios"
	liquidityPoolQuery = "SELECT token_a, token_b FROM harmolytics_default.liquidity_pools WHERE address = '%s'"
	blockQuery         = "SELECT block_num FROM harmolytics_profile_%s.transactions WHERE hash = '%s'"
	methodQuery        = "SELECT signature, name, parameters FROM harmolytics_default.methods WHERE signature = '%s'"
)

func RunQuery(query string) (err error) {
	rows, err := db.Query(query)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = rows.Close()
	if err != nil {
		return errors.Wrap(err, 0)
	}
	return
}

// GetStringsByQuery takes a query that returns a single column and returns the rows as a list of strings.
func GetStringsByQuery(query string) (r []string, err error) {
	log.Task("Getting strings from database", log.TraceLevel)
	rows, err := db.Query(query)
	defer rows.Close()
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		r = append(r, s)
	}
	log.Done()
	return
}

// GetStringByQuery takes a query that returns a single column and row and returns this value as string.
func GetStringByQuery(query string) (s string, err error) {
	log.Task("Getting string from database", log.TraceLevel)
	rows, err := db.Query(query)
	defer rows.Close()
	if err != nil {
		return "", errors.Wrap(err, 0)
	}
	rows.Next()
	err = rows.Scan(&s)
	if err != nil {
		return "", errors.Wrap(err, 0)
	}
	log.Done()
	return
}

func GetMethodBySignature(sig string) (m harmony.Method, err error) {
	log.Task("Getting method from database by signature", log.TraceLevel)
	rows, err := db.Query(fmt.Sprintf(methodQuery, sig))
	defer rows.Close()
	if err != nil {
		return harmony.Method{}, errors.Wrap(err, 0)
	}
	if !(rows.Next()) {
		log.Done()
		return harmony.Method{Signature: sig}, nil
	}
	var p string
	err = rows.Scan(&m.Signature, &m.Name, &p)
	if err != nil {
		return harmony.Method{}, errors.Wrap(err, 0)
	}
	m.Parameters = strings.Split(p, ":")
	log.Done()
	return
}

func GetSwaps() (swaps []harmony.Swap, err error) {
	log.Task("Getting swaps from database", log.TraceLevel)
	rows, err := db.Query(fmt.Sprintf(swapsQuery, prfl))
	defer rows.Close()
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	for rows.Next() {
		var s harmony.Swap
		var tokenA, tokenB, amountA, amountB, path string
		err = rows.Scan(&s.TxHash, &tokenA, &tokenB, &amountA, &amountB, &path)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		tokenAddress, err := address.New(tokenA)
		if err != nil {
			return nil, err
		}
		s.InToken = harmony.Token{Address: tokenAddress}
		tokenAddress, err = address.New(tokenB)
		if err != nil {
			return nil, err
		}
		s.OutToken = harmony.Token{Address: tokenAddress}
		s.InAmount = new(big.Int)
		s.OutAmount = new(big.Int)
		s.InAmount.SetString(amountA, 10)
		s.OutAmount.SetString(amountB, 10)
		for _, p := range strings.Split(path, ":") {
			lpAddr, err := address.New(p)
			if err != nil {
				return nil, err
			}
			s.Path = append(s.Path, harmony.LiquidityPool{LpToken: harmony.Token{Address: lpAddr}})
		}
		swaps = append(swaps, s)
	}
	log.Done()
	return
}

func GetLiquidityPool(lpAddr string) (lp harmony.LiquidityPool, err error) {
	log.Task("Getting liquidity pools from database", log.TraceLevel)
	rows, err := db.Query(fmt.Sprintf(liquidityPoolQuery, lpAddr))
	defer rows.Close()
	if err != nil {
		return harmony.LiquidityPool{}, errors.Wrap(err, 0)
	}
	rows.Next()
	var a, b string
	err = rows.Scan(&a, &b)
	if err != nil {
		return harmony.LiquidityPool{}, errors.Wrap(err, 0)
	}
	addr, err := address.New(lpAddr)
	if err != nil {
		return harmony.LiquidityPool{}, err
	}
	lp.LpToken.Address = addr
	lp.TokenA, err = GetToken(a)
	if err != nil {
		return harmony.LiquidityPool{}, err
	}
	lp.TokenB, err = GetToken(b)
	if err != nil {
		return harmony.LiquidityPool{}, err
	}
	log.Done()
	return
}

func GetRatios() (ratios []harmony.HistoricLiquidityRatio, err error) {
	log.Task("Getting ratios from database", log.TraceLevel)
	rows, err := db.Query(ratiosQuery)
	defer rows.Close()
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	for rows.Next() {
		var r harmony.HistoricLiquidityRatio
		var lp, amountA, amountB string
		err = rows.Scan(&lp, &r.BlockNum, &amountA, &amountB)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		r.LP, err = GetLiquidityPool(lp)
		if err != nil {
			return
		}
		r.ReserveA = new(big.Int)
		r.ReserveB = new(big.Int)
		r.ReserveA.SetString(amountA, 10)
		r.ReserveB.SetString(amountB, 10)
		ratios = append(ratios, r)
	}
	log.Done()
	return
}

func GetBlockByTx(tx string) (block uint64, err error) {
	rows, err := db.Query(fmt.Sprintf(blockQuery, prfl, tx))
	defer rows.Close()
	if err != nil {
		return 0, errors.Wrap(err, 0)
	}
	rows.Next()
	err = rows.Scan(&block)
	if err != nil {
		return 0, errors.Wrap(err, 0)
	}
	return
}