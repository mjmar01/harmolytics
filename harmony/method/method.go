// Package method handles looking up method information by signature
package method

import (
	"encoding/json"
	"github.com/go-errors/errors"
	"harmolytics/harmony"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

const (
	dictionaryUrl = "https://www.4byte.directory/api/v1/signatures/?hex_signature=0x"
)

type methodJson struct {
	Results []struct {
		TextSignature string `json:"text_signature"`
		ID            int    `json:"id"`
	} `json:"results"`
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
	var data methodJson
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
