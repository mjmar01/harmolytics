package hmysolidityio

import (
	"encoding/hex"
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/pkg/harmony"
	"math"
	"math/big"
	"sort"
)

type Encoder struct {
	in       []interface{}
	out      string
	offset   int
	carry    int
	prefixes []string
	suffixes map[int]string
}

func EncodeAll(in ...interface{}) (out string, err error) {
	e := &Encoder{
		in:       in,
		out:      "",
		offset:   0,
		carry:    0,
		prefixes: make([]string, len(in)),
		suffixes: map[int]string{},
	}
	err = checkInput(in)
	if err != nil {
		return
	}

	// First level
	next := map[int][]interface{}{}
	for i, value := range in {
		switch value.(type) {
		case *big.Int:
			e.prefixes[i] = encodeInt(value.(*big.Int))
		case harmony.Address:
			e.prefixes[i] = encodeAddress(value.(harmony.Address))
		case string:
			suffix := encodeString(value.(string))
			offset, prefix := e.getOffset(len(suffix) / 64)
			e.prefixes[i] = prefix
			e.suffixes[offset] = suffix
		case []interface{}:
			l := len(value.([]interface{}))
			offset, prefix := e.getOffset(l + 1)
			e.prefixes[i] = prefix
			next[offset] = value.([]interface{})
		}
	}

	// Next levels
	for offset, n := range next {
		e.recurseEncode(offset, n)
	}

	// Build out
	for _, prefix := range e.prefixes {
		e.out += prefix
	}
	var keys []int
	for k := range e.suffixes {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, key := range keys {
		e.out += e.suffixes[key]
	}
	return e.out, nil
}

func (e *Encoder) recurseEncode(offset int, in []interface{}) {
	n := new(big.Int)
	n.SetInt64(int64(len(in)))
	e.suffixes[offset] = encodeInt(n)
	next := map[int][]interface{}{}
	for _, value := range in {
		switch value.(type) {
		case *big.Int:
			e.suffixes[offset] += encodeInt(value.(*big.Int))
		case harmony.Address:
			e.suffixes[offset] += encodeAddress(value.(harmony.Address))
		case string:
			suffix := encodeString(value.(string))
			subOffset, prefix := e.getOffset(len(suffix) / 64)
			e.suffixes[offset] += prefix
			e.suffixes[subOffset] = suffix
		case []interface{}:
			l := len(value.([]interface{}))
			subOffset, prefix := e.getOffset(l + 1)
			e.suffixes[offset] += prefix
			next[subOffset] = value.([]interface{})
		}
	}
	for offset, n := range next {
		e.recurseEncode(offset, n)
	}
}

func checkInput(in []interface{}) (err error) {
	for _, i := range in {
		switch i.(type) {
		case *big.Int:
			if len(i.(*big.Int).Bytes()) > 32 {
				return errors.Errorf("Input integer exceeds maximum size of 256bit")
			}
		case string:
			continue
		case harmony.Address:
			continue
		case []interface{}:
			err = checkInput(i.([]interface{}))
			if err != nil {
				return
			}
		default:
			return errors.Errorf("Unsupported input type")
		}
	}
	return nil
}

func (e *Encoder) getOffset(l int) (offset int, s string) {
	offset = len(e.in)*32 + e.carry
	e.carry += l * 32
	b := new(big.Int)
	b.SetInt64(int64(offset))
	s = hex.EncodeToString(b.FillBytes(make([]byte, 32)))
	return
}

func encodeString(in string) (s string) {
	s = encodeInt(big.NewInt(int64(len(in))))
	size := int(math.Ceil(float64(len(in)) / 32))
	bytes := make([]byte, size*32)
	for i, c := range in {
		bytes[i] = byte(c)
	}
	s += hex.EncodeToString(bytes)
	return
}

func encodeInt(n *big.Int) (s string) {
	bytes := make([]byte, 32)
	bytes = n.FillBytes(bytes)
	s = hex.EncodeToString(bytes)
	return
}

func encodeAddress(a harmony.Address) (s string) {
	s = "000000000000000000000000" + a.HexAddress[2:]
	return
}
