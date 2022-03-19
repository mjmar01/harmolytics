package main

import (
	"github.com/mjmar01/harmolytics/internal/log"
	"github.com/mjmar01/harmolytics/internal/mysql"
	"github.com/mjmar01/harmolytics/pkg/harmony"
	"github.com/mjmar01/harmolytics/pkg/harmony/uniswapV2"
	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Set of commands to run analysis on collected data",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := initViper()
		log.CheckErr(err, log.PanicLevel)
		err = initFlags(cmd)
		log.CheckErr(err, log.PanicLevel)
		err = initConfigVars()
		log.CheckErr(err, log.PanicLevel)
		log.SetLogLevel(config.LogLevel)
		err = mysql.ConnectDatabase(config.DB.User, config.DB.Host, config.DB.Port, config.DB.Profile)
		log.CheckErr(err, log.PanicLevel)
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var analyzeFeesCmd = &cobra.Command{
	Use:   "fees",
	Short: "analyze how much has been to lost to fees in swaps",
	Run: func(cmd *cobra.Command, args []string) {
		swaps, err := mysql.GetSwaps()
		log.CheckErr(err, log.PanicLevel)
		ratios, err := mysql.GetRatios()
		log.CheckErr(err, log.PanicLevel)
		for i := 0; i < len(swaps); i++ {
			block, err := mysql.GetBlockByTx(swaps[i].TxHash)
			log.CheckErr(err, log.PanicLevel)
			tx, err := mysql.GetTransactionByHash(swaps[i].TxHash)
			log.CheckErr(err, log.PanicLevel)
			if err != nil {
				return
			}
			var pathRatios []harmony.HistoricLiquidityRatio
			for _, pool := range swaps[i].Path {
				for _, ratio := range ratios {
					if ratio.LP.LpToken.Address.OneAddress == pool.LpToken.Address.OneAddress && ratio.BlockNum == block-1 {
						pathRatios = append(pathRatios, ratio)
						break
					}
				}
			}
			err = uniswapV2.AnalyzeFees(&swaps[i], pathRatios, tx)
			log.CheckErr(err, log.PanicLevel)
		}
		err = mysql.UpdateSwapFees(swaps)
		log.CheckErr(err, log.PanicLevel)
	},
}

func init() {
	analyzeCmd.AddCommand(analyzeFeesCmd)
	analyzeCmd.PersistentFlags().StringVar(&config.DB.Host, HostParam, "127.0.0.1", "")
	analyzeCmd.PersistentFlags().StringVar(&config.DB.Port, PortParam, "3306", "")
	analyzeCmd.PersistentFlags().StringVar(&config.DB.User, UserParam, "", "")
	analyzeCmd.PersistentFlags().StringVarP(&config.DB.Profile, ProfileParam, "p", "", "")
	rootCmd.AddCommand(analyzeCmd)
}
