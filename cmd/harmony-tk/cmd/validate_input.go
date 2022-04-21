package cmd

import (
	"encoding/hex"
	"fmt"
	"github.com/mjmar01/harmolytics/pkg/types"
	"strings"
)

func validateAddresses(addrs ...string) error {
	for _, addr := range addrs {
		if _, err := types.CheckNewAddress(addr); err != nil {
			return fmt.Errorf("string: '%s' is not a valid address", addr)
		}
	}
	return nil
}

func validateTxHash(hashes ...string) error {
	for _, hash := range hashes {
		if data, err := hex.DecodeString(strings.TrimPrefix(hash, "0x")); err != nil || len(data) != 32 {
			return fmt.Errorf("string: '%s' is not a valid transaction hash", hash)
		}
	}
	return nil
}
