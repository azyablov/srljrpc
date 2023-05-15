package srljrpc

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/azyablov/srljrpc/actions"
	"github.com/azyablov/srljrpc/datastores"
	"github.com/azyablov/srljrpc/formats"
	"github.com/azyablov/srljrpc/methods"
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
}

// PV type to represent a path-value pair.
type PV struct {
	Path  string       `json:"path"`
	Value CommandValue `json:"value"`
}

// ClientOption is a function type that applies options to a JSONRPCClient object.
type ClientOption func(*JSONRPCClient) error

// Creates a new JSON RPC client and applies options in order of appearance.
func NewJSONRPCClient(host *string, opts ...ClientOption) (*JSONRPCClient, error) {
	// client object
	c := &JSONRPCClient{}
	c.target = &JSONRPCTarget{}
	// host
	if host == nil {
		return nil, fmt.Errorf("host is not set, but mandatory")
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
		return nil, fmt.Errorf("target verification failed: %v", err)
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
		return nil, err
	}
	reqHTTP, err := http.NewRequest("POST", fmt.Sprintf("https://%s:%v/jsonrpc", *c.target.host, *c.target.port), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("http request creation: %s", err)
	}

	// setting content type and authentication header
	reqHTTP.Header.Set("Content-Type", "application/json")
	reqHTTP.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", *c.target.username, *c.target.password))))

	resp, err := c.client.Do(reqHTTP)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status: %s", resp.Status)
	}
	rpcResp := Response{}
	err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	if err != nil {
		return nil, fmt.Errorf("decoding error: %s", err)
	}
	if rpcResp.GetID() != r.GetID() {
		return nil, fmt.Errorf("request and response IDs do not match: %v", rpcResp.ID)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("JSON-RPC error: %+v", rpcResp.Error)
	}

	return &rpcResp, nil
}

// Get method of JSONRPCClient. Executes a GET request against RUNNING datastore.
func (c *JSONRPCClient) Get(paths ...string) (*Response, error) {
	opts := []CommandOptions{WithDatastore(datastores.RUNNING)}
	var cmds []*Command
	for _, path := range paths {
		cmd, err := NewCommand(actions.NONE, path, CommandValue(""), opts...)
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, cmd)

	}
	r, err := NewRequest(methods.GET, cmds, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

// Get state method of JSONRPCClient. Executes a GET request against STATE datastore.
func (c *JSONRPCClient) State(paths ...string) (*Response, error) {
	opts := []CommandOptions{WithDatastore(datastores.STATE)}
	var cmds []*Command
	for _, path := range paths {
		cmd, err := NewCommand(actions.NONE, path, CommandValue(""), opts...)

		if err != nil {
			return nil, err
		}
		cmds = append(cmds, cmd)
	}
	r, err := NewRequest(methods.GET, cmds, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

// SetUpdate method of JSONRPCClient. Executes a SET/UPDATE action request against CANDIDATE datastore.
func (c *JSONRPCClient) Update(pvs ...PV) (*Response, error) {
	var cmds []*Command
	for _, pv := range pvs {
		cmd, err := NewCommand(actions.UPDATE, pv.Path, CommandValue(pv.Value))
		if err != nil {
			return nil, err
		}

		cmds = append(cmds, cmd)
	}
	r, err := NewRequest(methods.SET, cmds, WithRequestDatastore(datastores.CANDIDATE))
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

// SetReplace method of JSONRPCClient. Executes a SET/REPLACE action request against CANDIDATE datastore.
func (c *JSONRPCClient) Replace(pvs ...PV) (*Response, error) {
	var cmds []*Command
	for _, pv := range pvs {
		cmd, err := NewCommand(actions.REPLACE, pv.Path, pv.Value)
		if err != nil {
			return nil, err
		}

		cmds = append(cmds, cmd)
	}
	r, err := NewRequest(methods.SET, cmds, WithRequestDatastore(datastores.CANDIDATE))
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

// SetDelete method of JSONRPCClient. Executes a SET/DELETE action request against CANDIDATE datastore.
func (c *JSONRPCClient) Delete(paths ...string) (*Response, error) {
	var cmds []*Command
	for _, path := range paths {
		cmd, err := NewCommand(actions.DELETE, path, CommandValue(""))
		if err != nil {
			return nil, err
		}

		cmds = append(cmds, cmd)
	}
	r, err := NewRequest(methods.SET, cmds, WithRequestDatastore(datastores.CANDIDATE))
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

// Validate() action of the method SET. Executes a SET/VALIDATE specified action request against CANDIDATE datastore.
func (c *JSONRPCClient) Validate(action actions.EnumActions, pvs ...PV) (*Response, error) {
	var cmds []*Command
	for _, pv := range pvs {
		cmd, err := NewCommand(action, pv.Path, pv.Value)
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, cmd)
	}

	r, err := NewRequest(methods.VALIDATE, cmds, WithRequestDatastore(datastores.CANDIDATE))
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

// Tools() action of the method SET. Executes a SET/UPDATE action request against TOOLS datastore.
func (c *JSONRPCClient) Tools(pvs ...PV) (*Response, error) {
	var cmds []*Command
	for _, pv := range pvs {
		cmd, err := NewCommand(actions.UPDATE, pv.Path, CommandValue(pv.Value))
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, cmd)
	}
	r, err := NewRequest(methods.SET, cmds, WithRequestDatastore(datastores.TOOLS))
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

// Executes CLI commands against the target device (NE).
func (c *JSONRPCClient) CLI(cmds []string, of formats.EnumOutputFormats) (*Response, error) {
	r, err := NewCLIRequest(cmds, of)
	if err != nil {
		return nil, err
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
		return err
	}
	sysVerCmd, err := NewCommand(actions.NONE, "/system/information/version", CommandValue(""), WithDatastore(datastores.STATE))
	if err != nil {
		return err
	}
	cmds := []*Command{hostnameCmd, sysVerCmd}
	r, err := NewRequest(methods.GET, cmds, nil)
	if err != nil {
		return err
	}

	rpcResp, err := c.Do(r)
	if err != nil {
		return fmt.Errorf("target verification: %s", err)
	}

	var hostAndVer []string
	if err = json.Unmarshal(rpcResp.Result, &hostAndVer); err != nil {
		return err
	}
	c.hostname = hostAndVer[0]
	c.sysVer = hostAndVer[1]

	return nil
}

// ClientOption to update target port.
func WithOptPort(port *int) ClientOption {
	return func(c *JSONRPCClient) error {
		if port == nil {
			return fmt.Errorf("port could not be nil")
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
			return fmt.Errorf("username could not be nil")
		}
		c.target.username = u
		if p == nil {
			return fmt.Errorf("password could not be nil")
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
				return fmt.Errorf("one of more files for rootCA / certificate / key are not specified")
			}

			// Populating root CA certificates pool
			fh, err := os.Open(*t.CAFile)
			if err != nil {
				return fmt.Errorf("populating root CA certificates pool: %s", err)
			}
			bs, err := ioutil.ReadAll(fh)
			if err != nil {
				return fmt.Errorf("reading root CA cert: %s", err)
			}

			certCAPool := x509.NewCertPool()
			if !certCAPool.AppendCertsFromPEM(bs) {
				return fmt.Errorf("can't load PEM file for rootCAt")
			}
			tlsConfig.RootCAs = certCAPool

			// Loading certificate
			certTLS, err := tls.LoadX509KeyPair(*t.CertFile, *t.KeyFile)
			if err != nil {
				return fmt.Errorf("can't load certificate keypair: %s", err)
			}
			// Leaf is the parsed form of the leaf certificate, which may be initialized
			// using x509.ParseCertificate to reduce per-handshake processing.
			certTLS.Leaf, err = x509.ParseCertificate(certTLS.Certificate[0])
			if err != nil {
				return fmt.Errorf("cert parsing error: %s", err)
			}
			tlsConfig.Certificates = []tls.Certificate{certTLS}

			// Setting minimum version for TLS1.2 in accordance with specification
			tlsConfig.MinVersion = tls.VersionTLS12

		}
		c.target.tlsConfig = &tlsConfig
		return nil
	}
}
