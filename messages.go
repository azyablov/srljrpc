package srljrpc

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/azyablov/srljrpc/datastores"
	"github.com/azyablov/srljrpc/formats"
	"github.com/azyablov/srljrpc/methods"
)

// note for Request "JSON RPC Request: get / set / validate"
//
//	class Request {
//		<<message>>
//		note "Mandatory. Version, which must be ‟2.0”. No other JSON RPC versions are currently supported."
//		~string JSONRpcVersion
//		note "Mandatory. Client-provided integer. The JSON RPC responds with the same ID, which allows the client to match requests to responses when there are concurrent requests."
//		~int ID
//		+Marshal() List~byte~
//		+GetID() int
//	}
//
// Request *-- Method
// Request *-- Params
type Request struct {
	JSONRpcVersion string `json:"jsonrpc"`
	ID             int    `json:"id"`
	*methods.Method
	*Params
}

func (r *Request) Marshal() ([]byte, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *Request) GetID() int {
	return r.ID
}

func (r *Request) setID(id int) {
	r.ID = id
}

func (r *Request) SetOutputFormat(of formats.EnumOutputFormats) error {
	return r.Params.OutputFormat.SetFormat(of)
}

//	class Requester {
//		<<interface>>
//		+Marshal() List~byte~
//		+GetID() int
//		+GetMethod() EnumMethods
//		+MethodName() string
//		+SetOutputFormat(of formats.EnumOutputFormats) error
//	}
type Requester interface {
	Marshal() ([]byte, error)
	GetMethod() (methods.EnumMethods, error)
	MethodName() string
	GetID() int
	SetOutputFormat(of formats.EnumOutputFormats) error
}

//	class RequestOption {
//		<<function>>
//		(Request c) error
//	}
type RequestOption func(*Request) error

// +WithOutputFormat(EnumOutputFormats of) RequestOption
func WithOutputFormat(of formats.EnumOutputFormats) RequestOption {
	return func(r *Request) error {
		return r.SetOutputFormat(of)
	}
}

// +NewRequest(EnumMethods m, List~GetCommand~ cmds, List~RequestOption~ opts) Request
func NewRequest(m methods.EnumMethods, cmds []*Command, opts ...RequestOption) (*Request, error) {
	r := &Request{}
	// set version
	r.JSONRpcVersion = "2.0"

	// set method
	r.Method = &methods.Method{}
	err := r.Method.SetMethod(m)
	if err != nil {
		return nil, err
	}

	// set random ID
	rand.Seed(time.Now().UnixNano())
	id := rand.Int()
	r.setID(id)

	// set params and output format
	r.Params = &Params{}
	r.Params.OutputFormat = &formats.OutputFormat{}

	// set commands
	err = apply_cmds(r, cmds)
	if err != nil {
		return nil, err
	}

	// apply options to request
	err = apply_opts(r, opts)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func apply_cmds(r *Request, cmds []*Command) error {
	// check if commands are empty
	if len(cmds) == 0 {
		return fmt.Errorf("no commands given")
	}

	// check if commands are valid for the selected method
	m, err := r.Method.GetMethod()
	if err != nil {
		return err
	}
	switch m {
	case methods.GET:
		for _, c := range cmds {
			// path command - Mandatory with the get, set and validate methods.
			if c.Path == "" {
				return fmt.Errorf("path not found, but should be specified for method %s", r.Method.MethodName())
			}
			// The get method can be used with candidate, running, and state datastores, but cannot be used with the tools datastore.
			d, err := c.GetDatastore()
			if err != nil {
				return err
			}
			if d == datastores.TOOLS {
				return fmt.Errorf("datastore %s not allowed for method %s", c.DatastoreName(), r.Method.MethodName())
			}
			// Action and value are not allowed for the get method.
			if c.Action != nil {
				return fmt.Errorf("action not allowed for method %s", r.Method.MethodName())
			}
			if c.Value != "" {
				return fmt.Errorf("value not allowed for method %s", r.Method.MethodName())
			}
		}
	case methods.SET:
		for _, c := range cmds {
			// path command - Mandatory with the get, set and validate methods.
			if c.Path == "" {
				return fmt.Errorf("path not found, but should be specified for method %s", r.Method.MethodName())
			}
			// Used to verify that the system accepts a configuration transaction before applying it to the system.
			d, err := c.GetDatastore()
			if err != nil {
				return err
			}
			if d != datastores.CANDIDATE && d != datastores.TOOLS {
				return fmt.Errorf("datastore %s not allowed for method %s", c.DatastoreName(), r.Method.MethodName())
			}
			if c.Action == nil {
				return fmt.Errorf("action not found, but should be specified for method %s", r.Method.MethodName())
			}
			if c.Value == "" && !strings.Contains(c.Action.Action, ":") {
				return fmt.Errorf("value isn't specified or not found in the path for method %s", r.Method.MethodName())
			}
		}
	case methods.VALIDATE:
		for _, c := range cmds {
			// path command - Mandatory with the get, set and validate methods.
			if c.Path == "" {
				return fmt.Errorf("path not found, but should be specified for method %s", r.Method.MethodName())
			}
			// Used to verify that the system accepts a configuration transaction before applying it to the system.
			d, err := c.GetDatastore()
			if err != nil {
				return err
			}
			if d != datastores.CANDIDATE {
				return fmt.Errorf("datastore %s not allowed for method %s", c.DatastoreName(), r.Method.MethodName())
			}
			if c.Action == nil {
				return fmt.Errorf("action not found, but should be specified for method %s", r.Method.MethodName())
			}
			if c.Value == "" && !strings.Contains(c.Action.Action, ":") {
				return fmt.Errorf("value isn't specified or not found in the path for method %s", r.Method.MethodName())
			}
		}
	case methods.CLI:
		return fmt.Errorf("method %s not supported by Request, please use CLIRequest object", r.Method.MethodName())
	default:
		return fmt.Errorf("method %s not supported by Request", r.Method.MethodName())
	}
	// checks passed, append commands to request
	r.appendCommands(cmds)
	return nil
}

// function applies options to the request
func apply_opts(r *Request, opts []RequestOption) error {
	for _, o := range opts {
		if err := o(r); err != nil {
			return nil
		}
	}
	return nil
}

//	class CLIRequest {
//		<<message>>
//		note "Method set to CLI"
//		note "Mandatory. Version, which must be ‟2.0”. No other JSON RPC versions are currently supported."
//		~string JSONRpcVersion
//		note "Mandatory. Client-provided integer. The JSON RPC responds with the same ID, which allows the client to match requests to responses when there are concurrent requests."
//		~int ID
//		note "Mandatory. Supported options are cli. Set statically in the RPC request"
//		+Marshal() List~byte~
//		+GetID() int
//		~setID(int)
//	}
//
// CLIRequest *-- Method
// CLIRequest *-- CLIParams
type CLIRequest struct {
	JSONRpcVersion string `json:"jsonrpc"`
	ID             int    `json:"id"`
	*methods.Method
	*CLIParams
}

func (r *CLIRequest) Marshal() ([]byte, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *CLIRequest) GetID() int {
	return r.ID
}

func (r *CLIRequest) setID(id int) {
	r.ID = id
}

func (r *CLIRequest) SetOutputFormat(of formats.EnumOutputFormats) error {
	return r.CLIParams.OutputFormat.SetFormat(of)
}

// +NewCLIRequest(List~string~ cmds, List~RequestOption~ opts) CLIRequest
func NewCLIRequest(cmds []string, of formats.EnumOutputFormats) (*CLIRequest, error) {
	r := &CLIRequest{}
	// set version
	r.JSONRpcVersion = "2.0"

	// set method
	err := r.Method.SetMethod(methods.CLI)
	if err != nil {
		return nil, err
	}

	// set random ID
	rand.Seed(time.Now().UnixNano())
	id := rand.Int()
	r.setID(id)

	// set commands
	r.appendCommands(cmds)

	// apply options to request
	err = r.SetOutputFormat(of)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// note for RpcError "When a rpc call is made, the Server MUST reply with a Response, except for in the case of Notifications. The Response is expressed as a single JSON Object"
//
//	class RpcError {
//		<<element>>
//		note "A Number that indicates the error type that occurred. This MUST be an integer."
//		+int ID
//		note "A String providing a short description of the error. The message SHOULD be limited to a concise single sentence."
//		+string Message
//		note "A Primitive or Structured value that contains additional information about the error. This may be omitted. The value of this member is defined by the Server (e.g. detailed error information, nested errors etc.)."
//		+string Data
//	}
type RpcError struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// note for Response "JSON RPC response message. When a rpc call is made, the Server MUST reply with a Response, except for in the case of Notifications. The Response is expressed as a single JSON Object."
//
//	class Response {
//		<<message>>
//		note "Mandatory. Version, which must be ‟2.0”. No other JSON RPC versions are currently supported."
//		~string JSONRpcVersion
//		note "Mandatory. Client-provided integer. The JSON RPC responds with the same ID, which allows the client to match requests to responses when there are concurrent requests."
//		~int ID
//		note "This member is REQUIRED on success. This member MUST NOT exist if there was an error invoking the method. The value of this member is determined by the method invoked on the Server."
//		+jsonRawMessage Result
//		note "This member is REQUIRED on error. This member MUST NOT exist if there was no error triggered during invocation. The value for this member MUST be an Object as defined in section 5.1."
//		+RpcError Error
//	}
//
// Response o-- RpcError
type Response struct {
	JSONRpcVersion string          `json:"jsonrpc"`
	ID             int             `json:"id"`
	Result         json.RawMessage `json:"result,omitempty"`
	Error          *RpcError       `json:"error,omitempty"`
}

func (r *Response) Marshal() ([]byte, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *Response) GetID() int {
	return r.ID
}
