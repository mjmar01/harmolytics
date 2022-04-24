package cmd

import (
	"github.com/mjmar01/harmolytics/internal/helper"
	"github.com/mjmar01/harmolytics/pkg/cache"
	"github.com/spf13/cobra"
)

var cachePath string

var centralCache *cache.Cache

// cacheCmd represents the cache command
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Interact with and managed the local cache",
}

func openCache(cmd *cobra.Command, args []string) error {
	c, err := cache.NewCache(&cache.Opts{
		CacheDir:            cachePath,
		PreLoadTransactions: false,
	})
	if err != nil {
		return err
	}
	centralCache = c
	return nil
}

func init() {
	rootCmd.AddCommand(cacheCmd)

	cacheCmd.PersistentFlags().StringVarP(&cachePath, "path", "p", helper.CacheDir(), "path to the cache dir")
}
