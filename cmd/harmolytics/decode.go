package main

import (
	"fmt"
	"github.com/mjmar01/harmolytics/internal/log"
	"github.com/mjmar01/harmolytics/internal/mysql"
	"github.com/mjmar01/harmolytics/pkg/harmony"
	"github.com/mjmar01/harmolytics/pkg/harmony/transaction"
	"github.com/mjmar01/harmolytics/pkg/harmony/uniswapV2"
	"github.com/mjmar01/harmolytics/pkg/rpc"
	"github.com/spf13/cobra"
)

var decodeCmd = &cobra.Command{
	Use:   "decode",
	Short: "Analyzes transactions to extract useful information",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := initViper()
		log.CheckErr(err, log.PanicLevel)
		err = initFlags(cmd)
		log.CheckErr(err, log.PanicLevel)
		err = initConfigVars()
		log.CheckErr(err, log.PanicLevel)
		log.SetLogLevel(config.LogLevel)
		err = rpc.InitRpc(config.RpcUrl, config.HistoricRpcUrl)
		log.CheckErr(err, log.PanicLevel)
		err = mysql.ConnectDatabase(config.DB.User, config.DB.Host, config.DB.Port, config.DB.Profile)
		log.CheckErr(err, log.PanicLevel)
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		log.CheckErr(err, log.WarnLevel)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		rpc.CloseRpc()
	},
}

var decodeSwapsCmd = &cobra.Command{
	Use:   "swaps",
	Short: "Analyzes swaps to get input and output token amounts",
	Run: func(cmd *cobra.Command, args []string) {
		log.Task("Analyzing swaps", log.InfoLevel)
		txs, err := mysql.GetTransactionsByMethodName("swap%")
		log.CheckErr(err, log.PanicLevel)
		log.Info(fmt.Sprintf("Found %d swaps", len(txs)))
		var swaps []harmony.Swap
		for _, tx := range txs {
			s, err := uniswapV2.DecodeSwap(tx)
			log.CheckErr(err, log.PanicLevel)
			swaps = append(swaps, s)
		}
		log.Done()
		log.Task("Saving swaps to database", log.InfoLevel)
		err = mysql.SetSwaps(swaps)
		log.CheckErr(err, log.PanicLevel)
		log.Done()
	},
}

var decodeLiquidityActionsCmd = &cobra.Command{
	Use:   "liquidity-actions",
	Short: "Analyzes uniswap LP interactions",
	Run: func(cmd *cobra.Command, args []string) {
		log.Task("Analyzing liquidity", log.InfoLevel)
		adds, err := mysql.GetTransactionsByMethodName("addLiquidity%")
		log.CheckErr(err, log.PanicLevel)
		rems, err := mysql.GetTransactionsByMethodName("removeLiquidity%")
		log.CheckErr(err, log.PanicLevel)
		txs := append(adds, rems...)
		log.Info(fmt.Sprintf("Found %d LP interactions", len(txs)))
		var lpActions []harmony.LiquidityAction
		for _, tx := range txs {
			lpAction, err := uniswapV2.DecodeLiquidityAction(tx)
			log.CheckErr(err, log.PanicLevel)
			lpActions = append(lpActions, lpAction)
		}
		log.Done()
		log.Task("Saving liquidity actions to database", log.InfoLevel)
		err = mysql.SetLiquidityActions(lpActions)
		log.CheckErr(err, log.PanicLevel)
		log.Done()
	},
}

var decodeLiquidityPoolsCmd = &cobra.Command{
	Use:   "liquidity-pools",
	Short: "Reads liquidity pool information from swaps",
	Run: func(cmd *cobra.Command, args []string) {
		txs, err := mysql.GetTransactionsByMethodName("swap%")
		log.CheckErr(err, log.PanicLevel)
		uniqueLPs := make(map[string]harmony.LiquidityPool)
		for _, tx := range txs {
			lps, err := uniswapV2.DecodeLiquidity(tx)
			log.CheckErr(err, log.PanicLevel)
			for _, lp := range lps {
				uniqueLPs[lp.LpToken.Address.OneAddress] = lp
			}
		}
		var lps []harmony.LiquidityPool
		for _, lp := range uniqueLPs {
			lps = append(lps, lp)
		}
		err = mysql.SetLiquidityPools(lps)
		log.CheckErr(err, log.PanicLevel)
	},
}

var decodeTransfersCmd = &cobra.Command{
	Use:   "transfers",
	Short: "Analyze normal and internal token transfers",
	Run: func(cmd *cobra.Command, args []string) {
		log.Task("Analyzing transfers", log.InfoLevel)
		txs, err := mysql.GetTransactionsByLogId(transferEvent)
		log.CheckErr(err, log.PanicLevel)
		log.Info(fmt.Sprintf("Found %d transactions containing token transfers", len(txs)))
		var transfers []harmony.TokenTransaction
		for _, tx := range txs {
			tTx, err := transaction.DecodeTokenTransaction(tx)
			log.CheckErr(err, log.PanicLevel)
			transfers = append(transfers, tTx...)
		}
		log.Info(fmt.Sprintf("Found %d token transfers", len(transfers)))
		log.Done()
		log.Task("Saving token transfers to database", log.InfoLevel)
		err = mysql.SetTokenTransfers(transfers)
		log.CheckErr(err, log.PanicLevel)
		log.Done()
	},
}

func init() {
	decodeCmd.AddCommand(decodeSwapsCmd)
	decodeCmd.AddCommand(decodeLiquidityActionsCmd)
	decodeCmd.AddCommand(decodeLiquidityPoolsCmd)
	decodeCmd.AddCommand(decodeTransfersCmd)
	decodeCmd.PersistentFlags().StringVar(&config.DB.Host, HostParam, "127.0.0.1", "")
	decodeCmd.PersistentFlags().StringVar(&config.DB.Port, PortParam, "3306", "")
	decodeCmd.PersistentFlags().StringVar(&config.DB.User, UserParam, "", "")
	decodeCmd.PersistentFlags().StringVarP(&config.DB.Profile, ProfileParam, "p", "", "")
	rootCmd.AddCommand(decodeCmd)
}
