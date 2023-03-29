package jsonrpc

//	class JSONRPCClient {
//		<<entity>>
//		Call(Requester r) Response
//	}
type JSONRPCClient struct {
}

// Call method of JSONRPCClient
// note for Call "Mandatory. Calls the JSON RPC server and returns the response."
// note for r "Mandatory. The Requester interface."
func (c *JSONRPCClient) Call(r Requester) Response {
	return nil
}
