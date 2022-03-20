// Package solidityio contains helper functions to extract information from hex formatted data used in solidity contract interactions
package solidityio

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/pkg/harmony"
	"github.com/mjmar01/harmolytics/pkg/harmony/address"
	"math"
	"math/big"
	"strings"
)

// DecodeString returns the contained string given the entire data input and position of the string.
// The position usually corresponds to the parameter position of the function call.
func DecodeString(data string, stringPosition int) (s string, err error) {
	stringPosition *= 64
	data = strings.TrimPrefix(data, "0x")
	offset, err := hexutil.DecodeUint64("0x" + strings.TrimLeft(data[stringPosition:stringPosition+64], "0"))
	if err != nil {
		return "", errors.Wrap(err, 0)
	}
	data = data[offset*2:]
	strLen, err := hexutil.DecodeUint64("0x" + strings.TrimLeft(data[:64], "0"))
	if err != nil {
		return "", errors.Wrap(err, 0)
	}
	bytes, err := hex.DecodeString(data[64 : 64+strLen*2])
	if err != nil {
		return
	}
	s = string(bytes)
	return
}

// EncodeString returns the string as an input. Prepend offset accordingly
func EncodeString(in string) (s string, err error) {
	s, err = EncodeInt(big.NewInt(int64(len(in))))
	if err != nil {
		return "", err
	}
	size := int(math.Ceil(float64(len(in)) / 32))
	bytes := make([]byte, size*32)
	for i, c := range in {
		bytes[i] = byte(c)
	}
	s += hex.EncodeToString(bytes)
	return
}

// DecodeInt returns the contained uint256 as a *big.Int given the entire data input and position of the value.
// The position usually corresponds to the parameter position of the function call.
func DecodeInt(data string, intPosition int) (n *big.Int, err error) {
	intPosition *= 64
	data = strings.TrimPrefix(data, "0x")
	bytes, err := hex.DecodeString(data[intPosition : intPosition+64])
	if err != nil {
		return
	}
	n = new(big.Int)
	n.SetBytes(bytes)
	return
}

// EncodeInt returns the 256bit string representation for use as input
func EncodeInt(n *big.Int) (s string, err error) {
	if len(n.Bytes()) > 32 {
		return "", errors.Errorf("Input exceeds maximum size of 256bit")
	}
	bytes := make([]byte, 32)
	bytes = n.FillBytes(bytes)
	s = hex.EncodeToString(bytes)
	return
}

// DecodeArray returns the contained array given the entire data input and position of the array.
// The array values are returned as a slice of bytes.
// The position usually corresponds to the parameter position of the function call.
func DecodeArray(data string, arrayPosition int) (arr [][]byte, err error) {
	arrayPosition *= 64
	data = strings.TrimPrefix(data, "0x")
	offset, err := hexutil.DecodeUint64("0x" + strings.TrimLeft(data[arrayPosition:arrayPosition+64], "0"))
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	data = data[offset*2:]
	arrLen64, err := hexutil.DecodeUint64("0x" + strings.TrimLeft(data[:64], "0"))
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	arrLen := int(arrLen64)
	for i := 0; i < arrLen; i++ {
		data = data[64:]
		data, err := hex.DecodeString(data[:64])
		if err != nil {
			return nil, err
		}
		arr = append(arr, data)
	}
	return
}

// DecodeAddress returns the contained harmony.Address given the entire data input and position of the address.
// The position usually corresponds to the parameter position of the function call.
func DecodeAddress(data string, addressPosition int) (a harmony.Address, err error) {
	addressPosition *= 64
	data = strings.TrimPrefix(data, "0x")
	a, err = address.New("0x" + data[addressPosition+24:addressPosition+64])
	if err != nil {
		return
	}
	return
}

// EncodeAddress returns the 256 bit representation of an address for use as input
func EncodeAddress(a harmony.Address) (s string) {
	s = hex.EncodeToString(append(make([]byte, 12), a.EthAddress.Bytes()...))
	return
}
