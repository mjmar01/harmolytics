package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/mjmar01/harmolytics/internal/log"
	"github.com/mjmar01/harmolytics/internal/mysql"
	"github.com/spf13/cobra"
)

const (
	overwriteParam = "overwrite"
	wipeParam      = "wipe"
)

var (
	overwriteProfile bool
	wipeSchema       bool
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
		err = mysql.ConnectDatabase(config.DB.User, config.DB.Host, config.DB.Port, config.DB.Profile)
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
		err := mysql.InitSchema(overwriteProfile, wipeSchema)
		log.CheckErr(err, log.PanicLevel)
	},
}

func init() {
	sqlInitCmd.PersistentFlags().BoolVarP(&overwriteProfile, overwriteParam, "o", false, "Allow overwrite of previous profile")
	sqlInitCmd.PersistentFlags().BoolVar(&wipeSchema, wipeParam, false, "Wipes everything before initializing schema")
	sqlCmd.PersistentFlags().StringVar(&config.DB.Host, HostParam, "127.0.0.1", "")
	sqlCmd.PersistentFlags().StringVar(&config.DB.Port, PortParam, "3306", "")
	sqlCmd.PersistentFlags().StringVar(&config.DB.User, UserParam, "", "")
	sqlCmd.PersistentFlags().StringVarP(&config.DB.Profile, ProfileParam, "p", "default", "analytics profile to identify wallet")
	sqlCmd.AddCommand(sqlInitCmd)
	rootCmd.AddCommand(sqlCmd)
}