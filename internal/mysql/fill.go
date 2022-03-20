package mysql

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/internal/log"
	"github.com/mjmar01/harmolytics/pkg/harmony"
	"strings"
	"text/template"
)

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

//go:embed queries/fill_fees.tmpl
var swapFeesQ string

func runTemplate(queries string) (err error) {
	log.Trace("Running SQL template")
	for _, query := range strings.Split(strings.TrimRight(queries, " ;\n\t"), ";") {
		if query == "" {
			continue
		}
		rows, err := db.Query(query)
		if err != nil {
			fmt.Println(query)
			return errors.Wrap(err, 0)
		}
		err = rows.Close()
		if err != nil {
			return errors.Wrap(err, 0)
		}
	}
	return
}

func SetMethods(methods []harmony.Method) (err error) {
	log.Task("Saving methods to database", log.InfoLevel)
	var buf bytes.Buffer
	t, err := template.New("fillMethods").Parse(methodsQ)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = t.Execute(&buf, methods)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = runTemplate(buf.String())
	log.Done()
	return
}

func SetSwaps(swaps []harmony.Swap) (err error) {
	log.Task("Saving swaps to database", log.InfoLevel)
	data := struct {
		Profile string
		Swaps   []harmony.Swap
	}{
		Profile: prfl,
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
	err = runTemplate(buf.String())
	log.Done()
	return
}

func SetLiquidityActions(liquidityActions []harmony.LiquidityAction) (err error) {
	log.Task("Saving liquidity actions to database", log.InfoLevel)
	data := struct {
		Profile   string
		Liquidity []harmony.LiquidityAction
	}{
		Profile:   prfl,
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
	err = runTemplate(buf.String())
	log.Done()
	return
}

func SetLiquidityPools(liquidityPools []harmony.LiquidityPool) (err error) {
	log.Task("Saving liquidity pools to database", log.InfoLevel)
	var buf bytes.Buffer
	t, err := template.New("fillLiquidityPools").Parse(liquidityPoolsQ)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = t.Execute(&buf, liquidityPools)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = runTemplate(buf.String())
	log.Done()
	return
}

func SetLiquidityRatios(ratios []harmony.HistoricLiquidityRatio) (err error) {
	log.Task("Saving liquidity ratios to database", log.InfoLevel)
	var buf bytes.Buffer
	t, err := template.New("fillLiquidityRatios").Parse(liquidityRatiosQ)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = t.Execute(&buf, ratios)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = runTemplate(buf.String())
	log.Done()
	return
}

func UpdateSwapFees(swaps []harmony.Swap) (err error) {
	log.Task("Adding fees to swap entries in database", log.InfoLevel)
	data := struct {
		Profile string
		Swaps   []harmony.Swap
	}{
		Profile: prfl,
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
	err = runTemplate(buf.String())
	log.Done()
	return
}
