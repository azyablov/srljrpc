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
	"github.com/azyablov/srljrpc/methods"
)

type TLSAttr struct {
	RootCA     *string // CA certificate file in PEM format.
	Cert       *string // Client certificate file in PEM format.
	Key        *string // Client private key file.
	InsecConn  *bool   // Insecure connection.
	SkipVerify *bool   // Disable certificate validation during TLS session ramp-up.
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

type JSONRPCTarget struct {
	targetHost
	cred
	tlsConfig *tls.Config
}

//	class JSONRPCClient {
//		<<entity>>
//		Call(Requester r) Response
//	}
type JSONRPCClient struct {
	client   *http.Client
	hostname string
	sysVer   string
	target   *JSONRPCTarget
}

type ClientOption func(*JSONRPCClient) error

// +NewJSONRPCClient(JSONRPCTarget t) JSONRPCClient
// Creates a new JSON RPC client.
func NewJSONRPCClient(host *string, opts []ClientOption) (*JSONRPCClient, error) {

	// client
	c := &JSONRPCClient{}

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
		return nil, fmt.Errorf("got an JSON-RPC error: %v", rpcResp.Error)
	}

	return &rpcResp, nil
}

// Get method of JSONRPCClient. Executes a GET request against running datastore.
func (c *JSONRPCClient) Get(path string) (*Response, error) {
	opts := []CommandOptions{WithDatastore(datastores.RUNNING)}
	cmd, err := NewCommand(actions.NONE, path, CommandValue(""), opts...)

	if err != nil {
		return nil, err
	}
	r, err := NewRequest(methods.GET, []*Command{cmd}, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

// SetUpdate method of JSONRPCClient. Executes a SET UPDATE action request against running datastore.
func (c *JSONRPCClient) SetUpdate(path string, value CommandValue) (*Response, error) {
	opts := []CommandOptions{WithDatastore(datastores.CANDIDATE)}
	cmd, err := NewCommand(actions.UPDATE, path, value, opts...)
	if err != nil {
		return nil, err
	}
	r, err := NewRequest(methods.SET, []*Command{cmd}, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

// SetReplace method of JSONRPCClient. Executes a SET  REPLACE action request against running datastore.
func (c *JSONRPCClient) SetReplace(path string, value CommandValue) (*Response, error) {
	opts := []CommandOptions{WithDatastore(datastores.CANDIDATE)}
	cmd, err := NewCommand(actions.REPLACE, path, value, opts...)
	if err != nil {
		return nil, err
	}
	r, err := NewRequest(methods.SET, []*Command{cmd}, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

// SetDelete method of JSONRPCClient. Executes a SET DELETE action request against running datastore.
func (c *JSONRPCClient) SetDelete(path string) (*Response, error) {
	opts := []CommandOptions{WithDatastore(datastores.CANDIDATE)}
	cmd, err := NewCommand(actions.DELETE, path, CommandValue(""), opts...)
	if err != nil {
		return nil, err
	}
	r, err := NewRequest(methods.SET, []*Command{cmd}, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

// SetCreate method of JSONRPCClient. Executes a SET request against running datastore.
func (c *JSONRPCClient) Validate(action actions.EnumActions, path string, value CommandValue) (*Response, error) {
	opts := []CommandOptions{WithDatastore(datastores.CANDIDATE)}
	cmd, err := NewCommand(action, path, value, opts...)
	if err != nil {
		return nil, err
	}
	r, err := NewRequest(methods.VALIDATE, []*Command{cmd}, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

func (c *JSONRPCClient) populateDefaults() error {
	// host
	if c.target.host == nil {
		return fmt.Errorf("host is not set, but mandatory")
	}

	// port
	if c.target.port == nil {
		*c.target.port = 443
	}

	// setting the timeout
	if c.target.timeout == 0 {
		c.target.timeout = 4 * time.Second
	}

	// credentials
	if c.target.username == nil {
		*c.target.username = "admin"
		*c.target.password = "admin"
	}

	// ... setting the TLS configuration
	if c.target.tlsConfig == nil {
		c.target.tlsConfig = &tls.Config{InsecureSkipVerify: true} // Skipping verification
	}
	return nil
}

func (c *JSONRPCClient) targetVerification() error {

	// checking for the system version and hostname
	hostnameCmd, err := NewCommand(actions.NONE, "/system/name/host-name", CommandValue(""), nil)
	if err != nil {
		return err
	}
	sysVerCmd, err := NewCommand(actions.NONE, "/system/version", CommandValue(""), nil)
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

func WithOptPort(port *int) ClientOption {
	return func(c *JSONRPCClient) error {
		if port == nil {
			return fmt.Errorf("port could not be nil")
		}
		c.target.port = port
		return nil
	}
}

func WithOptTimeout(t time.Duration) ClientOption {
	return func(c *JSONRPCClient) error {
		c.target.timeout = t
		return nil
	}
}

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

func WithOptTLS(t *TLSAttr) ClientOption {
	return func(c *JSONRPCClient) error {
		tlsConfig := tls.Config{}
		// Applying skipVerify
		tlsConfig.InsecureSkipVerify = *t.SkipVerify
		if !*t.SkipVerify {
			tlsConfig.ServerName = *c.target.host
			if len(*t.RootCA) == 0 || (!(len(*t.Cert) == 0) && len(*t.Key) == 0) || (len(*t.Cert) == 0 && !(len(*t.Key) == 0)) {
				return fmt.Errorf("one of more files for rootCA / certificate / key are not specified")
			}

			// Populating root CA certificates pool
			fh, err := os.Open(*t.RootCA)
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
			certTLS, err := tls.LoadX509KeyPair(*t.Cert, *t.Key)
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
