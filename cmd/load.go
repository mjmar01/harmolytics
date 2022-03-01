package cmd

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/mjmar01/cli-log/log"
	"github.com/spf13/cobra"
	"harmolytics/harmony"
	addressPkg "harmolytics/harmony/address"
	"harmolytics/harmony/method"
	"harmolytics/harmony/rpc"
	"harmolytics/harmony/token"
	"harmolytics/harmony/transaction"
	"harmolytics/harmony/uniswapV2"
	"harmolytics/mysql"
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
		rpc.InitRpc(config.RpcUrl, config.HistoricRpcUrl)
		_, err = mysql.ConnectDatabase(config.DB.User, config.DB.Password, config.DB.Host, config.DB.Port, config.DB.Profile, cryptKey)
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
		if address != "" {
			w, err := addressPkg.New(address)
			log.CheckErr(err, log.PanicLevel)
			txs, err := transaction.GetTransactionsByWallet(w)
			log.CheckErr(err, log.PanicLevel)
			err = mysql.SetTransactions(txs)
			log.CheckErr(err, log.PanicLevel)
		}
		if len(args) > 0 {
			var txs []harmony.Transaction
			for _, arg := range args {
				tx, err := transaction.GetFullTransaction(arg)
				log.CheckErr(err, log.PanicLevel)
				txs = append(txs, tx)
			}
			err := mysql.SetTransactions(txs)
			log.CheckErr(err, log.PanicLevel)
		}
	},
}

var loadTokensCmd = &cobra.Command{
	Use:   "tokens",
	Short: "Loads token information for all tokens that have been directly or indirectly interacted with",
	Run: func(cmd *cobra.Command, args []string) {
		log.Task("Loading token information", log.InfoLevel)
		addrs, err := mysql.GetStringsByQuery(fmt.Sprintf(transferQuery, config.DB.Profile, transferEvent))
		log.CheckErr(err, log.PanicLevel)
		tokens, err := token.GetTokens(addrs)
		log.Info(fmt.Sprintf("Found %d distinct tokens", len(tokens)))
		log.Done()
		err = mysql.SetTokens(tokens)
		log.CheckErr(err, log.PanicLevel)
	},
}

var loadMethodsCmd = &cobra.Command{
	Use:   "methods",
	Short: "Look up method signatures of loaded transactions",
	Run: func(cmd *cobra.Command, args []string) {
		log.Task("Looking up method signatures", log.InfoLevel)
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
		log.Done()
		err = mysql.SetMethods(methods)
		log.CheckErr(err, log.PanicLevel)
	},
}

var loadLiquidityRatiosCmd = &cobra.Command{
	Use:   "ratios",
	Short: "Loads relevant historic liquidity ratios during swaps",
	Run: func(cmd *cobra.Command, args []string) {
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
		var ratios []harmony.HistoricLiquidityRatio
		for _, lk := range lookups {
			lk, err = uniswapV2.GetLiquidityRatio(lk.LP, lk.BlockNum)
			if err != nil {
				fmt.Println(err.(*errors.Error).ErrorStack())
				panic(err)
			}
			ratios = append(ratios, lk)
		}
		err = mysql.SetLiquidityRatios(ratios)
		log.CheckErr(err, log.PanicLevel)
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
