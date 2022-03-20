package main

import (
	"fmt"
	"github.com/mjmar01/harmolytics/internal/log"
	"github.com/mjmar01/harmolytics/internal/mysql"
	"github.com/mjmar01/harmolytics/pkg/harmony"
	addressPkg "github.com/mjmar01/harmolytics/pkg/harmony/address"
	"github.com/mjmar01/harmolytics/pkg/harmony/method"
	"github.com/mjmar01/harmolytics/pkg/harmony/token"
	"github.com/mjmar01/harmolytics/pkg/harmony/transaction"
	"github.com/mjmar01/harmolytics/pkg/harmony/uniswapV2"
	"github.com/mjmar01/harmolytics/pkg/rpc"
	"github.com/spf13/cobra"
	"sync"
)

const (
	transferQuery = "SELECT DISTINCT address FROM harmolytics_profile_%s.transaction_logs WHERE topics LIKE '%s%%'"
	methodsQuery  = "SELECT DISTINCT method_signature FROM harmolytics_profile_%s.transactions ORDER BY method_signature ASC"
)

const (
	transferEvent = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
)

var (
	address string
)

var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Set of commands to load information from the harmony blockchain",
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
		cmd.Help()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		rpc.CloseRpc()
	},
}

var loadTransactionsCmd = &cobra.Command{
	Use:   "transactions",
	Short: "Load all transactions for a given address or hash(s)",
	Run: func(cmd *cobra.Command, args []string) {
		log.Task("Loading transaction information from blockchain", log.InfoLevel)
		if address != "" {
			log.Debug("Getting transactions by wallet")
			w, err := addressPkg.New(address)
			log.CheckErr(err, log.PanicLevel)
			txs, err := transaction.GetTransactionsByWallet(w)
			log.CheckErr(err, log.PanicLevel)
			log.Info(fmt.Sprintf("Found %d transactions", len(txs)))
			err = mysql.SetTransactions(txs)
			log.CheckErr(err, log.PanicLevel)
		}
		if len(args) > 0 {
			log.Debug("Getting manually specified transactions")
			var txs []harmony.Transaction
			for _, arg := range args {
				tx, err := transaction.GetFullTransaction(arg)
				log.CheckErr(err, log.PanicLevel)
				txs = append(txs, tx)
			}
			err := mysql.SetTransactions(txs)
			log.CheckErr(err, log.PanicLevel)
		}
		log.Done()
	},
}

var loadTokensCmd = &cobra.Command{
	Use:   "tokens",
	Short: "Loads token information for all tokens that have been directly or indirectly interacted with",
	Run: func(cmd *cobra.Command, args []string) {
		log.Task("Loading token information from blockchain", log.InfoLevel)
		addrs, err := mysql.GetStringsByQuery(fmt.Sprintf(transferQuery, config.DB.Profile, transferEvent))
		log.CheckErr(err, log.PanicLevel)
		tokens, err := token.GetTokens(addrs)
		log.Info(fmt.Sprintf("Found %d distinct tokens", len(tokens)))
		err = mysql.SetTokens(tokens)
		log.CheckErr(err, log.PanicLevel)
		log.Done()
	},
}

var loadMethodsCmd = &cobra.Command{
	Use:   "methods",
	Short: "Look up method signatures of loaded transactions",
	Run: func(cmd *cobra.Command, args []string) {
		log.Task("Looking up method signatures in the 4byte.directory", log.InfoLevel)
		signatures, err := mysql.GetStringsByQuery(fmt.Sprintf(methodsQuery, config.DB.Profile))
		log.CheckErr(err, log.PanicLevel)
		if signatures[0] == "" {
			signatures = signatures[1:]
		}
		log.Info(fmt.Sprintf("Found %d unique method signature", len(signatures)))
		wg := sync.WaitGroup{}
		wg.Add(len(signatures))
		ch := make(chan harmony.Method, len(signatures))
		for _, signature := range signatures {
			go func(sig string) {
				method, err := method.GetMethod(sig)
				log.CheckErr(err, log.PanicLevel)
				ch <- method
				wg.Done()
			}(signature)
		}
		wg.Wait()
		var methods []harmony.Method
		for i := 0; i < len(signatures); i++ {
			m := <-ch
			if m.Signature != "" {
				methods = append(methods, m)
			}
		}
		diff := len(signatures) - len(methods)
		if diff > 0 {
			log.Warn(fmt.Sprintf("Unable to resolve %d method signatures", diff))
		}
		err = mysql.SetMethods(methods)
		log.CheckErr(err, log.PanicLevel)
		log.Done()
	},
}

var loadLiquidityRatiosCmd = &cobra.Command{
	Use:   "ratios",
	Short: "Loads relevant historic liquidity ratios during swaps",
	Run: func(cmd *cobra.Command, args []string) {
		log.Task("Loading historic liquidity ratios from blockchain", log.InfoLevel)
		txs, err := mysql.GetTransactionsByMethodName("swap%")
		log.CheckErr(err, log.PanicLevel)
		var lookups []harmony.HistoricLiquidityRatio
		for _, tx := range txs {
			s, err := uniswapV2.DecodeSwap(tx)
			log.CheckErr(err, log.PanicLevel)
			for _, pool := range s.Path {
				lookups = append(lookups, harmony.HistoricLiquidityRatio{LP: pool, BlockNum: tx.BlockNum - 1})
			}
		}
		log.Info(fmt.Sprintf("Determined %d relevant ratios that need to be fetched", len(lookups)))
		ratios, err := uniswapV2.GetLiquidityRatios(lookups)
		log.CheckErr(err, log.PanicLevel)
		err = mysql.SetLiquidityRatios(ratios)
		log.CheckErr(err, log.PanicLevel)
		log.Done()
	},
}

func init() {
	loadTransactionsCmd.PersistentFlags().StringVarP(&address, "address", "a", "", "")
	loadCmd.PersistentFlags().StringVar(&config.DB.Host, HostParam, "127.0.0.1", "")
	loadCmd.PersistentFlags().StringVar(&config.DB.Port, PortParam, "3306", "")
	loadCmd.PersistentFlags().StringVar(&config.DB.User, UserParam, "", "")
	loadCmd.PersistentFlags().StringVarP(&config.DB.Profile, ProfileParam, "p", "", "")
	loadCmd.AddCommand(loadMethodsCmd)
	loadCmd.AddCommand(loadTransactionsCmd)
	loadCmd.AddCommand(loadTokensCmd)
	loadCmd.AddCommand(loadLiquidityRatiosCmd)
	rootCmd.AddCommand(loadCmd)
}
