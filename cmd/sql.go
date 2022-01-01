package cmd

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/mjmar01/cli-log/log"
	"github.com/spf13/cobra"
	"harmolytics/mysql"
)

const (
	overwriteParam = "overwrite"
)

var (
	overwriteProfile bool
)

var sqlCmd = &cobra.Command{
	Use:   "sql",
	Short: "Set of commands to interact with the MySQL backend",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := initViper()
		log.CheckErr(err, log.PanicLevel)
		err = initFlags(cmd)
		log.CheckErr(err, log.PanicLevel)
		err = initConfigVars()
		log.CheckErr(err, log.PanicLevel)
		log.SetLogLevel(config.LogLevel)
		_, err = mysql.ConnectDatabase(config.DB.User, config.DB.Password, config.DB.Host, config.DB.Port, config.DB.Profile, cryptKey)
		log.CheckErr(err, log.PanicLevel)
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var sqlInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize database schema and necessary tables. This deletes profiles before recreating if they exist",
	Run: func(cmd *cobra.Command, args []string) {
		err := mysql.InitSchema(overwriteProfile)
		log.CheckErr(err, log.PanicLevel)
		err = mysql.AddKnown(config.KnownInfo)
		log.CheckErr(err, log.PanicLevel)
	},
}

func init() {
	sqlInitCmd.PersistentFlags().BoolVarP(&overwriteProfile, overwriteParam, "o", false, "Allow overwrite of previous profile")
	sqlCmd.PersistentFlags().StringVar(&config.DB.Host, HostParam, "127.0.0.1", "")
	sqlCmd.PersistentFlags().StringVar(&config.DB.Port, PortParam, "3306", "")
	sqlCmd.PersistentFlags().StringVar(&config.DB.User, UserParam, "", "")
	sqlCmd.PersistentFlags().StringVarP(&config.DB.Profile, ProfileParam, "p", "default", "analytics profile to identify wallet")
	sqlCmd.AddCommand(sqlInitCmd)
	rootCmd.AddCommand(sqlCmd)
}
