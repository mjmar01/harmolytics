package main

import (
	_ "embed"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "harmony-tool",
	Short: "A collection of semi useful tools for harmony",
	Long:  `harmony-tool is a simple CLI tool to implement various features related to working with harmony`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute executes the root command.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func init() {

}

func main() {
	Execute()
}
