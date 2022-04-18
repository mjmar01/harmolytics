package harmony

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/go-errors/errors"
	"strings"
)

// NewAddress reads the addr in 0x... or one1... format and returns an Address.
// Panics if the provided input is invalid (use CheckNew for unsafe inputs).
func NewAddress(addr string) (a Address) {
	var err error
	if strings.HasPrefix(addr, "0x") {
		a.HexAddress = addr
		a.OneAddress, err = hexToOne(addr)
		if err != nil {
			panic(err)
		}
	} else if strings.HasPrefix(addr, "one1") {
		a.OneAddress = addr
		a.HexAddress, err = oneToHex(addr)
		if err != nil {
			panic(err)
		}
	}
	return
}

// CheckNewAddress reads the addr in 0x... or one1... format and returns an Address.
func CheckNewAddress(addr string) (a Address, err error) {
	if strings.HasPrefix(addr, "0x") {
		a.HexAddress = addr
		a.OneAddress, err = hexToOne(addr)
		if err != nil {
			return
		}
	} else if strings.HasPrefix(addr, "one1") {
		a.OneAddress = addr
		a.HexAddress, err = oneToHex(addr)
		if err != nil {
			return
		}
	}
	return
}

func oneToHex(oneAddr string) (str string, err error) {
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

func hexToOne(hexAddr string) (str string, err error) {
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
