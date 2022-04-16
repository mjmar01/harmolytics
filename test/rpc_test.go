package test

import (
	"encoding/json"
	"github.com/mjmar01/harmolytics/pkg/rpc"
	"strconv"
	"sync"
	"testing"
	"time"
)

const (
	TransactionByHashMethod = "hmyv2_getTransactionByHash"
	BlockNumberMethod       = "hmyv2_blockNumber"
	CallMethod              = "hmyv2_call"
)

func TestNewRpc(t *testing.T) {
	t.Parallel()
	var err error
	_, err = rpc.NewRpc(url, nil)
	if err != nil {
		t.Error(err)
	}
	_, err = rpc.NewRpc(url, &rpc.Opts{Timeout: time.Minute * 3})
	if err != nil {
		t.Error(err)
	}

	_, err = rpc.NewRpcs(url, 2, nil)
	if err != nil {
		t.Error(err)
	}
	_, err = rpc.NewRpcs(url, 2, &rpc.Opts{Timeout: time.Minute * 3})
	if err != nil {
		t.Error(err)
	}
}

func TestCall(t *testing.T) {
	t.Parallel()
	r, _ := rpc.NewRpc(url, nil)

	res, _ := r.Call(TransactionByHashMethod, "0x41d6e74ff3a7e615080b98fcfb7bce8be7b1ba4a8671e1ba2e9527eb3e1da20d")
	// Valid TX
	if res.(map[string]interface{})["from"] != "one1a5fznwvnr3fed9676g42u7q30crtmmkk5qspe9" {
		t.Errorf("Result did not contain expected key: %v", res)
	}
}

func TestBatchCall(t *testing.T) {
	t.Parallel()
	r, _ := rpc.NewRpc(url, nil)
	bodies := make([]rpc.Body, 3)
	bodies[0] = r.NewBody(TransactionByHashMethod, "0x41d6e74ff3a7e615080b98fcfb7bce8be7b1ba4a8671e1ba2e9527eb3e1da20d")
	bodies[1] = r.NewBody(BlockNumberMethod)
	bodies[2] = r.NewBody(CallMethod, map[string]string{"to": "0xcf664087a5bb0237a0bad6742852ec6c8d69a27a", "data": "0x313ce567"}, "latest")
	ress, _ := r.BatchCall(bodies)
	// Valid TX
	if ress[0].(map[string]interface{})["from"] != "one1a5fznwvnr3fed9676g42u7q30crtmmkk5qspe9" {
		t.Errorf("BatchCall0 did not return valid map: %v", ress[0])
	}
	// Valid number
	if i, ok := ress[1].(float64); !ok || float64(int64(i)) != i {
		t.Errorf("BatchCall1 did not return valid int: %v", ress[1])
	}
	// Valid string
	if _, ok := ress[2].(string); !ok {
		t.Errorf("BatchCall2 did not return valid string: %v", ress[2])
	}
}

func TestRawCall(t *testing.T) {
	t.Parallel()
	r, _ := rpc.NewRpc(url, nil)
	res, _ := r.RawCall(TransactionByHashMethod, "0x41d6e74ff3a7e615080b98fcfb7bce8be7b1ba4a8671e1ba2e9527eb3e1da20d")
	// Valid JSON
	var tmp interface{}
	if json.Unmarshal(res, &tmp) != nil {
		t.Errorf("Result did not contain valid JSON: %v", res)
	}
}

func TestRawBatchCall(t *testing.T) {
	t.Parallel()
	r, _ := rpc.NewRpc(url, nil)
	bodies := make([]rpc.Body, 3)
	bodies[0] = r.NewBody(TransactionByHashMethod, "0x41d6e74ff3a7e615080b98fcfb7bce8be7b1ba4a8671e1ba2e9527eb3e1da20d")
	bodies[1] = r.NewBody(BlockNumberMethod)
	bodies[2] = r.NewBody(CallMethod, map[string]string{"to": "0xcf664087a5bb0237a0bad6742852ec6c8d69a27a", "data": "0x313ce567"}, "latest")
	ress, _ := r.RawBatchCall(bodies)
	// Valid JSON
	var tmp interface{}
	if json.Unmarshal(ress[0], &tmp) != nil {
		t.Errorf("BatchCall0 did not return valid map: %v", ress[0])
	}
	// Valid number
	if _, err := strconv.Atoi(string(ress[1])); err != nil {
		t.Errorf("BatchCall1 did not return valid int: %v", ress[1])
	}
	// Valid string
	if string(ress[2]) != "\"0x0000000000000000000000000000000000000000000000000000000000000012\"" {
		t.Errorf("BatchCall2 did not return valid string: %v", ress[2])
	}
}

func TestBatches(t *testing.T) {
	t.Parallel()
	rs, _ := rpc.NewRpcs(url, 10, nil)
	wg := sync.WaitGroup{}
	wg.Add(10)
	ch := make(chan interface{}, 20)
	for i := 0; i < 10; i++ {
		go func(r *rpc.Rpc) {
			ress, _ := r.BatchCall([]rpc.Body{r.NewBody(BlockNumberMethod), r.NewBody(BlockNumberMethod)})
			ch <- ress[0]
			ch <- ress[1]
			wg.Done()
		}(rs[i])
	}
	wg.Wait()
	for i := 0; i < 20; i++ {
		n := <-ch
		if i, ok := n.(float64); !ok || float64(int64(i)) != i {
			t.Errorf("BatchCall1 did not return valid int: %v", n)
		}
	}
}
