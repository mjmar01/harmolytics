package cmd

import (
	"github.com/mjmar01/harmolytics/internal/helper"
	"github.com/spf13/cobra"
)

var cachePath string

// cacheCmd represents the cache command
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Interact with and managed the local cache",
}

func init() {
	rootCmd.AddCommand(cacheCmd)

	cacheCmd.PersistentFlags().StringVarP(&cachePath, "path", "p", helper.CacheDir(), "path to the cache dir")
}
