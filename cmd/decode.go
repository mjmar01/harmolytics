package cmd

import (
	"fmt"
	"github.com/mjmar01/cli-log/log"
	"github.com/spf13/cobra"
	"harmolytics/harmony"
	"harmolytics/harmony/rpc"
	"harmolytics/harmony/transaction"
	"harmolytics/harmony/uniswapV2"
	"harmolytics/mysql"
)

var decodeCmd = &cobra.Command{
	Use:   "decode",
	Short: "Analyzes transactions to extract useful information",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := initViper()
		log.CheckErr(err, log.FatalLevel)
		err = initFlags(cmd)
		log.CheckErr(err, log.FatalLevel)
		err = initConfigVars()
		log.CheckErr(err, log.FatalLevel)
		log.SetLogLevel(config.LogLevel)
		rpc.SetRpcUrl(config.RpcUrl)
		_, err = mysql.ConnectDatabase(config.DB.User, config.DB.Password, config.DB.Host, config.DB.Port, config.DB.Profile, cryptKey)
		log.CheckErr(err, log.FatalLevel)
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		log.CheckErr(err, log.WarnLevel)
	},
}

var decodeSwapsCmd = &cobra.Command{
	Use:   "swaps",
	Short: "Analyzes swaps to get input and output token amounts",
	Run: func(cmd *cobra.Command, args []string) {
		log.Task("Analyzing swaps", log.InfoLevel)
		txs, err := mysql.GetTransactionsByMethodName("swap%")
		log.CheckErr(err, log.FatalLevel)
		log.Info(fmt.Sprintf("Found %d swaps", len(txs)))
		var swaps []harmony.Swap
		for _, tx := range txs {
			s, err := uniswapV2.DecodeSwap(tx)
			log.CheckErr(err, log.FatalLevel)
			swaps = append(swaps, s)
		}
		log.Done()
		log.Task("Saving swaps to database", log.InfoLevel)
		err = mysql.SetSwaps(swaps)
		log.CheckErr(err, log.FatalLevel)
		log.Done()
	},
}

var decodeLiquidityCmd = &cobra.Command{
	Use:   "liquidity",
	Short: "Analyzes uniswap LP interactions",
	Run: func(cmd *cobra.Command, args []string) {
		log.Task("Analyzing liquidity", log.InfoLevel)
		adds, err := mysql.GetTransactionsByMethodName("addLiquidity%")
		log.CheckErr(err, log.FatalLevel)
		rems, err := mysql.GetTransactionsByMethodName("removeLiquidity%")
		log.CheckErr(err, log.FatalLevel)
		txs := append(adds, rems...)
		log.Info(fmt.Sprintf("Found %d LP interactions", len(txs)))
		var lpActions []harmony.LiquidityAction
		for _, tx := range txs {
			lpAction, err := uniswapV2.DecodeLiquidity(tx)
			log.CheckErr(err, log.FatalLevel)
			lpActions = append(lpActions, lpAction)
		}
		log.Done()
		log.Task("Saving liquidity actions to database", log.InfoLevel)
		err = mysql.SetLiquidity(lpActions)
		log.CheckErr(err, log.FatalLevel)
		log.Done()
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
			log.CheckErr(err, log.FatalLevel)
			transfers = append(transfers, tTx...)
		}
		log.Info(fmt.Sprintf("Found %d token transfers", len(transfers)))
		log.Done()
		log.Task("Saving token transfers to database", log.InfoLevel)
		err = mysql.SetTokenTransfers(transfers)
		log.CheckErr(err, log.FatalLevel)
		log.Done()
	},
}

func init() {
	decodeCmd.AddCommand(decodeSwapsCmd)
	decodeCmd.AddCommand(decodeLiquidityCmd)
	decodeCmd.AddCommand(decodeTransfersCmd)
	decodeCmd.PersistentFlags().StringVar(&config.DB.Host, HostParam, "127.0.0.1", "")
	decodeCmd.PersistentFlags().StringVar(&config.DB.Port, PortParam, "3306", "")
	decodeCmd.PersistentFlags().StringVar(&config.DB.User, UserParam, "", "")
	decodeCmd.PersistentFlags().StringVarP(&config.DB.Profile, ProfileParam, "p", "", "")
	rootCmd.AddCommand(decodeCmd)
}
