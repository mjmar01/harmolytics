// Package hmysolidityio is used in conjunction with the harmony package to process transaction inputs/outputs
package hmysolidityio

import (
	"encoding/json"
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/pkg/types"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

const (
	dictionaryUrl = "https://www.4byte.directory/api/v1/signatures/?hex_signature=0x"
)

func GetMethod(sig string) (m types.Method, err error) {
	// Get method information from dictionary
	resp, err := http.Get(dictionaryUrl + sig)
	if err != nil {
		return types.Method{}, errors.Wrap(err, 0)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return types.Method{}, errors.Wrap(err, 0)
	}
	err = resp.Body.Close()
	if err != nil {
		return types.Method{}, errors.Wrap(err, 0)
	}
	var data struct {
		Results []struct {
			TextSignature string `json:"text_signature"`
			ID            int    `json:"id"`
		} `json:"results"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return types.Method{}, errors.Wrap(err, 0)
	}
	// If results were found parse information
	if len(data.Results) > 0 {
		// Sort to get most likely match
		sort.Slice(data.Results, func(i, j int) bool {
			return data.Results[i].ID < data.Results[j].ID
		})
		// Some string cutting and fill method
		split := strings.IndexRune(data.Results[0].TextSignature, '(')
		m = types.Method{
			Signature:  sig,
			Name:       data.Results[0].TextSignature[:split],
			Parameters: strings.Split(strings.Trim(data.Results[0].TextSignature[split:], "()"), ","),
		}
	}
	return
}

func cleanInput(in string) (out string, err error) {
	out = strings.TrimPrefix(in, "0x")
	if len(out)%64 != 0 {
		out = out[8:]
	}
	if len(out)%64 != 0 {
		return "", errors.Errorf("input does not match expected pattern: (0x){0,1}(MethodSig){0,1}(32byteWord)+")
	}
	return
}
