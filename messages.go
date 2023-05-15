package srljrpc

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/azyablov/srljrpc/actions"
	"github.com/azyablov/srljrpc/datastores"
	"github.com/azyablov/srljrpc/formats"
	"github.com/azyablov/srljrpc/methods"
)

// NewRequest provides a new Request with the given method, commands and options.
// Sequence of functions is applied to the Request in the order of appearance.
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
	r.Params.Datastore = &datastores.Datastore{}

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

// JSON RPC Request for get / set / validate methods.
//
//	JSONRpcVersion is mandatory. Version, which must be ‟2.0”. No other JSON RPC versions are currently supported.
//	ID is mandatory. Client-provided integer. The JSON RPC responds with the same ID, which allows the client to match requests to responses when there are concurrent requests.
//	Implementation uses random numbers and verifies Response ID is the same as Request ID.
//	Embeds Method and Params.
type Request struct {
	JSONRpcVersion string `json:"jsonrpc"`
	ID             int    `json:"id"`
	*methods.Method
	Params *Params `json:"params"`
}

// Marshaling of the Request into JSON.
func (r *Request) Marshal() ([]byte, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Get Request ID.
func (r *Request) GetID() int {
	return r.ID
}

// Set Request ID.
func (r *Request) setID(id int) {
	r.ID = id
}

// Set output format for the request via embedded Params.
func (r *Request) SetOutputFormat(of formats.EnumOutputFormats) error {
	return r.Params.OutputFormat.SetFormat(of)
}

// Requester is an interface used by the JSON RPC client to send a request to the server.
type Requester interface {
	Marshal() ([]byte, error)
	GetMethod() (methods.EnumMethods, error)
	MethodName() string
	GetID() int
	SetOutputFormat(of formats.EnumOutputFormats) error
}

// RequestOption is a function type that applies options to a Request.
type RequestOption func(*Request) error

// Defines output format RequestOption.
func WithOutputFormat(of formats.EnumOutputFormats) RequestOption {
	return func(r *Request) error {
		return r.SetOutputFormat(of)
	}
}

// RequestOption that sets the datastore for the request in Params level. Overrides the datastore in Command level!
func WithRequestDatastore(ds datastores.EnumDatastores) RequestOption {
	return func(r *Request) error {
		m, _ := r.GetMethod()
		switch m {
		case methods.GET:
			if ds == datastores.TOOLS {
				return fmt.Errorf("datastore TOOLS is not allowed for method %s", r.Method.MethodName())
			}
			return r.Params.withDatastore(ds)
		case methods.SET:
			for _, c := range r.Params.Commands {
				c.CleanDatastore() // clean datastore in commands, to be later if such protective measures are needed, since c.IsDefaultDatastore() added as verification check for SET/VALIDATE
				a, err := c.Action.GetAction()
				if err != nil {
					return err
				}
				// The set method can be used with tools datastores only with the update action.
				if ds == datastores.TOOLS && a != actions.UPDATE {
					return fmt.Errorf("only update method is allowed with TOOLS datastore")
				}
			}
			if ds != datastores.CANDIDATE && ds != datastores.TOOLS {
				return fmt.Errorf("only CANDIDATE and TOOLS datastores allowed for method %s", r.Method.MethodName())
			}
			return r.Params.withDatastore(ds)
		case methods.VALIDATE:
			// clean datastore in commands
			for _, c := range r.Params.Commands {
				c.CleanDatastore() // clean datastore in commands, to be later if such protective measures are needed, since c.IsDefaultDatastore() added as verification check for SET/VALIDATE
			}
			if ds != datastores.CANDIDATE {
				return fmt.Errorf("only CANDIDATE datastore allowed for method %s", r.Method.MethodName())
			}
			return r.Params.withDatastore(ds)
		default:
			return fmt.Errorf("datastore specification on Request.Params level is not supported for method %s", r.MethodName())
		}
	}
}

// Helper function to add commands to the request and verify if they are valid for the selected method i.e. implements JSON RPC API specification rules.
func apply_cmds(r *Request, cmds []*Command) error {
	// check if commands are empty
	if len(cmds) == 0 {
		return fmt.Errorf("no commands given")
	}
	// check if commands aren't empty
	for _, c := range cmds {
		if c == nil {
			return fmt.Errorf("nil commands aren't allowed")
		}
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
				return fmt.Errorf("datastore TOOLS is not allowed for method %s", r.Method.MethodName())
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

			// Check if datastore has been set to default datastore.
			if !c.IsDefaultDatastore() {
				return fmt.Errorf("command level datastore must not be set for method %s", r.Method.MethodName())
			}

			// Action command is mandatory with the set method.
			if c.Action == nil {
				return fmt.Errorf("action not found, but should be specified for method %s", r.Method.MethodName())
			}
			a, err := c.Action.GetAction()
			if err != nil {
				return err
			}
			// check if action is valid for the set method
			if a == actions.NONE {
				return fmt.Errorf("action not found, but should be specified for method %s", r.Method.MethodName())
			}
			// Check if value is specified for the set method.
			if c.Value == "" && !strings.Contains(c.Path, ":") && a != actions.DELETE {
				return fmt.Errorf("value isn't specified or not found in the path for method %s", r.Method.MethodName())
			}
			if c.Value != "" && a == actions.DELETE {
				return fmt.Errorf("value specified for action DELETE for method %s", r.Method.MethodName())
			}
			// Check if value is specified in the path and as a separate value for the set method.
			if strings.Contains(c.Path, ":") {
				sl := strings.Split(c.Path, ":")
				if len(sl) != 2 {
					return fmt.Errorf("invalid k:v path specification for method %s", r.Method.MethodName())
				}
				if c.Value != "" {
					return fmt.Errorf("value specified in the path and as a separate value for method %s", r.Method.MethodName())
				}
			}
		}
	case methods.VALIDATE:
		for _, c := range cmds {
			// path command - Mandatory with the get, set and validate methods.
			if c.Path == "" {
				return fmt.Errorf("path not found, but should be specified for method %s", r.Method.MethodName())
			}

			// Check if datastore has been set to default datastore.
			if !c.IsDefaultDatastore() {
				return fmt.Errorf("command level datastore must not be set for method %s", r.Method.MethodName())
			}
			// Action command is mandatory with the set method.
			if c.Action == nil {
				return fmt.Errorf("action not found, but should be specified for method %s", r.Method.MethodName())
			}
			a, err := c.Action.GetAction()
			if err != nil {
				return err
			}
			// check if action is valid for the validate method
			if a == actions.NONE {
				return fmt.Errorf("action not found, but should be specified for method %s", r.Method.MethodName())
			}
			// Check if value is specified for the set method.
			if c.Value == "" && !strings.Contains(c.Path, ":") {
				return fmt.Errorf("value isn't specified or not found in the path for method %s", r.Method.MethodName())
			}
			if c.Value != "" && a == actions.DELETE {
				return fmt.Errorf("value specified for action DELETE for method %s", r.Method.MethodName())
			}
			if strings.Contains(c.Path, ":") {
				sl := strings.Split(c.Path, ":")
				if len(sl) != 2 {
					return fmt.Errorf("invalid k:v path specification for method %s", r.Method.MethodName())
				}
				if c.Value != "" {
					return fmt.Errorf("value specified in the path and as a separate value for method %s", r.Method.MethodName())
				}
			}
		}
	case methods.CLI:
		return fmt.Errorf("method %s not supported by Request, please use CLIRequest object", r.Method.MethodName())
	default:
		return fmt.Errorf("method %s not supported by Request", r.Method.MethodName())
	}
	// checks passed, append commands to request
	err = r.Params.appendCommands(cmds)
	if err != nil {
		return err
	}

	return nil
}

// Helper function applies options to the request.
func apply_opts(r *Request, opts []RequestOption) error {
	for _, o := range opts {
		if o != nil { // check that's not nil
			if err := o(r); err != nil {
				return nil
			}
		}
	}
	return nil
}

// Creates a new CLIRequest object using the provided list of commands executed one by one and output format.
// Each command should have a response under JSON RPC response message - Response under "result" field with respective command index.
func NewCLIRequest(cmds []string, of formats.EnumOutputFormats) (*CLIRequest, error) {
	r := &CLIRequest{}
	// set version
	r.JSONRpcVersion = "2.0"

	// set method
	r.Method = &methods.Method{}
	err := r.Method.SetMethod(methods.CLI)
	if err != nil {
		return nil, err
	}

	// set random ID
	rand.Seed(time.Now().UnixNano())
	id := rand.Int()
	r.setID(id)
	// set params
	r.Params = &CLIParams{}
	r.Params.OutputFormat = &formats.OutputFormat{}

	// set commands
	err = r.Params.appendCommands(cmds)
	if err != nil {
		return nil, err
	}

	// apply options to request
	err = r.SetOutputFormat(of)
	if err != nil {
		return nil, err
	}

	return r, nil
}

//	JSONRpcVersion is mandatory. Version, which must be ‟2.0”. No other JSON RPC versions are currently supported.
//
// ID is mandatory. Client-provided integer. The JSON RPC responds with the same ID, which allows the client to match requests to responses when there are concurrent requests.
// Implementation uses random numbers and verifies Response ID is the same as Request ID.
// Embeds Method (set to CLI by NewCLIRequest) and CLIParams.
type CLIRequest struct {
	JSONRpcVersion string `json:"jsonrpc"`
	ID             int    `json:"id"`
	*methods.Method
	Params *CLIParams `json:"params"`
}

// Marshalling CLIRequest into JSON.
func (r *CLIRequest) Marshal() ([]byte, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Returns the ID of the request.
func (r *CLIRequest) GetID() int {
	return r.ID
}

// Sets the ID of the request. Internal method.
func (r *CLIRequest) setID(id int) {
	r.ID = id
}

// Sets the output format of the request.
func (r *CLIRequest) SetOutputFormat(of formats.EnumOutputFormats) error {
	return r.Params.OutputFormat.SetFormat(of)
}

// RpcError is generic JSON RPC error object.
// When a rpc call is made, the Server MUST reply with a Response, except for in the case of Notifications. The Response is expressed as a single JSON Object.
//
//	ID should be set to client provided ID.
//	Message is a string providing a short description of the error. The message SHOULD be limited to a concise single sentence."
//	Data is a primitive or structured value that contains additional information about the error. This may be omitted. The value of this member is defined by the Server (e.g. detailed error information, nested errors etc.).
type RpcError struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// JSON RPC response message. When a rpc call is made, the Server MUST reply with a Response, except for in the case of Notifications. The Response is expressed as a single JSON Object.
// Result and error are mutually exclusive, so only one of them can be expected. Error is represented as a pointer to RpcError, so it can be nil.
//
//	JSONRpcVersion is mandatory. Version, which must be ‟2.0”. No other JSON RPC versions are currently supported.
//	ID is mandatory. Client-provided integer. The JSON RPC responds with the same ID, which allows the client to match requests to responses when there are concurrent requests.
//	Result is REQUIRED on success (jsonRawMessage). This member MUST NOT exist if there was an error invoking the method. The value of this member is determined by the method invoked on the Server.
//	Error is REQUIRED on error. This member MUST NOT exist if there was no error triggered during invocation. The value for this member MUST be an RpcError object.
type Response struct {
	JSONRpcVersion string          `json:"jsonrpc"`
	ID             int             `json:"id"`
	Result         json.RawMessage `json:"result,omitempty"`
	Error          *RpcError       `json:"error,omitempty"`
}

// Marshalling Response into JSON.
func (r *Response) Marshal() ([]byte, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Returns ID of the response in order to compare with request ID.
func (r *Response) GetID() int {
	return r.ID
}
