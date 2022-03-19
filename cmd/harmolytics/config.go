package main

import (
	_ "embed"
	"fmt"
	"github.com/mjmar01/harmolytics/internal/helper"
	"github.com/mjmar01/harmolytics/internal/log"
	"github.com/mjmar01/harmolytics/internal/mysql"
	"github.com/spf13/cobra"
	viperPkg "github.com/spf13/viper"
	"os"
	"strings"
	"text/template"
)

//go:embed config/config.tmpl
var configTemplate string

const (
	LogLevelParam       = "verbosity"
	HostParam           = "db-host"
	PortParam           = "db-port"
	UserParam           = "db-user"
	PasswordParam       = "db-password"
	ProfileParam        = "profile"
	RpcUrlParam         = "rpc-url"
	HistoricRpcUrlParam = "historic-rpc-url"
	KnownAddressesParam = "known-addresses"
)

type DbSubConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Profile  string
}

var config = struct {
	LogLevel       int
	DB             DbSubConfig
	RpcUrl         string
	HistoricRpcUrl string
}{}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Interact with the config file",
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
		err := cmd.Help()
		log.CheckErr(err, log.PanicLevel)
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Lists currently set config variables",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Config file in use: %s\n", viper.ConfigFileUsed())
		fmt.Print("Flag default values:\n")
		fmt.Printf(" %s:  %d\n", LogLevelParam, viper.GetInt(LogLevelParam))
		fmt.Printf(" %s:    %s\n", ProfileParam, viper.GetString(ProfileParam))
		fmt.Printf(" %s:    %s\n", HostParam, viper.GetString(HostParam))
		fmt.Printf(" %s:    %s\n", PortParam, viper.GetString(PortParam))
		fmt.Printf(" %s:    %s\n", UserParam, viper.GetString(UserParam))
		fmt.Print("Config Variables:\n")
		fmt.Printf(" %s: %s\n", RpcUrlParam, viper.GetString(RpcUrlParam))
		fmt.Printf(" %s: %s\n", HistoricRpcUrlParam, viper.GetString(HistoricRpcUrlParam))
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set values in the config",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			keys := viper.AllKeys()
			for _, arg := range args {
				a := strings.Split(arg, ":")
				key, value := a[0], a[1]
				if helper.StringInSlice(key, keys) {
					viper.Set(key, value)
				} else {
					log.Error(fmt.Sprintf("Key %s not recognized as parameter", key))
				}
			}
			writeConfig()
		} else {
			err := cmd.Help()
			log.CheckErr(err, log.PanicLevel)
		}
	},
}

var configDatabaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Configures database connection and saves it to config",
	Run: func(cmd *cobra.Command, args []string) {
		dbPassword, err := mysql.ConnectDatabase(config.DB.User, "", config.DB.Host, config.DB.Port, config.DB.Profile, cryptKey)
		log.CheckErr(err, log.PanicLevel)
		viperPkg.Set(PasswordParam, dbPassword)
		viper.Set(UserParam, config.DB.User)
		viper.Set(HostParam, config.DB.Host)
		viper.Set(PortParam, config.DB.Port)
		writeConfig()
	},
}

func writeConfig() {
	log.Task("Rewriting config file", log.DebugLevel)
	loadViperValues()
	t, err := template.New("config").Parse(configTemplate)
	log.CheckErr(err, log.PanicLevel)
	f, err := os.OpenFile(viper.ConfigFileUsed(), os.O_WRONLY, 0660)
	log.CheckErr(err, log.PanicLevel)
	err = f.Truncate(0)
	log.CheckErr(err, log.PanicLevel)
	_, err = f.Seek(0, 0)
	log.CheckErr(err, log.PanicLevel)
	err = t.Execute(f, config)
	log.CheckErr(err, log.PanicLevel)
	log.Done()
}

func loadViperValues() {
	config.LogLevel = viper.GetInt(LogLevelParam)
	config.RpcUrl = viper.GetString(RpcUrlParam)
	config.HistoricRpcUrl = viper.GetString(HistoricRpcUrlParam)
	config.DB.Profile = viper.GetString(ProfileParam)
	config.DB.Host = viper.GetString(HostParam)
	config.DB.Port = viper.GetString(PortParam)
	config.DB.User = viper.GetString(UserParam)
	config.DB.Password = viper.GetString(PasswordParam)
}

func init() {
	configDatabaseCmd.PersistentFlags().StringVar(&config.DB.Host, HostParam, "127.0.0.1", "")
	configDatabaseCmd.PersistentFlags().StringVar(&config.DB.Port, PortParam, "3306", "")
	configDatabaseCmd.PersistentFlags().StringVar(&config.DB.User, UserParam, "", "")
	configCmd.AddCommand(configDatabaseCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	rootCmd.AddCommand(configCmd)
}
