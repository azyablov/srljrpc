package srljrpc_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/azyablov/srljrpc"
	"github.com/azyablov/srljrpc/actions"
	"github.com/azyablov/srljrpc/formats"
)

const (
	intDataFile = "testdata/integration_tests_params.json"
)

type intParams struct {
	Host     string          `json:"host"`
	Username string          `json:"username,omitempty"`
	Password string          `json:"password,omitempty"`
	Port     int             `json:"port,omitempty"`
	TLSAttr  srljrpc.TLSAttr `json:"tls_attr,omitempty"`
}

type defTarget struct {
	DefaultTarget json.RawMessage `json:"default_target"`
}

type certTarget struct {
	CertTarget json.RawMessage `json:"cert_target"`
}

type incorrectCATarget struct {
	IncorrectCATarget json.RawMessage `json:"cert_target_incorrect_ca"`
}

func TestNewJSONRPCClient(t *testing.T) {
	// Read integration tests parameters
	dt := defTarget{}
	fh, err := os.Open(intDataFile)
	if err != nil {
		t.Fatalf("can't open %s: %v", intDataFile, err)
	}
	defer fh.Close()
	bIntParams, err := ioutil.ReadAll(fh)
	if err != nil {
		t.Fatalf("can't read %s: %v", intDataFile, err)
	}
	err = json.Unmarshal(bIntParams, &dt)
	if err != nil {
		t.Fatalf("can't unmarshal %s: %v", intDataFile, err)
	}
	defIP := intParams{}
	json.Unmarshal(dt.DefaultTarget, &defIP)

	// Table driven tests to check NewJSONRPCClient() function
	var defTestData = []struct {
		testName string
		host     *string
		opts     []srljrpc.ClientOption
		expErr   error
		errMsg   string
	}{
		{testName: "Creating client with valid creds", host: &defIP.Host, opts: []srljrpc.ClientOption{srljrpc.WithOptCredentials(&defIP.Username, &defIP.Password)}, expErr: nil, errMsg: "client with valid host isn't created: "},
		{testName: "Creating client with valid creds and port", host: &defIP.Host, opts: []srljrpc.ClientOption{srljrpc.WithOptCredentials(&defIP.Username, &defIP.Password), srljrpc.WithOptPort(&defIP.Port)}, expErr: nil, errMsg: "client with valid host isn't created: "},
		{testName: "Creating client with valid creds, port and TLS skip_verify", host: &defIP.Host, opts: []srljrpc.ClientOption{srljrpc.WithOptCredentials(&defIP.Username, &defIP.Password), srljrpc.WithOptPort(&defIP.Port), srljrpc.WithOptTLS(&defIP.TLSAttr)}, expErr: nil, errMsg: "client with valid host isn't created: "},
	}

	for _, td := range defTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := srljrpc.NewJSONRPCClient(td.host, td.opts...)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if err.Error() != td.expErr.Error() {
					t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			default:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			}
		})
	}

	ct := certTarget{}
	err = json.Unmarshal(bIntParams, &ct)
	if err != nil {
		t.Fatalf("can't unmarshal %s: %v", intDataFile, err)
	}
	certIP := intParams{}
	err = json.Unmarshal(ct.CertTarget, &certIP)
	if err != nil {
		t.Fatalf("can't unmarshal %s: %v", intDataFile, err)
	}

	// With all TLS options
	var certTestData = []struct {
		testName string
		host     *string
		opts     []srljrpc.ClientOption
		expErr   error
		errMsg   string
	}{
		{testName: "Creating client with valid TLS inputs and skip_verify=false",
			host:   &certIP.Host,
			opts:   []srljrpc.ClientOption{srljrpc.WithOptCredentials(&certIP.Username, &certIP.Password), srljrpc.WithOptPort(&certIP.Port), srljrpc.WithOptTLS(&certIP.TLSAttr)},
			expErr: nil, errMsg: "client with valid TLS inputs isn't created: "},
	}
	for _, td := range certTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := srljrpc.NewJSONRPCClient(td.host, td.opts...)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if err.Error() != td.expErr.Error() {
					t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			default:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			}
		})
	}

	icat := incorrectCATarget{}
	err = json.Unmarshal(bIntParams, &icat)
	if err != nil {
		t.Fatalf("can't unmarshal %s: %v", intDataFile, err)
	}
	icaIP := intParams{}
	err = json.Unmarshal(ct.CertTarget, &icaIP)
	if err != nil {
		t.Fatalf("can't unmarshal %s: %v", intDataFile, err)
	}

	// Incorrect CA certificate
	var icaTestData = []struct {
		testName string
		host     *string
		opts     []srljrpc.ClientOption
		expErr   error
		errMsg   string
	}{
		{testName: "Creating client with valid TLS inputs but ca_cert is incorrect",
			host:   &icaIP.Host,
			opts:   []srljrpc.ClientOption{srljrpc.WithOptCredentials(&icaIP.Username, &icaIP.Password), srljrpc.WithOptPort(&icaIP.Port), srljrpc.WithOptTLS(&icaIP.TLSAttr)},
			expErr: nil, errMsg: "error expected: "},
	}
	for _, td := range icaTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := srljrpc.NewJSONRPCClient(td.host, td.opts...)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if err.Error() != td.expErr.Error() {
					t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			default:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			}
		})
	}

}

func TestGet(t *testing.T) {
	// Get default client
	c := helperGetDefClient(t)

	// Get with default target
	var getTestData = []struct {
		testName string
		path     string
		expErr   error
		errMsg   string
	}{
		{testName: "Get against RUNNING datastore with default target",
			path:   "/system/json-rpc-server",
			expErr: nil, errMsg: "GET method failed: "},
		{testName: "Get against RUNNING datastore with default target and invalid path",
			path:   "/system/json-rpc-server/invalid",
			expErr: fmt.Errorf("JSON-RPC error:"), errMsg: "expect JSON-RPC error: "},
	}
	for _, td := range getTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.Get(td.path)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if !strings.Contains(err.Error(), td.expErr.Error()) {
					t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			default:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			}
		})
	}
}

func TestState(t *testing.T) {
	// Get default client
	c := helperGetDefClient(t)

	// State with default target
	var getTestData = []struct {
		testName string
		path     string
		expErr   error
		errMsg   string
	}{
		{testName: "Get against STATE datastore with default target",
			path:   "/system/lldp/statistics",
			expErr: nil, errMsg: "get method failed: "},
		{testName: "Get with default target and invalid path",
			path:   "/lldp/statistics/invalid",
			expErr: fmt.Errorf("JSON-RPC error:"), errMsg: "expect JSON-RPC error: "},
	}
	for _, td := range getTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.State(td.path)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if !strings.Contains(err.Error(), td.expErr.Error()) {
					t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			default:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	// Get default client
	c := helperGetDefClient(t)

	// SetUpdate with default target
	var setTestData = []struct {
		testName string
		path     string
		value    srljrpc.CommandValue
		expErr   error
		errMsg   string
	}{
		{testName: "Set Update against CANDIDATE datastore with default target",
			path:   "/interface[name=system0]/description",
			value:  srljrpc.CommandValue("test"),
			expErr: nil, errMsg: "set update method failed: "},
		{testName: "Set Update against CANDIDATE datastore with default target and invalid path",
			path:   "/interface[name=system0]/invalid",
			value:  srljrpc.CommandValue("test"),
			expErr: fmt.Errorf("JSON-RPC error:"), errMsg: "expect JSON-RPC error: "},
		{testName: "Set Update against CANDIDATE datastore with default target and missed value",
			path:   "/interface[name=system0]/description",
			value:  srljrpc.CommandValue(""),
			expErr: fmt.Errorf("value isn't specified or not found in the path for method set"),
			errMsg: "expect value not specified error: "},
	}
	for _, td := range setTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.Update(td.path, td.value)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if !strings.Contains(err.Error(), td.expErr.Error()) {
					t.Errorf(td.errMsg+"got %+s, while should be %s", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf(td.errMsg+"got %+s, while should be %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf(td.errMsg+"got %+s, while should be %s", err, td.expErr)
			default:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			}
		})
	}
}

func TestReplace(t *testing.T) {
	// Get default client
	c := helperGetDefClient(t)

	// SetReplace with default target
	var getTestData = []struct {
		testName string
		path     string
		expErr   error
		errMsg   string
	}{
		{testName: "Set Replace against CANDIDATE datastore with default target",
			path:   "/interface[name=system0]/description:test",
			expErr: nil, errMsg: "set replace method failed: "},
		{testName: "Set Replace against CANDIDATE datastore with default target and invalid path",
			path:   "/interface[name=system0]/invalid:test",
			expErr: fmt.Errorf("JSON-RPC error:"), errMsg: "expect JSON-RPC error: "},
	}
	for _, td := range getTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.Replace(td.path, srljrpc.CommandValue(""))
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if !strings.Contains(err.Error(), td.expErr.Error()) {
					t.Errorf(td.errMsg+"got %+s, while should be %s", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf(td.errMsg+"got %+s, while should be %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf(td.errMsg+"got %+s, while should be %s", err, td.expErr)
			default:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	// Get default client
	c := helperGetDefClient(t)

	// Delete with default target
	var setTestData = []struct {
		testName string
		path     string
		expErr   error
		errMsg   string
	}{
		{testName: "Delete against CANDIDATE datastore with default target",
			path:   "/interface[name=system0]/description",
			expErr: nil, errMsg: "delete method failed: "},
		{testName: "Delete against CANDIDATE datastore with default target and invalid path",
			path:   "/interface[name=system0]/invalid",
			expErr: fmt.Errorf("JSON-RPC error:"), errMsg: "expect JSON-RPC error: "},
	}
	for _, td := range setTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.Delete(td.path)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if !strings.Contains(err.Error(), td.expErr.Error()) {
					t.Errorf(td.errMsg+"got %+s, while should be %s", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf(td.errMsg+"got %+s, while should be %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf(td.errMsg+"got %+s, while should be %s", err, td.expErr)
			default:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	// Get default client
	c := helperGetDefClient(t)

	// SetUpdate with default target
	var validateTestData = []struct {
		testName string
		path     string
		value    srljrpc.CommandValue
		expErr   error
		errMsg   string
	}{
		{testName: "Validate against CANDIDATE datastore with default target",
			path:   "/interface[name=system0]/description",
			value:  srljrpc.CommandValue("test"),
			expErr: nil, errMsg: "Set Update method failed: "},
		{testName: "Validate against CANDIDATE datastore with default target and invalid path",
			path:   "/interface[name=system0]/invalid",
			value:  srljrpc.CommandValue("test"),
			expErr: fmt.Errorf("JSON-RPC error:"), errMsg: "expect JSON-RPC error: "},
	}
	for _, td := range validateTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.Validate(actions.UPDATE, td.path, td.value)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if !strings.Contains(err.Error(), td.expErr.Error()) {
					t.Errorf(td.errMsg+"got %+s, while should be %s", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf(td.errMsg+"got %+s, while should be %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf(td.errMsg+"got %+s, while should be %s", err, td.expErr)
			default:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			}
		})
	}
}

func TestCLI(t *testing.T) {
	// Get default client
	c := helperGetDefClient(t)

	// CLI with default target
	shVerTABLE, err := srljrpc.NewCLIRequest([]string{"show version"}, formats.TABLE)
	if err != nil {
		t.Fatalf("can't create CLI request: %v", err)
	}

	shRouteTableJSON, err := srljrpc.NewCLIRequest([]string{"show network-instance default route-table"}, formats.JSON)
	if err != nil {
		t.Fatalf("can't create CLI request: %v", err)
	}
	// SetUpdate with default target
	var validateTestData = []struct {
		testName string
		cliReq   *srljrpc.CLIRequest
		expErr   error
		errMsg   string
	}{
		{testName: "CLI show version",
			cliReq: shVerTABLE,
			expErr: nil, errMsg: "cli method failed: "},
		{testName: "CLI show network-instance default route-table",
			cliReq: shRouteTableJSON,
			expErr: nil, errMsg: "cli method failed: "},
	}
	for _, td := range validateTestData {
		t.Run(td.testName, func(t *testing.T) {
			r, err := c.Do(td.cliReq)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if !strings.Contains(err.Error(), td.expErr.Error()) {
					t.Errorf(td.errMsg+"got %+s, while should be %s", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf(td.errMsg+"got %+s, while should be %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf(td.errMsg+"got %+s, while should be %s", err, td.expErr)
			default:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErr)
			}
			_, err = r.Marshal()
			if err != nil {
				t.Fatalf("can't marshal response: %v", err)
			}
			// for debug purposes
			// t.Logf("got response: %+v", string(b))
		})
	}

}

func helperGetDefClient(t *testing.T) *srljrpc.JSONRPCClient {
	// Read integration tests parameters
	dt := defTarget{}
	fh, err := os.Open(intDataFile)
	if err != nil {
		t.Fatalf("can't open %s: %v", intDataFile, err)
	}
	defer fh.Close()
	bIntParams, err := ioutil.ReadAll(fh)
	if err != nil {
		t.Fatalf("can't read %s: %v", intDataFile, err)
	}
	err = json.Unmarshal(bIntParams, &dt)
	if err != nil {
		t.Fatalf("can't unmarshal %s: %v", intDataFile, err)
	}
	defIP := intParams{}
	err = json.Unmarshal(dt.DefaultTarget, &defIP)
	if err != nil {
		t.Fatalf("can't unmarshal %s: %v", intDataFile, err)
	}

	c, err := srljrpc.NewJSONRPCClient(&defIP.Host, srljrpc.WithOptCredentials(&defIP.Username, &defIP.Password), srljrpc.WithOptPort(&defIP.Port))
	if err != nil {
		t.Fatalf("can't create client: %v", err)
	}
	return c
}
