// Package rpc exports slightly adapted versions of harmony RPC endpoints as functions for convenience.
package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-errors/errors"
	"io/ioutil"
	"net/http"
)

var (
	rpcUrl         string
	historicRpcUrl string
	queryId        = 1
)

func SetRpcUrl(url, historicUrl string) {
	rpcUrl = url
	historicRpcUrl = historicUrl
}

func safeRpcCall(method string, params interface{}) (result interface{}, err error) {
	client := &http.Client{}
	for i := 0; i < 10; i++ {
		body, err := json.Marshal(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      queryId,
			"method":  method,
			"params":  params,
		})
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		queryId++
		req, err := http.NewRequest("POST", rpcUrl, bytes.NewReader(body))
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		req.Header.Add("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		ret, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		res.Body.Close()
		var rst struct {
			Result interface{} `json:"result"`
		}
		err = json.Unmarshal(ret, &rst)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		result = rst.Result
		if result != nil && result != "0x" {
			break
		}
	}
	return
}

func rawSafeRpcCall(method string, params interface{}) (result []byte, err error) {
	client := &http.Client{}
	for i := 0; i < 10; i++ {
		body, err := json.Marshal(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      queryId,
			"method":  method,
			"params":  params,
		})
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		queryId++
		req, err := http.NewRequest("POST", rpcUrl, bytes.NewReader(body))
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		req.Header.Add("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		ret, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		res.Body.Close()
		var rst struct {
			Result interface{} `json:"result"`
		}
		err = json.Unmarshal(ret, &rst)
		if err != nil {
			fmt.Println(string(ret))
			return nil, errors.Wrap(err, 0)
		}
		out := rst.Result
		result = ret
		if out != nil && out != "0x" {
			break
		}
	}
	return
}

func historicSafeRpcCall(method string, params interface{}) (result interface{}, err error) {
	client := &http.Client{}
	for i := 0; i < 10; i++ {
		body, err := json.Marshal(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      queryId,
			"method":  method,
			"params":  params,
		})
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		queryId++
		req, err := http.NewRequest("POST", historicRpcUrl, bytes.NewReader(body))
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		req.Header.Add("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		ret, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		res.Body.Close()
		var rst struct {
			Result interface{} `json:"result"`
		}
		err = json.Unmarshal(ret, &rst)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		result = rst.Result
		if result != nil && result != "0x" {
			break
		}
	}
	return
}
