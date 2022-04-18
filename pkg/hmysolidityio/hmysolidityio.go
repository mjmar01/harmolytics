// Package hmysolidityio is used in conjunction with the harmony package to process transaction inputs/outputs
package hmysolidityio

import (
	"encoding/hex"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/pkg/harmony"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"sort"
	"strings"
)

const (
	dictionaryUrl = "https://www.4byte.directory/api/v1/signatures/?hex_signature=0x"
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
	a, err = harmony.CheckNewAddress("0x" + data[addressPosition+24:addressPosition+64])
	return
}

// EncodeAddress returns the 256 bit representation of an address for use as input
func EncodeAddress(a harmony.Address) (s string) {
	s = "0x000000000000000000000000" + a.HexAddress[2:]
	return
}

func GetMethod(sig string) (m harmony.Method, err error) {
	// Get method information from dictionary
	resp, err := http.Get(dictionaryUrl + sig)
	if err != nil {
		return harmony.Method{}, errors.Wrap(err, 0)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return harmony.Method{}, errors.Wrap(err, 0)
	}
	err = resp.Body.Close()
	if err != nil {
		return harmony.Method{}, errors.Wrap(err, 0)
	}
	var data struct {
		Results []struct {
			TextSignature string `json:"text_signature"`
			ID            int    `json:"id"`
		} `json:"results"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return harmony.Method{}, errors.Wrap(err, 0)
	}
	// If results were found parse information
	if len(data.Results) > 0 {
		// Sort to get most likely match
		sort.Slice(data.Results, func(i, j int) bool {
			return data.Results[i].ID < data.Results[j].ID
		})
		// Some string cutting and fill method
		split := strings.IndexRune(data.Results[0].TextSignature, '(')
		m = harmony.Method{
			Signature:  sig,
			Name:       data.Results[0].TextSignature[:split],
			Parameters: strings.Split(strings.Trim(data.Results[0].TextSignature[split:], "()"), ","),
		}
	}
	return
}
