package srljrpc

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/azyablov/srljrpc/actions"
	"github.com/azyablov/srljrpc/apierr"
	"github.com/azyablov/srljrpc/datastores"
	"github.com/azyablov/srljrpc/formats"
	"github.com/azyablov/srljrpc/methods"
	"github.com/azyablov/srljrpc/yms"
)

// TLSAttr type to represent TLS attributes
type TLSAttr struct {
	CAFile     *string `json:"ca_file,omitempty"`     // CA certificate file in PEM format.
	CertFile   *string `json:"cert_file,omitempty"`   // Client certificate file in PEM format.
	KeyFile    *string `json:"key_file,omitempty"`    // Client private key file.
	SkipVerify *bool   `json:"skip_verify,omitempty"` // Disable certificate validation during TLS session ramp-up.
}

type cred struct {
	username *string
	password *string
}

type targetHost struct {
	host    *string
	port    *int
	timeout time.Duration
}

// JSONRPCTarget type to represent a JSON RPC target: NE(target), TLS attributes, credentials.
type JSONRPCTarget struct {
	targetHost
	cred
	tlsConfig *tls.Config
}

// JSONRPCClient type to represent a JSON RPC client: HTTP client, NE(target) and related info.
type JSONRPCClient struct {
	client   *http.Client
	hostname string
	sysVer   string
	target   *JSONRPCTarget
	mux      sync.Mutex
}

// PV type to represent a path-value pair.
type PV struct {
	Path  string       `json:"path"`
	Value CommandValue `json:"value"`
}

// ClientOption is a function type that applies options to a JSONRPCClient object.
type ClientOption func(*JSONRPCClient) error

// CallBackConfirm type to represent a callback function to confirm a request.
// In case of confirm commit must return true, otherwise false.
type CallBackConfirm func(req *Request, resp *Response) (bool, error)

// Creates a new JSON RPC client and applies options in order of appearance.
func NewJSONRPCClient(host *string, opts ...ClientOption) (*JSONRPCClient, error) {
	// client object
	c := &JSONRPCClient{}
	c.target = &JSONRPCTarget{}
	// host
	if host == nil {
		return nil, apierr.ClientError{
			CltFunction: "NewJSONRPCClient",
			Code:        apierr.ErrClntNoHost,
			Err:         nil,
		}
	}
	c.target.host = host

	// applying options
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}

	// checking inputs and populating defaults
	err := c.populateDefaults()
	if err != nil {
		return nil, err
	}

	// ... creating a new HTTP client
	c.client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:          32,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       c.target.tlsConfig,
		},
		Timeout: c.target.timeout,
	}

	// verify target validity and availability
	err = c.targetVerification()
	if err != nil {
		return nil, apierr.ClientError{
			CltFunction: "NewJSONRPCClient",
			Code:        apierr.ErrClntTargetVerification,
			Err:         err,
		}
	}

	return c, nil
}

// GetSysVer returns the system version of the target after verification.
func (c *JSONRPCClient) GetSysVer() string {
	return c.sysVer
}

// GetHostname returns the hostname of the target after verification.
func (c *JSONRPCClient) GetHostname() string {
	return c.hostname
}

// Calls the JSON RPC server and returns the response.
func (c *JSONRPCClient) Do(r Requester) (*Response, error) {
	body, err := r.Marshal()
	if err != nil {
		return nil, apierr.ClientError{
			CltFunction: "Do",
			Code:        apierr.ErrClntMarshalling,
			Err:         err,
		}
	}

	reqHTTP, err := http.NewRequest("POST", fmt.Sprintf("https://%s:%v/jsonrpc", *c.target.host, *c.target.port), bytes.NewBuffer(body))
	if err != nil {
		return nil, apierr.ClientError{
			CltFunction: "Do",
			Code:        apierr.ErrClntHTTPReqCreation,
			Err:         err,
		}
	}

	// setting content type and authentication header
	reqHTTP.Header.Set("Content-Type", "application/json")
	reqHTTP.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", *c.target.username, *c.target.password))))

	resp, err := c.client.Do(reqHTTP)
	if err != nil {
		return nil, apierr.ClientError{
			CltFunction: "Do",
			Code:        apierr.ErrClntHTTPSend,
			Err:         err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apierr.ClientError{
			CltFunction: "Do",
			Code:        apierr.ErrClntHTTPStatus,
			Err:         err,
		}
	}
	rpcResp := Response{}
	err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	if err != nil {
		return nil, apierr.ClientError{
			CltFunction: "Do",
			Code:        apierr.ErrClntJSONUnmarshalling,
			Err:         err,
		}
	}
	if rpcResp.GetID() != r.GetID() {
		return nil, apierr.ClientError{
			CltFunction: "Do",
			Code:        apierr.ErrClntIDMismatch,
			Err:         err,
		}
	}

	if rpcResp.Error != nil {
		return &rpcResp, apierr.ClientError{
			CltFunction: "Do",
			Code:        apierr.ErrClntJSONRPC,
			Err:         err,
		}
	}

	return &rpcResp, nil
}

// Get method of JSONRPCClient. Executes a GET request against RUNNING datastore.
func (c *JSONRPCClient) Get(paths ...string) (*Response, error) {
	opts := []CommandOption{WithDatastore(datastores.RUNNING)}
	var cmds []*Command
	for _, path := range paths {
		cmd, err := NewCommand(actions.NONE, path, CommandValue(""), opts...)
		if err != nil {
			return nil, apierr.ClientError{
				CltFunction: "Get",
				Code:        apierr.ErrClntCmdCreation,
				Err:         err,
			}
		}
		cmds = append(cmds, cmd)

	}
	// create the request
	r, err := NewRequest(methods.GET, cmds)
	if err != nil {
		return nil, apierr.ClientError{
			CltFunction: "Get",
			Code:        apierr.ErrClntRPCReqCreation,
			Err:         err,
		}
	}
	return c.Do(r)
}

// Get state method of JSONRPCClient. Executes a GET request against STATE datastore.
func (c *JSONRPCClient) State(paths ...string) (*Response, error) {
	opts := []CommandOption{WithDatastore(datastores.STATE)}
	var cmds []*Command
	for _, path := range paths {
		cmd, err := NewCommand(actions.NONE, path, CommandValue(""), opts...)
		if err != nil {
			return nil, apierr.ClientError{
				CltFunction: "State",
				Code:        apierr.ErrClntCmdCreation,
				Err:         err,
			}
		}
		cmds = append(cmds, cmd)
	}
	r, err := NewRequest(methods.GET, cmds, nil)
	if err != nil {
		return nil, apierr.ClientError{
			CltFunction: "State",
			Code:        apierr.ErrClntRPCReqCreation,
			Err:         err,
		}
	}
	return c.Do(r)
}

// SetUpdate method of JSONRPCClient executing a SET/UPDATE action request against CANDIDATE datastore.
// ct is the timeout in seconds for the confirm operation, set to 0 to disable.
// pvs is the list of path-value pairs. Yang model type is default(SRL).
func (c *JSONRPCClient) Update(ct int, pvs ...PV) (*Response, error) {
	var cmds []*Command
	for _, pv := range pvs {
		cmd, err := NewCommand(actions.UPDATE, pv.Path, CommandValue(pv.Value))
		if err != nil {
			return nil, apierr.ClientError{
				CltFunction: "Update",
				Code:        apierr.ErrClntCmdCreation,
				Err:         err,
			}
		}

		cmds = append(cmds, cmd)
	}

	// build the request
	var r *Request
	var err error
	if ct == 0 {
		r, err = NewRequest(methods.SET, cmds, WithRequestDatastore(datastores.CANDIDATE))
	} else {
		r, err = NewRequest(methods.SET, cmds, WithRequestDatastore(datastores.CANDIDATE), WithConfirmTimeout(ct))
	}
	if err != nil {
		return nil, apierr.ClientError{
			CltFunction: "Update",
			Code:        apierr.ErrClntRPCReqCreation,
			Err:         err,
		}
	}
	return c.Do(r)
}

// SetReplace method of JSONRPCClient. Executes a SET/REPLACE action request against CANDIDATE datastore.
// ct is the timeout in seconds for the confirm operation, set to 0 to disable.
// pvs is the list of path-value pairs. Yang model type is default(SRL).
func (c *JSONRPCClient) Replace(ct int, pvs ...PV) (*Response, error) {
	var cmds []*Command
	for _, pv := range pvs {
		cmd, err := NewCommand(actions.REPLACE, pv.Path, pv.Value)
		if err != nil {
			return nil, apierr.ClientError{
				CltFunction: "Replace",
				Code:        apierr.ErrClntCmdCreation,
				Err:         err,
			}
		}

		cmds = append(cmds, cmd)
	}

	// build the request
	var r *Request
	var err error
	if ct == 0 {
		r, err = NewRequest(methods.SET, cmds, WithRequestDatastore(datastores.CANDIDATE))
	} else {
		r, err = NewRequest(methods.SET, cmds, WithRequestDatastore(datastores.CANDIDATE), WithConfirmTimeout(ct))
	}
	if err != nil {
		return nil, apierr.ClientError{
			CltFunction: "Replace",
			Code:        apierr.ErrClntRPCReqCreation,
			Err:         err,
		}
	}
	return c.Do(r)
}

// SetDelete method of JSONRPCClient. Executes a SET/DELETE action request against CANDIDATE datastore.
// t is the timeout in seconds for the confirm operation, set to 0 to disable.
// paths is the list of path to delete. Yang model type is default(SRL).
func (c *JSONRPCClient) Delete(ct int, paths ...string) (*Response, error) {
	// build the commands
	var cmds []*Command
	for _, path := range paths {
		cmd, err := NewCommand(actions.DELETE, path, CommandValue(""))
		if err != nil {
			return nil, apierr.ClientError{
				CltFunction: "Delete",
				Code:        apierr.ErrClntCmdCreation,
				Err:         err,
			}
		}

		cmds = append(cmds, cmd)
	}

	// build the request
	var r *Request
	var err error
	if ct == 0 {
		r, err = NewRequest(methods.SET, cmds, WithRequestDatastore(datastores.CANDIDATE))
	} else {
		r, err = NewRequest(methods.SET, cmds, WithRequestDatastore(datastores.CANDIDATE), WithConfirmTimeout(ct))
	}
	if err != nil {
		return nil, apierr.ClientError{
			CltFunction: "Delete",
			Code:        apierr.ErrClntRPCReqCreation,
			Err:         err,
		}
	}
	return c.Do(r)
}

// DiffCandidate method of JSONRPCClient. Executes a DIFF/<action> action request against CANDIDATE datastore. Yang model type is default(SRL).
// pvs are path-value pairs. The action parameter must be one of DELETE, REPLACE, or UPDATE.
func (c *JSONRPCClient) DiffCandidate(action actions.EnumActions, ym yms.EnumYmType, pvs ...PV) (*Response, error) {
	var delete, replace, update []PV
	// identify the action
	switch action {
	case actions.DELETE:
		delete = pvs
	case actions.REPLACE:
		replace = pvs
	case actions.UPDATE:
		update = pvs
	case actions.NONE:
		return nil, apierr.ClientError{
			CltFunction: "DiffCandidate",
			Code:        apierr.ErrClntActNONE,
			Err:         nil,
		}
	default:
		return nil, apierr.ClientError{
			CltFunction: "DiffCandidate",
			Code:        apierr.ErrClntActUnsupported,
			Err:         nil,
		}
	}
	r, err := NewDiffRequest(delete, replace, update, ym, formats.JSON, datastores.CANDIDATE)
	if err != nil {
		return nil, apierr.ClientError{
			CltFunction: "DiffCandidate",
			Code:        apierr.ErrClntRPCReqCreation,
			Err:         err,
		}
	}
	return c.Do(r)
}

// Bulk CRUD method of JSONRPCClient. Executes a SET method with REPLACE/UPDATE/DELETE action request against CANDIDATE datastore.
// ct is the timeout in seconds for the confirm operation, set to 0 to disable. delete/replace/update are path-value pairs.
// All the PVs are applied immediately in the same order as they are provided. yang model type is mandatory for diff to specify: SRL or OC.
func (c *JSONRPCClient) BulkSet(delete []PV, replace []PV, update []PV, ym yms.EnumYmType, ct int) (*Response, error) {
	// build the request
	r, err := NewSetRequest(delete, replace, update, ym, formats.JSON, datastores.CANDIDATE, ct)
	if err != nil {
		return nil, apierr.ClientError{
			CltFunction: "BulkSetCandidate",
			Code:        apierr.ErrClntRPCReqCreation,
			Err:         err,
		}
	}
	return c.Do(r)
}

// Bulk CRUD method of JSONRPCClient w/ CallBackConfirm callback and mandatory confirm timeout.
// Executes a SET method with REPLACE/UPDATE/DELETE action request against CANDIDATE datastore.
// All the PVs are applied immediately in the same order as they are provided. yang model type is mandatory for diff to specify: SRL or OC.
// JSON RPC Response is not nil if the callback function returns true. if callback function returns false, the both response&error are nil to indicate changes rolled back and NE back to previous state.
func (c *JSONRPCClient) BulkSetCallBack(delete []PV, replace []PV, update []PV, ym yms.EnumYmType, ct int, cbt int, cbf CallBackConfirm) (*Response, error) {
	// check cbt is lower than ct anc ct > 0 (mandatory)
	if ct <= cbt+1 && ct > 0 { // +1 to avoid 0
		return nil, apierr.ClientError{
			CltFunction: "BulkSetCallBack",
			Code:        apierr.ErrClntCBFuncLowerThanCT,
			Err:         nil,
		}
	}
	// check if the callback is nil
	if cbf == nil {
		return nil, apierr.ClientError{
			CltFunction: "BulkSetCallBack",
			Code:        apierr.ErrClntCBFuncIsNil,
			Err:         nil,
		}
	}
	// build the request
	req, err := NewSetRequest(delete, replace, update, ym, formats.JSON, datastores.CANDIDATE, ct)
	if err != nil {
		return nil, apierr.ClientError{
			CltFunction: "BulkSetCallBack",
			Code:        apierr.ErrClntRPCReqCreation,
			Err:         err,
		}
	}
	// execute the request
	c.mux.Lock()
	defer c.mux.Unlock()
	resp, err := c.Do(req)
	if err != nil {
		return resp, err
	}

	if ct-cbt > 2 {
		tch := time.After(time.Duration(cbt) * time.Second)
		<-tch
	}
	// execute the callback
	confirm, err := cbf(req, resp)
	if err != nil {
		return nil, apierr.ClientError{
			CltFunction: "BulkSetCallBack",
			Code:        apierr.ErrClntCBFuncExec,
			Err:         err,
		}
	}
	if confirm {
		_, err := c.Tools(PV{Path: "/system/configuration/confirmed-accept", Value: CommandValue("")})
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
	return nil, nil
}

// Bulk CRUD method of JSONRPCClient. Executes a DIFF method with REPLACE/UPDATE/DELETE action request against CANDIDATE datastore.
// delete/replace/update are path-value pairs. yang model type is mandatory for diff to specify: SRL or OC.
func (c *JSONRPCClient) BulkDiff(delete []PV, replace []PV, update []PV, ym yms.EnumYmType) (*Response, error) {
	// build the request
	r, err := NewDiffRequest(delete, replace, update, ym, formats.JSON, datastores.CANDIDATE)
	if err != nil {
		return nil, apierr.ClientError{
			CltFunction: "BulkDiffCandidate",
			Code:        apierr.ErrClntRPCReqCreation,
			Err:         err,
		}
	}
	return c.Do(r)
}

// Validate() action of the method SET. Executes a SET/VALIDATE specified action request against CANDIDATE datastore. Yang model type is default(SRL).
func (c *JSONRPCClient) Validate(action actions.EnumActions, pvs ...PV) (*Response, error) {
	var cmds []*Command
	for _, pv := range pvs {
		cmd, err := NewCommand(action, pv.Path, pv.Value)
		if err != nil {
			return nil, apierr.ClientError{
				CltFunction: "Validate",
				Code:        apierr.ErrClntCmdCreation,
				Err:         err,
			}
		}
		cmds = append(cmds, cmd)
	}

	r, err := NewRequest(methods.VALIDATE, cmds, WithRequestDatastore(datastores.CANDIDATE))
	if err != nil {
		return nil, apierr.ClientError{
			CltFunction: "Validate",
			Code:        apierr.ErrClntRPCReqCreation,
			Err:         err,
		}
	}
	return c.Do(r)
}

// Tools() action of the method SET. Executes a SET/UPDATE action request against TOOLS datastore. Yang model type is default(SRL).
func (c *JSONRPCClient) Tools(pvs ...PV) (*Response, error) {
	var cmds []*Command
	for _, pv := range pvs {
		cmd, err := NewCommand(actions.UPDATE, pv.Path, CommandValue(pv.Value))
		if err != nil {
			//return nil, fmt.Errorf("tools(): %w", err)
			return nil, apierr.ClientError{
				CltFunction: "Tools",
				Code:        apierr.ErrClntCmdCreation,
				Err:         err,
			}
		}
		cmds = append(cmds, cmd)
	}
	r, err := NewRequest(methods.SET, cmds, WithRequestDatastore(datastores.TOOLS))
	if err != nil {
		return nil, apierr.ClientError{
			CltFunction: "Tools",
			Code:        apierr.ErrClntRPCReqCreation,
			Err:         err,
		}
	}
	return c.Do(r)
}

// Executes CLI commands against the target device (NE).
func (c *JSONRPCClient) CLI(cmds []string, of formats.EnumOutputFormats) (*Response, error) {
	r, err := NewCLIRequest(cmds, of)
	if err != nil {
		return nil, apierr.ClientError{
			CltFunction: "CLI",
			Code:        apierr.ErrClntRPCReqCreation,
			Err:         err,
		}
	}
	return c.Do(r)
}

// Helper function to populate default values for the JSONRPCClient.
func (c *JSONRPCClient) populateDefaults() error {
	var (
		defUsername = "admin"
		defPass     = "NokiaSrl1!" // default password for SRL starting from 22.11. Should we provide "admin" permutation as well to check dynamically?
		defPort     = 443
		defTLS      = tls.Config{InsecureSkipVerify: true}
	)
	// port
	if c.target.port == nil {
		c.target.port = &defPort
	}

	// setting the timeout
	if c.target.timeout == 0 {
		c.target.timeout = 4 * time.Second
	}

	// credentials
	if c.target.username == nil {
		c.target.username = &defUsername
		c.target.password = &defPass
	}

	// ... setting the TLS configuration
	if c.target.tlsConfig == nil {
		c.target.tlsConfig = &defTLS // Skipping verification
	}
	return nil
}

// Internal function to verify the target device (NE) version and hostname, which could be used un the future to provide different behavior for different versions.
func (c *JSONRPCClient) targetVerification() error {
	// checking for the system version and hostname
	hostnameCmd, err := NewCommand(actions.NONE, "/system/name/host-name", CommandValue(""), WithDatastore(datastores.STATE))
	if err != nil {
		return apierr.ClientError{
			CltFunction: "targetVerification",
			Code:        apierr.ErrClntCmdCreation,
			Err:         err,
		}
	}
	sysVerCmd, err := NewCommand(actions.NONE, "/system/information/version", CommandValue(""), WithDatastore(datastores.STATE))
	if err != nil {
		return apierr.ClientError{
			CltFunction: "targetVerification",
			Code:        apierr.ErrClntCmdCreation,
			Err:         err,
		}
	}
	cmds := []*Command{hostnameCmd, sysVerCmd}
	r, err := NewRequest(methods.GET, cmds, nil)
	if err != nil {
		return apierr.ClientError{
			CltFunction: "targetVerification",
			Code:        apierr.ErrClntRPCReqCreation,
			Err:         err,
		}
	}

	rpcResp, err := c.Do(r)
	if err != nil {
		return err
	}

	var hostAndVer []string
	if err = json.Unmarshal(rpcResp.Result, &hostAndVer); err != nil {
		return apierr.ClientError{
			CltFunction: "targetVerification",
			Code:        apierr.ErrClntJSONUnmarshalling,
			Err:         err,
		}
	}
	c.hostname = hostAndVer[0]
	c.sysVer = hostAndVer[1]

	return nil
}

// ClientOption to update target port.
func WithOptPort(port *int) ClientOption {
	return func(c *JSONRPCClient) error {
		if port == nil {
			return apierr.ClientError{
				CltFunction: "WithOptPort",
				Code:        apierr.ErrClntNoPort,
				Err:         nil,
			}
		}
		c.target.port = port
		return nil
	}
}

// ClientOption to set connection timeout.
func WithOptTimeout(t time.Duration) ClientOption {
	return func(c *JSONRPCClient) error {
		c.target.timeout = t
		return nil
	}
}

// ClientOption to specify credentials.
func WithOptCredentials(u, p *string) ClientOption {
	return func(c *JSONRPCClient) error {
		if u == nil {
			return apierr.ClientError{
				CltFunction: "WithOptCredentials",
				Code:        apierr.ErrClntNoUsername,
				Err:         nil,
			}
		}
		c.target.username = u
		if p == nil {
			return apierr.ClientError{
				CltFunction: "WithOptCredentials",
				Code:        apierr.ErrClntNoPassword,
				Err:         nil,
			}
		}
		c.target.password = p
		return nil
	}
}

// ClientOption to specify TLS configuration.
// Setting the TLS configuration will override the default skipVerify option and will enforce the verification of the server certificate.
// Assumes minimum TLS version 1.2.
func WithOptTLS(t *TLSAttr) ClientOption {
	return func(c *JSONRPCClient) error {
		tlsConfig := tls.Config{}
		// Applying skipVerify
		tlsConfig.InsecureSkipVerify = *t.SkipVerify
		if !*t.SkipVerify {
			tlsConfig.ServerName = *c.target.host
			if len(*t.CAFile) == 0 || (!(len(*t.CertFile) == 0) && len(*t.KeyFile) == 0) || (len(*t.CertFile) == 0 && !(len(*t.KeyFile) == 0)) {
				return apierr.ClientError{
					CltFunction: "WithOptTLS",
					Code:        apierr.ErrClntTLSFilesUnspecified,
					Err:         nil,
				}
			}

			// Populating root CA certificates pool
			fh, err := os.Open(*t.CAFile)
			if err != nil {
				return apierr.ClientError{
					CltFunction: "WithOptTLS",
					Code:        apierr.ErrClntTLSFOpenCA,
					Err:         err,
				}
			}
			bs, err := ioutil.ReadAll(fh)
			if err != nil {
				return apierr.ClientError{
					CltFunction: "WithOptTLS",
					Code:        apierr.ErrClntTLSFOpenCA,
					Err:         err,
				}
			}

			certCAPool := x509.NewCertPool()
			if !certCAPool.AppendCertsFromPEM(bs) {
				return apierr.ClientError{
					CltFunction: "WithOptTLS",
					Code:        apierr.ErrClntTLSLoadCAPEM,
					Err:         nil,
				}
			}
			tlsConfig.RootCAs = certCAPool

			// Loading certificate
			certTLS, err := tls.LoadX509KeyPair(*t.CertFile, *t.KeyFile)
			if err != nil {
				return apierr.ClientError{
					CltFunction: "WithOptTLS",
					Code:        apierr.ErrClntTLSLoadCertPair,
					Err:         err,
				}
			}
			// Leaf is the parsed form of the leaf certificate, which may be initialized
			// using x509.ParseCertificate to reduce per-handshake processing.
			certTLS.Leaf, err = x509.ParseCertificate(certTLS.Certificate[0])
			if err != nil {
				return apierr.ClientError{
					CltFunction: "WithOptTLS",
					Code:        apierr.ErrClntTLSCertParsing,
					Err:         err,
				}
			}
			tlsConfig.Certificates = []tls.Certificate{certTLS}

			// Setting minimum version for TLS1.2 in accordance with specification
			tlsConfig.MinVersion = tls.VersionTLS12

		}
		c.target.tlsConfig = &tlsConfig
		return nil
	}
}
