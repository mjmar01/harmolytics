package hmysolidityio

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/pkg/harmony"
	"math/big"
	"strings"
)

// DecodeInt returns the contained uint256 as a *big.Int given the entire data input and position of the value.
// The position usually corresponds to the parameter position of the function call.
func DecodeInt(data string, intPosition int) (n *big.Int, err error) {
	intPosition *= 64
	data, err = cleanInput(data)
	if err != nil {
		return
	}
	bytes, err := hex.DecodeString(data[intPosition : intPosition+64])
	if err != nil {
		return
	}
	n = new(big.Int)
	n.SetBytes(bytes)
	return
}

// DecodeArray returns the contained array given the entire data input and position of the array.
// The array values are returned as a slice of bytes.
// The position usually corresponds to the parameter position of the function call.
func DecodeArray(data string, arrayPosition int) (arr [][]byte, err error) {
	arrayPosition *= 64
	data, err = cleanInput(data)
	if err != nil {
		return
	}
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
	data, err = cleanInput(data)
	if err != nil {
		return
	}
	a, err = harmony.CheckNewAddress("0x" + data[addressPosition+24:addressPosition+64])
	return
}

// DecodeString returns the contained string given the entire data input and position of the string.
// The position usually corresponds to the parameter position of the function call.
func DecodeString(data string, stringPosition int) (s string, err error) {
	stringPosition *= 64
	data, err = cleanInput(data)
	if err != nil {
		return
	}
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
