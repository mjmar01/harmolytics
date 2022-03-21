package main

import (
	"fmt"
	"github.com/mjmar01/harmolytics/pkg/harmony/address"
	"github.com/spf13/cobra"
	"os"
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "convert a given wallet address into the other format (one1/0x)",
	Run: func(cmd *cobra.Command, args []string) {
		for _, arg := range args {
			addr, err := address.New(arg)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			fmt.Printf("%s %s\n", addr.OneAddress, addr.EthAddress.String())
		}
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
}
