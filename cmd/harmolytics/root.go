package main

import (
	_ "embed"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/internal/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	viperPkg "github.com/spf13/viper"
	"os"
)

const (
	defaultConfigName = ".harmolytics-config.yaml"
	defaultConfigType = "yaml"
	defaultConfigPath = "."
)

//go:embed config/defaults.yaml
var defaultConfig []byte

// package vars
var (
	viper   *viperPkg.Viper
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "harmolytics",
	Short: "An analytics tool for harmony wallets",
	Long: `harmolytics is a CLI tool to retrieve and analyze harmony transactions.
You will need to supply your own MySQL DB as backend`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := initViper()
		log.CheckErr(err, log.PanicLevel)
		err = initFlags(cmd)
		log.CheckErr(err, log.PanicLevel)
		err = initConfigVars()
		log.CheckErr(err, log.PanicLevel)
		log.SetLogLevel(config.LogLevel)
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func initViper() (err error) {
	viper = viperPkg.New()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(defaultConfigPath)
		viper.SetConfigType(defaultConfigType)
		viper.SetConfigName(defaultConfigName)
	}

	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viperPkg.ConfigFileNotFoundError); ok {
			log.Warn("Config File not found at " + defaultConfigPath + "/" + defaultConfigName)
			log.Warn("Creating config with default values at " + defaultConfigPath + "/" + defaultConfigName)
			err = os.WriteFile(defaultConfigName, defaultConfig, 0660)
			if err != nil {
				return errors.Wrap(err, 0)
			}
			log.Warn("Continuing with default config\n")
			err = viper.ReadInConfig()
			if err != nil {
				return errors.Wrap(err, 0)
			}
		} else {
			if err != nil {
				return errors.Wrap(err, 0)
			}
		}
	}
	return
}

func initFlags(cmd *cobra.Command) (err error) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed && viper.IsSet(f.Name) {
			val := viper.Get(f.Name)
			err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			if err != nil {
				return
			}
		}
	})
	return
}

func initConfigVars() (err error) {
	config.RpcUrl = viper.GetString(RpcUrlParam)
	config.HistoricRpcUrl = viper.GetString(HistoricRpcUrlParam)
	return
}

// Execute executes the root command.
func Execute() {
	err := rootCmd.Execute()
	log.CheckErr(err, log.PanicLevel)
	fmt.Print("\n")
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Path to config file")
	rootCmd.PersistentFlags().IntVarP(&config.LogLevel, LogLevelParam, "v", 1, "Lowest log level to be written to console. Range -1(TRACE) to 6(NONE)")
}

func main() {
	Execute()
}
