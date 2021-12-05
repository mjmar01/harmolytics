package cmd

import (
	"fmt"
	"github.com/mjmar01/cli-log/log"
	"github.com/spf13/cobra"
	"harmolytics/harmony"
	addressPkg "harmolytics/harmony/address"
	"harmolytics/harmony/method"
	"harmolytics/harmony/rpc"
	"harmolytics/harmony/token"
	"harmolytics/harmony/transaction"
	"harmolytics/mysql"
	"sync"
)

const (
	transferQuery = "SELECT DISTINCT address FROM harmolytics_%s.transaction_logs WHERE topics LIKE '%s%%'"
	methodsQuery  = "SELECT DISTINCT method FROM harmolytics_%s.transactions ORDER BY method ASC"
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
		cmd.Help()
	},
}

var loadTransactionsCmd = &cobra.Command{
	Use:   "transactions",
	Short: "Load all transactions for a given address or hash(s)",
	Run: func(cmd *cobra.Command, args []string) {
		if address != "" {
			w, err := addressPkg.New(address)
			log.CheckErr(err, log.FatalLevel)
			txs, err := transaction.GetTransactionsByWallet(w)
			log.CheckErr(err, log.FatalLevel)
			err = mysql.SetTransactions(txs)
			log.CheckErr(err, log.FatalLevel)
		}
		if len(args) > 0 {
			var txs []harmony.Transaction
			for _, arg := range args {
				tx, err := transaction.GetFullTransaction(arg)
				log.CheckErr(err, log.FatalLevel)
				txs = append(txs, tx)
			}
			err := mysql.SetTransactions(txs)
			log.CheckErr(err, log.FatalLevel)
		}
	},
}

var loadTokensCmd = &cobra.Command{
	Use:   "tokens",
	Short: "Loads token information for all tokens that have been directly or indirectly interacted with",
	Run: func(cmd *cobra.Command, args []string) {
		log.Task("Loading token information", log.InfoLevel)
		addrs, err := mysql.GetStringsByQuery(fmt.Sprintf(transferQuery, config.DB.Profile, transferEvent))
		log.CheckErr(err, log.FatalLevel)
		var tokens []harmony.Token
		wg := sync.WaitGroup{}
		wg.Add(len(addrs))
		ch := make(chan harmony.Token, len(addrs))
		for _, addr := range addrs {
			go func(addr string) {
				a, err := addressPkg.New(addr)
				log.CheckErr(err, log.FatalLevel)
				token, err := token.GetToken(a)
				ch <- token
				log.CheckErr(err, log.FatalLevel)
			}(addr)
			wg.Done()
		}
		wg.Wait()
		for i := 0; i < len(addrs); i++ {
			tk := <-ch
			if tk.Name != "" {
				tokens = append(tokens, tk)
			}
		}
		log.Info(fmt.Sprintf("Found %d distinct tokens", len(tokens)))
		log.Done()
		err = mysql.SetTokens(tokens)
		log.CheckErr(err, log.FatalLevel)
	},
}

var loadMethodsCmd = &cobra.Command{
	Use:   "methods",
	Short: "Look up method signatures of loaded transactions",
	Run: func(cmd *cobra.Command, args []string) {
		log.Task("Looking up method signatures", log.InfoLevel)
		signatures, err := mysql.GetStringsByQuery(fmt.Sprintf(methodsQuery, config.DB.Profile))
		log.CheckErr(err, log.FatalLevel)
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
				log.CheckErr(err, log.FatalLevel)
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
		log.CheckErr(err, log.FatalLevel)
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
	rootCmd.AddCommand(loadCmd)
}
