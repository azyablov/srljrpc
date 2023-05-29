//go:build integration

package srljrpc_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/azyablov/srljrpc"
	"github.com/azyablov/srljrpc/actions"
	"github.com/azyablov/srljrpc/apierr"
	"github.com/azyablov/srljrpc/formats"
	"github.com/azyablov/srljrpc/yms"
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

type ocTarget struct {
	DefaultTarget json.RawMessage `json:"oc_target"`
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
	}{
		{testName: "Creating client with valid creds", host: &defIP.Host, opts: []srljrpc.ClientOption{srljrpc.WithOptCredentials(&defIP.Username, &defIP.Password)}, expErr: nil},                                                                                                 // should succeed
		{testName: "Creating client with valid creds and port", host: &defIP.Host, opts: []srljrpc.ClientOption{srljrpc.WithOptCredentials(&defIP.Username, &defIP.Password), srljrpc.WithOptPort(&defIP.Port)}, expErr: nil},                                                      // should succeed
		{testName: "Creating client with valid creds, port and TLS skip_verify", host: &defIP.Host, opts: []srljrpc.ClientOption{srljrpc.WithOptCredentials(&defIP.Username, &defIP.Password), srljrpc.WithOptPort(&defIP.Port), srljrpc.WithOptTLS(&defIP.TLSAttr)}, expErr: nil}, // should succeed
	}

	for _, td := range defTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := srljrpc.NewJSONRPCClient(td.host, td.opts...)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if err.Error() != td.expErr.Error() {
					t.Errorf("got: %s, while should be: %v", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf("got: %v, while should be: %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			default:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
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
	}{
		{testName: "Creating client with valid TLS inputs and skip_verify=false",
			host:   &certIP.Host,
			opts:   []srljrpc.ClientOption{srljrpc.WithOptCredentials(&certIP.Username, &certIP.Password), srljrpc.WithOptPort(&certIP.Port), srljrpc.WithOptTLS(&certIP.TLSAttr)},
			expErr: nil}, // should succeed
	}

	for _, td := range certTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := srljrpc.NewJSONRPCClient(td.host, td.opts...)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if err.Error() != td.expErr.Error() {
					t.Errorf("got: %s, while should be: %v", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf("got: %v, while should be: %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			default:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			}
		})
	}

	icat := incorrectCATarget{}
	err = json.Unmarshal(bIntParams, &icat)
	if err != nil {
		t.Fatalf("can't unmarshal %s: %v", intDataFile, err)
	}
	icaIP := intParams{}
	err = json.Unmarshal(icat.IncorrectCATarget, &icaIP)
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
			host: &icaIP.Host,
			opts: []srljrpc.ClientOption{srljrpc.WithOptCredentials(&icaIP.Username, &icaIP.Password), srljrpc.WithOptPort(&icaIP.Port), srljrpc.WithOptTLS(&icaIP.TLSAttr)},
			expErr: apierr.ClientError{
				CltFunction: "NewJSONRPCClient",
				Code:        apierr.ErrClntTargetVerification,
			}}, // should fail, CA cert is incorrect
	}
	for _, td := range icaTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := srljrpc.NewJSONRPCClient(td.host, td.opts...)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if err.Error() != td.expErr.Error() {
					t.Errorf("got: %s, while should be: %v", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf("got: %v, while should be: %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			default:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
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
		paths    []string
		expErr   error
	}{
		{testName: "Get against RUNNING datastore with default target",
			paths:  []string{"/system/json-rpc-server", "/network-instance[name=mgmt]"},
			expErr: nil,
		}, // should succeed
		{testName: "Get against RUNNING datastore with default target and invalid path",
			paths: []string{"/system/json-rpc-server/invalid"},
			expErr: apierr.ClientError{
				CltFunction: "Do",
				Code:        apierr.ErrClntJSONRPC},
		}, // should fail, invalid path
	}
	for _, td := range getTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.Get(td.paths...)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if err.Error() != td.expErr.Error() {
					t.Errorf("got: %s, while should be: %v", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf("got: %v, while should be: %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			default:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
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
		paths    []string
		expErr   error
		errMsg   string
	}{
		{testName: "Get against STATE datastore with default target",
			paths:  []string{"/system/lldp/statistics", "/network-instance[name=mgmt]/interface[name=mgmt0.0]/oper-state"},
			expErr: nil,
		}, // should succeed
		{testName: "Get with default target and invalid path",
			paths: []string{"/lldp/statistics/invalid"},
			expErr: apierr.ClientError{
				CltFunction: "Do",
				Code:        apierr.ErrClntJSONRPC},
		}, // should fail, invalid path
	}
	for _, td := range getTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.State(td.paths...)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if err.Error() != td.expErr.Error() {
					t.Errorf("got: %s, while should be: %v", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf("got: %v, while should be: %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			default:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
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
		pvs      []srljrpc.PV
		expErr   error
	}{
		{testName: "Set Update against CANDIDATE datastore with default target",
			pvs: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("test")},
				{"/interface[name=mgmt0]/description", srljrpc.CommandValue("MGMT")},
			},
			expErr: nil,
		}, // should succeed
		{testName: "Set Update against CANDIDATE datastore with default target and invalid path",
			pvs: []srljrpc.PV{
				{"/interface[name=system0]/invalid", srljrpc.CommandValue("test")}},
			expErr: apierr.ClientError{
				CltFunction: "Do",
				Code:        apierr.ErrClntJSONRPC},
		}, // should fail, invalid path
		{testName: "Set Update against CANDIDATE datastore with default target and missed value",
			pvs: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("")}},
			expErr: apierr.ClientError{
				CltFunction: "Update",
				Code:        apierr.ErrClntRPCReqCreation},
		}, // should fail, missed value
	}
	for _, td := range setTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.Update(td.pvs...)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if err.Error() != td.expErr.Error() {
					t.Errorf("got: %s, while should be: %v", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf("got: %v, while should be: %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			default:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			}
		})
	}
}

func TestBulkSetCandidate(t *testing.T) {
	// Get OC client
	c := helperGetOCClient(t)
	// Test data
	var setOCTestData = []struct {
		testName string
		delete   []srljrpc.PV
		replace  []srljrpc.PV
		update   []srljrpc.PV
		expErr   error
	}{
		{testName: "Set Update against CANDIDATE datastore with OC target",
			delete: []srljrpc.PV{
				{"/interfaces/interface[name=ethernet-1/2]/config/description", srljrpc.CommandValue("")},
			},
			replace: []srljrpc.PV{
				{"/interfaces/interface[name=ethernet-1/2]/config/name", srljrpc.CommandValue("ethernet-1/2")},
				{"/interfaces/interface[name=ethernet-1/2]/config/type", srljrpc.CommandValue("ethernetCsmacd")},
				{"/interfaces/interface[name=ethernet-1/2]/config/description", srljrpc.CommandValue("TestBulkSetCandidate_TOUPDATE")},
			},
			update: []srljrpc.PV{
				{"/interfaces/interface[name=ethernet-1/2]/config/description", srljrpc.CommandValue("TestBulkSetCandidate")},
			},
			expErr: nil}, // should succeed
		{testName: "Set Update against CANDIDATE datastore with OC target and invalid path",
			delete: []srljrpc.PV{
				{"/interfaces/interface[name=ethernet-1/2]/config/description", srljrpc.CommandValue("")},
			},
			replace: []srljrpc.PV{
				{"/interfaces/interface[name=ethernet-1/2]/config/name", srljrpc.CommandValue("ethernet-1/2")},
				{"/interfaces/interface[name=ethernet-1/2]/config/invalid", srljrpc.CommandValue("ethernetCsmacd")}, // invalid path
				{"/interfaces/interface[name=ethernet-1/2]/config/description", srljrpc.CommandValue("TestBulkSetCandidate_TOUPDATE")},
			},
			update: []srljrpc.PV{
				{"/interfaces/interface[name=ethernet-1/2]/config/description", srljrpc.CommandValue("TestBulkSetCandidate")},
			},
			expErr: apierr.ClientError{
				CltFunction: "Do",
				Code:        apierr.ErrClntJSONRPC}}, // should fail, invalid path
		{testName: "Set Update against CANDIDATE datastore with OC target and missed value",
			delete: []srljrpc.PV{
				{"/interfaces/interface[name=ethernet-1/2]/config/description", srljrpc.CommandValue("")},
			},
			replace: []srljrpc.PV{
				{"/interfaces/interface[name=ethernet-1/2]/config/name", srljrpc.CommandValue("ethernet-1/2")},
				{"/interfaces/interface[name=ethernet-1/2]/config/type", srljrpc.CommandValue("")},
				{"/interfaces/interface[name=ethernet-1/2]/config/description", srljrpc.CommandValue("TestBulkSetCandidate_TOUPDATE")},
			},
			update: []srljrpc.PV{
				{"/interfaces/interface[name=ethernet-1/2]/config/description", srljrpc.CommandValue("TestBulkSetCandidate")},
			},
			expErr: apierr.ClientError{
				CltFunction: "BulkSetCandidate",
				Code:        apierr.ErrClntRPCReqCreation,
			}}, // should fail, missed value
	}
	for _, td := range setOCTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.BulkSetCandidate(td.delete, td.replace, td.update, yms.OC)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if err.Error() != td.expErr.Error() {
					t.Errorf("got: %s, while should be: %v", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf("got: %v, while should be: %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			default:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			}
		})
	}

	// Get with default client
	c = helperGetDefClient(t)
	// Test data
	var setTestData = []struct {
		testName string
		delete   []srljrpc.PV
		replace  []srljrpc.PV
		update   []srljrpc.PV
		expErr   error
	}{
		{testName: "Set Update against CANDIDATE datastore with SRL default target",
			delete: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("")},
			},
			replace: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("System")},
				{"/interface[name=mgmt0]/description", srljrpc.CommandValue("MGMT")},
			},
			update: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("System loopback")},
			},
			expErr: nil}, // should succeed
		{testName: "Set Update against CANDIDATE datastore with SRL default target and invalid path",
			delete: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("")},
			},
			replace: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("System")},
				{"/interface[name=mgmt0]/invalid", srljrpc.CommandValue("MGMT")},
			},
			update: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("System loopback")},
			},
			expErr: apierr.ClientError{
				CltFunction: "Do",
				Code:        apierr.ErrClntJSONRPC}}, // should fail, invalid path
		{testName: "Set Update against CANDIDATE datastore with SRL default target and missed value",
			delete: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("")},
			},
			replace: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("System")},
				{"/interface[name=mgmt0]/description", srljrpc.CommandValue("")},
			},
			update: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("System loopback")},
			},
			expErr: apierr.ClientError{
				CltFunction: "BulkSetCandidate",
				Code:        apierr.ErrClntRPCReqCreation,
			}}, // should fail, missed value
	}
	for _, td := range setTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.BulkSetCandidate(td.delete, td.replace, td.update, yms.SRL)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if err.Error() != td.expErr.Error() {
					t.Errorf("got: %s, while should be: %v", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf("got: %v, while should be: %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			default:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			}
		})
	}

}

func TestBulkDiffCandidate(t *testing.T) {
	// Get OC client
	c := helperGetOCClient(t)
	// Test data
	var setOCTestData = []struct {
		testName string
		delete   []srljrpc.PV
		replace  []srljrpc.PV
		update   []srljrpc.PV
		expErr   error
	}{
		{testName: "Set Update against CANDIDATE datastore with OC target",
			delete: []srljrpc.PV{
				{"/interfaces/interface[name=ethernet-1/2]/config/description", srljrpc.CommandValue("")},
			},
			replace: []srljrpc.PV{
				{"/interfaces/interface[name=ethernet-1/2]/config/name", srljrpc.CommandValue("ethernet-1/2")},
				{"/interfaces/interface[name=ethernet-1/2]/config/type", srljrpc.CommandValue("ethernetCsmacd")},
				{"/interfaces/interface[name=ethernet-1/2]/config/description", srljrpc.CommandValue("TestBulkSetCandidate_TOUPDATE")},
			},
			update: []srljrpc.PV{
				{"/interfaces/interface[name=ethernet-1/2]/config/description", srljrpc.CommandValue("TestBulkSetCandidate")},
			},
			expErr: nil}, // should succeed
		{testName: "Set Update against CANDIDATE datastore with OC target and invalid path",
			delete: []srljrpc.PV{
				{"/interfaces/interface[name=ethernet-1/2]/config/description", srljrpc.CommandValue("")},
			},
			replace: []srljrpc.PV{
				{"/interfaces/interface[name=ethernet-1/2]/config/name", srljrpc.CommandValue("ethernet-1/2")},
				{"/interfaces/interface[name=ethernet-1/2]/config/invalid", srljrpc.CommandValue("ethernetCsmacd")}, // invalid path
				{"/interfaces/interface[name=ethernet-1/2]/config/description", srljrpc.CommandValue("TestBulkSetCandidate_TOUPDATE")},
			},
			update: []srljrpc.PV{
				{"/interfaces/interface[name=ethernet-1/2]/config/description", srljrpc.CommandValue("TestBulkSetCandidate")},
			},
			expErr: apierr.ClientError{
				CltFunction: "Do",
				Code:        apierr.ErrClntJSONRPC}}, // should fail, invalid path
		{testName: "Set Update against CANDIDATE datastore with OC target and missed value",
			delete: []srljrpc.PV{
				{"/interfaces/interface[name=ethernet-1/2]/config/description", srljrpc.CommandValue("")},
			},
			replace: []srljrpc.PV{
				{"/interfaces/interface[name=ethernet-1/2]/config/name", srljrpc.CommandValue("ethernet-1/2")},
				{"/interfaces/interface[name=ethernet-1/2]/config/type", srljrpc.CommandValue("")},
				{"/interfaces/interface[name=ethernet-1/2]/config/description", srljrpc.CommandValue("TestBulkSetCandidate_TOUPDATE")},
			},
			update: []srljrpc.PV{
				{"/interfaces/interface[name=ethernet-1/2]/config/description", srljrpc.CommandValue("TestBulkSetCandidate")},
			},
			expErr: apierr.ClientError{
				CltFunction: "BulkDiffCandidate",
				Code:        apierr.ErrClntRPCReqCreation,
			}}, // should fail, missed value
	}
	for _, td := range setOCTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.BulkDiffCandidate(td.delete, td.replace, td.update, yms.OC)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if err.Error() != td.expErr.Error() {
					t.Errorf("got: %s, while should be: %v", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf("got: %v, while should be: %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			default:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			}
		})
	}

	// Get with default client
	c = helperGetDefClient(t)
	// Test data
	var setTestData = []struct {
		testName string
		delete   []srljrpc.PV
		replace  []srljrpc.PV
		update   []srljrpc.PV
		expErr   error
	}{
		{testName: "Set Update against CANDIDATE datastore with SRL default target",
			delete: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("")},
			},
			replace: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("System")},
				{"/interface[name=mgmt0]/description", srljrpc.CommandValue("MGMT")},
			},
			update: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("System loopback")},
			},
			expErr: nil}, // should succeed
		{testName: "Set Update against CANDIDATE datastore with SRL default target and invalid path",
			delete: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("")},
			},
			replace: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("System")},
				{"/interface[name=mgmt0]/invalid", srljrpc.CommandValue("MGMT")},
			},
			update: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("System loopback")},
			},
			expErr: apierr.ClientError{
				CltFunction: "Do",
				Code:        apierr.ErrClntJSONRPC}}, // should fail, invalid path
		{testName: "Set Update against CANDIDATE datastore with SRL default target and missed value",
			delete: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("")},
			},
			replace: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("System")},
				{"/interface[name=mgmt0]/description", srljrpc.CommandValue("")},
			},
			update: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("System loopback")},
			},
			expErr: apierr.ClientError{
				CltFunction: "BulkDiffCandidate",
				Code:        apierr.ErrClntRPCReqCreation,
			}}, // should fail, missed value
	}
	for _, td := range setTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.BulkDiffCandidate(td.delete, td.replace, td.update, yms.SRL)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if err.Error() != td.expErr.Error() {
					t.Errorf("got: %s, while should be: %v", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf("got: %v, while should be: %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			default:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
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
		pvs      []srljrpc.PV
		expErr   error
		errMsg   string
	}{
		{testName: "Set Replace against CANDIDATE datastore with default target",
			pvs:    []srljrpc.PV{{"/interface[name=system0]/description:test", srljrpc.CommandValue("")}},
			expErr: nil,
		}, // should succeed
		{testName: "Set Replace against CANDIDATE datastore with default target and invalid path",
			pvs: []srljrpc.PV{{"/interface[name=system0]/invalid:test", srljrpc.CommandValue("")}},
			expErr: apierr.ClientError{
				CltFunction: "Do",
				Code:        apierr.ErrClntJSONRPC},
		}, // should fail, invalid path
	}
	for _, td := range getTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.Replace(td.pvs...)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if err.Error() != td.expErr.Error() {
					t.Errorf("got: %s, while should be: %v", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf("got: %v, while should be: %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			default:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
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
		paths    []string
		expErr   error
		errMsg   string
	}{
		{testName: "Delete against CANDIDATE datastore with default target",
			paths:  []string{"/interface[name=system0]/description", "/interface[name=mgmt0]/description"},
			expErr: nil,
		}, // should succeed
		{testName: "Delete against CANDIDATE datastore with default target and invalid path",
			paths: []string{"/interface[name=system0]/invalid"},
			expErr: apierr.ClientError{
				CltFunction: "Do",
				Code:        apierr.ErrClntJSONRPC},
		}, // should fail, invalid path
	}
	for _, td := range setTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.Delete(td.paths...)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if err.Error() != td.expErr.Error() {
					t.Errorf("got: %s, while should be: %v", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf("got: %v, while should be: %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			default:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
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
		pvs      []srljrpc.PV
		expErr   error
	}{
		{testName: "Validate against CANDIDATE datastore with default target",
			pvs: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("test")},
				{"/interface[name=mgmt0]/description", srljrpc.CommandValue("MGMT")},
			},
			expErr: nil,
		}, // should succeed
		{testName: "Validate against CANDIDATE datastore with default target and invalid path",
			pvs: []srljrpc.PV{
				{"/interface[name=system0]/invalid", srljrpc.CommandValue("test")}},
			expErr: apierr.ClientError{
				CltFunction: "Do",
				Code:        apierr.ErrClntJSONRPC},
		}, // should fail, invalid path
		{testName: "Validate against CANDIDATE datastore with default target and missed value",
			pvs: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("")}},
			expErr: apierr.ClientError{
				CltFunction: "Validate",
				Code:        apierr.ErrClntRPCReqCreation},
		}, // should fail, missed value
	}

	for _, td := range validateTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.Validate(actions.UPDATE, td.pvs...)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if err.Error() != td.expErr.Error() {
					t.Errorf("got: %s, while should be: %v", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf("got: %v, while should be: %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			default:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			}
		})
	}
}

func TestTools(t *testing.T) {
	// Get default client
	c := helperGetDefClient(t)
	// Test data
	var toolsTestData = []struct {
		testName string
		pvs      []srljrpc.PV
		expErr   error
	}{
		{testName: "Set against TOOLS w/o value",
			pvs: []srljrpc.PV{
				{"/interface[name=ethernet-1/1]/ethernet/statistics/clear", srljrpc.CommandValue("")},
			},
			expErr: nil,
		}, // should succeed
		{testName: "Set against TOOLS and invalid path",
			pvs: []srljrpc.PV{
				{"/interface[name=ethernet-1/1]/ethernet/INVALID/clear", srljrpc.CommandValue("")},
			},
			expErr: apierr.ClientError{
				CltFunction: "Do",
				Code:        apierr.ErrClntJSONRPC},
		}, // should fail, invalid path
		{testName: "Set against TOOLS with double value",
			pvs: []srljrpc.PV{
				{"/network-instance[name=default]/protocols/bgp/group[group-name=underlay]/soft-clear/peer-as:65020", srljrpc.CommandValue("65020")}},
			expErr: apierr.ClientError{
				CltFunction: "Tools",
				Code:        apierr.ErrClntRPCReqCreation},
		}, // should fail, value specified two times
	}

	for _, td := range toolsTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.Tools(td.pvs...)
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if err.Error() != td.expErr.Error() {
					t.Errorf("got: %s, while should be: %v", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf("got: %v, while should be: %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			default:
				t.Errorf("got: %s, while should be: %s", err, td.expErr)
			}
		})
	}
}

func TestCLI(t *testing.T) {
	// Get default client
	c := helperGetDefClient(t)

	// CLI with default target using Do()
	shVerTABLE, err := srljrpc.NewCLIRequest([]string{"show version"}, formats.TABLE)
	if err != nil {
		t.Fatalf("can't create CLI request: %v", err)
	}

	shRouteTableJSON, err := srljrpc.NewCLIRequest([]string{"show network-instance default route-table"}, formats.JSON)
	if err != nil {
		t.Fatalf("can't create CLI request: %v", err)
	}
	// CLI Do() with default target
	var cliDoTestData = []struct {
		testName string
		cliReq   *srljrpc.CLIRequest
		expErr   error
		errMsg   string
	}{
		{testName: "CLI show version via Do()",
			cliReq: shVerTABLE,
			expErr: nil, errMsg: "cli method failed: "},
		{testName: "CLI show network-instance default route-table via Do()",
			cliReq: shRouteTableJSON,
			expErr: nil, errMsg: "cli Do() method failed: "},
	}
	for _, td := range cliDoTestData {
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
		var cliTestData = []struct {
			testName string
			cmds     []string
			of       formats.EnumOutputFormats
			expErr   error
		}{
			{testName: "CLI bulk via CLI() in TABLE format",
				cmds:   []string{"show version", "show network-instance default route-table", "show acl summary"},
				of:     formats.TABLE,
				expErr: nil,
			}, // should succeed
			{testName: "CLI bulk via CLI() in JSON format",
				cmds:   []string{"show version", "show network-instance default route-table", "show acl summary"},
				of:     formats.JSON,
				expErr: nil,
			}, // should succeed
			{testName: "CLI bulk via CLI() with empty commands",
				cmds: []string{"show version", "", "show acl summary"},
				of:   formats.TEXT,
				expErr: apierr.ClientError{
					CltFunction: "CLI",
					Code:        apierr.ErrClntRPCReqCreation},
			}, // should fail, empty command
		}
		// CLI with default target using CLI()
		for _, td := range cliTestData {
			t.Run(td.testName, func(t *testing.T) {
				r, err := c.CLI(td.cmds, td.of)
				switch {
				case err == nil && td.expErr == nil:
				case err != nil && td.expErr != nil:
					if err.Error() != td.expErr.Error() {
						t.Errorf("got: %s, while should be: %v", err, td.expErr)
					}
				case err == nil && td.expErr != nil:
					t.Errorf("got: %v, while should be: %s", err, td.expErr)
				case err != nil && td.expErr == nil:
					t.Errorf("got: %s, while should be: %s", err, td.expErr)
				default:
					t.Errorf("got: %s, while should be: %s", err, td.expErr)
				}
				_, err = r.Marshal()
				if err != nil {
					t.Fatalf("can't marshal response: %v", err)
				}
				// for debug purposes
				//t.Logf("got response: %+v", string(b))
			})
		}

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

func helperGetOCClient(t *testing.T) *srljrpc.JSONRPCClient {
	// Read integration tests parameters
	dt := ocTarget{}
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
	ocIP := intParams{}
	err = json.Unmarshal(dt.DefaultTarget, &ocIP)
	if err != nil {
		t.Fatalf("can't unmarshal %s: %v", intDataFile, err)
	}

	c, err := srljrpc.NewJSONRPCClient(&ocIP.Host, srljrpc.WithOptCredentials(&ocIP.Username, &ocIP.Password), srljrpc.WithOptPort(&ocIP.Port))
	if err != nil {
		t.Fatalf("can't create client: %v", err)
	}
	return c
}
