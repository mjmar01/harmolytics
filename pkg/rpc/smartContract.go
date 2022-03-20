package rpc

const (
	contractCall = "hmyv2_call"
)

// SimpleCall executes a read only transaction to the specified contract with given data.
func SimpleCall(to, data string) (r string, err error) {
	// Get return value
	params := []interface{}{
		struct {
			To   string `json:"to"`
			Data string `json:"data"`
		}{
			To:   to,
			Data: data,
		},
		"latest",
	}
	result, err := rpcCall(contractCall, params)
	if err != nil {
		return
	}
	r = result.(string)
	return
}

// HistoricCall executes a read only transaction to the specified contract with given data at the given block.
func HistoricCall(to, data string, blockNum uint64) (r string, err error) {
	// Get return value
	params := []interface{}{
		struct {
			To   string `json:"to"`
			Data string `json:"data"`
		}{
			To:   to,
			Data: data,
		},
		blockNum,
	}
	result, err := historicRpcCall(contractCall, params)
	if err != nil {
		return
	}
	r = result.(string)
	return
}