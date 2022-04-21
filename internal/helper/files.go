package helper

import (
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

func CacheDir() (dir string) {
	dir, err := os.UserCacheDir()
	cobra.CheckErr(err)
	dir = filepath.Join(dir, "harmony-tk")
	err = os.MkdirAll(dir, 0750)
	cobra.CheckErr(err)
	return
}

func ConfigDir() (dir string) {
	dir, err := os.UserConfigDir()
	cobra.CheckErr(err)
	dir = filepath.Join(dir, "harmony-tk")
	err = os.MkdirAll(dir, 0750)
	cobra.CheckErr(err)
	return
}
