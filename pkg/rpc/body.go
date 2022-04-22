package rpc

// NewBody prepares a body with correct syntax and incremental IDs
func (r *RPC) NewBody(method string, params ...interface{}) (b Body) {
	if params == nil {
		params = []interface{}{}
	}
	b = Body{
		RpcVersion: "2.0",
		Id:         r.queryId,
		Method:     method,
		Params:     params,
	}
	r.queryId++
	return
}
