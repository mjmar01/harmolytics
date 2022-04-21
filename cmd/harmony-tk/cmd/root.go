package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "harmony-tk",
	Short: "Harmony Analytics Toolkit",
	Long: `------------------------------------------------------------------------------------------
| harmony-tk is a collection of tools to read and analyse transaction data.              |
|                                                                                        |
| Transaction data and similar is cached locally and almost all subcommands              |
| operate on this cache. Thus manually editing it is discouraged                         |
| and should be done with the cache sub command.                                         |
|                                                                                        |
| It's build with performance and RPC stability in mind.                                 |
| Some functions are already heavily optimized,                                          |
| as a result multiple running instances of harmony-tk cause slow downs. So:             |
| ###################################################################################### |
| DO NOT run this tool concurrently. This will NOT speed up loading times in most cases. |
| ###################################################################################### |
|                                                                                        |                                                                                                          
| Bug reports and feature requests here: https://github.com/mjmar01/harmolytics/issues   |
------------------------------------------------------------------------------------------

Example usage:
%s

Use --help or -h to see any commands available subcommands and options
`,
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := os.Executable()
		fmt.Printf(cmd.Long, filepath.Base(name))
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $XDG_CONFIG_HOME/harmony-tk/config.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// User default
		dir, err := os.UserConfigDir()
		cobra.CheckErr(err)

		configDir := filepath.Join(dir, "harmony-tk")
		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config.yaml")
	}

	// read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
