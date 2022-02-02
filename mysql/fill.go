package mysql

import (
	"bytes"
	_ "embed"
	"github.com/go-errors/errors"
	"harmolytics/harmony"
	"text/template"
)

//go:embed queries/fill_transactions.tmpl
var transactionsQ string

//go:embed queries/fill_tokens.tmpl
var tokensQ string

//go:embed queries/fill_methods.tmpl
var methodsQ string

//go:embed queries/fill_swaps.tmpl
var swapsQ string

//go:embed queries/fill_liquidity_actions.tmpl
var liquidityActionsQ string

//go:embed queries/fill_liquidity_pools.tmpl
var liquidityPoolsQ string

//go:embed queries/fill_liquidity_ratios.tmpl
var liquidityRatiosQ string

//go:embed queries/fill_tokenTransfers.tmpl
var tokenTransfersQ string

//go:embed queries/fill_fees.tmpl
var swapFeesQ string

// SetTransactions takes a list of transaction.Transaction and saves those to the tables transactions and transaction_logs.
func SetTransactions(transactions []harmony.Transaction) (err error) {
	data := struct {
		Profile      string
		Transactions []harmony.Transaction
	}{
		Profile:      profile,
		Transactions: transactions,
	}
	var buf bytes.Buffer
	t, err := template.New("fillTransaction").Parse(transactionsQ)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = t.Execute(&buf, data)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = RunTemplate(buf.String())
	return
}

// SetTokens takes a list of token.Token and saves those to the table tokens.
func SetTokens(tokens []harmony.Token) (err error) {
	var buf bytes.Buffer
	t, err := template.New("fillTokens").Parse(tokensQ)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = t.Execute(&buf, tokens)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = RunTemplate(buf.String())
	return
}

func SetMethods(methods []harmony.Method) (err error) {
	var buf bytes.Buffer
	t, err := template.New("fillMethods").Parse(methodsQ)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = t.Execute(&buf, methods)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = RunTemplate(buf.String())
	return
}

func SetSwaps(swaps []harmony.Swap) (err error) {
	data := struct {
		Profile string
		Swaps   []harmony.Swap
	}{
		Profile: profile,
		Swaps:   swaps,
	}
	var buf bytes.Buffer
	t, err := template.New("fillSwaps").Parse(swapsQ)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = t.Execute(&buf, data)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = RunTemplate(buf.String())
	return
}

func SetLiquidityActions(liquidityActions []harmony.LiquidityAction) (err error) {
	data := struct {
		Profile   string
		Liquidity []harmony.LiquidityAction
	}{
		Profile:   profile,
		Liquidity: liquidityActions,
	}
	var buf bytes.Buffer
	t, err := template.New("fillLiquidity").Parse(liquidityActionsQ)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = t.Execute(&buf, data)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = RunTemplate(buf.String())
	return
}

func SetLiquidityPools(liquidityPools []harmony.LiquidityPool) (err error) {
	var buf bytes.Buffer
	t, err := template.New("fillLiquidityPools").Parse(liquidityPoolsQ)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = t.Execute(&buf, liquidityPools)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = RunTemplate(buf.String())
	return
}

func SetLiquidityRatios(ratios []harmony.HistoricLiquidityRatio) (err error) {
	var buf bytes.Buffer
	t, err := template.New("fillLiquidityRatios").Parse(liquidityRatiosQ)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = t.Execute(&buf, ratios)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = RunTemplate(buf.String())
	return
}

func SetTokenTransfers(transfers []harmony.TokenTransaction) (err error) {
	data := struct {
		Profile   string
		Transfers []harmony.TokenTransaction
	}{
		Profile:   profile,
		Transfers: transfers,
	}
	var buf bytes.Buffer
	t, err := template.New("fillTokenTransfers").Parse(tokenTransfersQ)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = t.Execute(&buf, data)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = RunTemplate(buf.String())
	return
}

func UpdateSwapFees(swaps []harmony.Swap) (err error) {
	data := struct {
		Profile string
		Swaps   []harmony.Swap
	}{
		Profile: profile,
		Swaps:   swaps,
	}
	var buf bytes.Buffer
	t, err := template.New("updateSwapFees").Parse(swapFeesQ)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = t.Execute(&buf, data)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = RunTemplate(buf.String())
	return
}
