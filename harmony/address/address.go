// Package address stores and converts harmony address information
package address

import (
	"encoding/hex"
	"github.com/btcsuite/btcutil/bech32"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"harmolytics/harmony"
	"strings"
)

// New reads the addr in 0x... or one1... format and returns an Address
func New(addr string) (a harmony.Address, err error) {
	if strings.HasPrefix(addr, "0x") {
		a.OneAddress, err = ethToOne(addr)
		if err != nil {
			return
		}
		a.EthAddress = ethCommon.HexToAddress(addr)
	} else if strings.HasPrefix(addr, "one1") {
		ethAddr, err := oneToEth(addr)
		if err != nil {
			return harmony.Address{}, err
		}
		a.OneAddress = addr
		a.EthAddress = ethCommon.HexToAddress(ethAddr)
	}
	return
}

func oneToEth(oneAddr string) (str string, err error) {
	_, data, err := bech32.Decode(oneAddr)
	if err != nil {
		return "", errors.Wrap(err, 0)
	}
	data, err = bech32.ConvertBits(data, 5, 8, true)
	if err != nil {
		return "", errors.Wrap(err, 0)
	}
	str = hex.EncodeToString(data)
	str = "0x" + str
	return
}

func ethToOne(hexAddr string) (str string, err error) {
	hexAddr = strings.TrimPrefix(hexAddr, "0x")
	addr, err := hex.DecodeString(hexAddr)
	if err != nil {
		return
	}
	conv, err := bech32.ConvertBits(addr, 8, 5, true)
	if err != nil {
		return "", errors.Wrap(err, 0)
	}
	str, err = bech32.Encode("one", conv)
	if err != nil {
		return "", errors.Wrap(err, 0)
	}
	return
}
