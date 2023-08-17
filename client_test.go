//go:build integration

package srljrpc_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"

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
			checkErrGotVSExp(err, td.expErr, t)
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
			checkErrGotVSExp(err, td.expErr, t)
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
			host:   &icaIP.Host,
			opts:   []srljrpc.ClientOption{srljrpc.WithOptCredentials(&icaIP.Username, &icaIP.Password), srljrpc.WithOptPort(&icaIP.Port), srljrpc.WithOptTLS(&icaIP.TLSAttr)},
			expErr: apierr.ErrClntTargetVerification}, // should fail, CA cert is incorrect
	}
	for _, td := range icaTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := srljrpc.NewJSONRPCClient(td.host, td.opts...)
			checkErrGotVSExp(err, td.expErr, t)
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
			paths:  []string{"/system/json-rpc-server/invalid"},
			expErr: apierr.ErrClntJSONRPCResp,
		}, // should fail, invalid path
	}
	for _, td := range getTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.Get(td.paths...)
			checkErrGotVSExp(err, td.expErr, t)
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
			paths:  []string{"/lldp/statistics/invalid"},
			expErr: apierr.ErrClntJSONRPCResp,
		}, // should fail, invalid path
	}
	for _, td := range getTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.State(td.paths...)
			checkErrGotVSExp(err, td.expErr, t)
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
		ct       int
		expErr   error
	}{
		{testName: "Set Update against CANDIDATE datastore with default target",
			pvs: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("test")},
				{"/interface[name=mgmt0]/description", srljrpc.CommandValue("MGMT")},
			},
			ct:     0,
			expErr: nil,
		}, // should succeed
		{testName: "Set Update against CANDIDATE datastore with default target and invalid path",
			pvs: []srljrpc.PV{
				{"/interface[name=system0]/invalid", srljrpc.CommandValue("test")}},
			ct:     0,
			expErr: apierr.ErrClntJSONRPCResp,
		}, // should fail, invalid path
		{testName: "Set Update against CANDIDATE datastore with default target and missed value",
			pvs: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("")}},
			ct:     0,
			expErr: apierr.ErrClntRPCReqCreation,
		}, // should fail, missed value
		{testName: "Set Update against CANDIDATE datastore with default target and confirm timeout",
			pvs: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("test_CT_22")}, // after timeout should be "test", as per test 0.
				{"/interface[name=mgmt0]/description", srljrpc.CommandValue("MGMT_CT_22")},   // after timeout should be "MGMT", as per test 0.
			},
			ct:     5,
			expErr: nil,
		}, // should succeed with w/ confirm timeout
	}
	for n, td := range setTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.Update(td.ct, td.pvs...)
			checkErrGotVSExp(err, td.expErr, t)
			if n == 3 { // test with confirm timeout
				t.Logf("Set Update against CANDIDATE datastore with default target and confirm timeout => Waiting for 1 seconds...")
				time.Sleep(1 * time.Second)
				ctResp, err := c.Get("/interface[name=system0]/description", "/interface[name=mgmt0]/description")
				if err != nil {
					t.Fatal(err)
				}
				// Unmarshal response
				var ctRespData []string
				err = json.Unmarshal(ctResp.Result, &ctRespData)
				if err != nil {
					t.Fatal(err)
				}
				// Check if values are updated
				if ctRespData[0] != "test_CT_22" {
					t.Errorf("got: %s, while should be: %s", ctRespData[0], "test_CT_22")
				}
				if ctRespData[1] != "MGMT_CT_22" {
					t.Errorf("got: %s, while should be: %s", ctRespData[1], "MGMT_CT_22")
				}
				// t.Log(ctRespData)
				t.Logf("Set Update against CANDIDATE datastore with default target and confirm timeout => Waiting for 5 seconds...")
				time.Sleep(5 * time.Second)
				ctResp, err = c.Get("/interface[name=system0]/description", "/interface[name=mgmt0]/description")
				if err != nil {
					t.Fatal(err)
				}
				// Unmarshal response
				err = json.Unmarshal(ctResp.Result, &ctRespData)
				if err != nil {
					t.Fatal(err)
				}
				// Check if values are updated
				if ctRespData[0] != "test" {
					t.Errorf("got: %s, while should be: %s", ctRespData[0], "test")
				}
				if ctRespData[1] != "MGMT" {
					t.Errorf("got: %s, while should be: %s", ctRespData[1], "MGMT")
				}
				// t.Log(ctRespData)
			}
		})
	}

}

func TestBulkSet(t *testing.T) {
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
			expErr: apierr.ErrClntJSONRPCResp, // should fail, invalid path
		},
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
			expErr: apierr.ErrClntRPCReqCreation, // should fail, missed value
		},
	}
	for _, td := range setOCTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.BulkSet(td.delete, td.replace, td.update, yms.OC, 0) // ct set to 0 to avoid confirm timeout
			checkErrGotVSExp(err, td.expErr, t)
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
			expErr: apierr.ErrClntJSONRPCResp}, // should fail, invalid path
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
			expErr: apierr.ErrClntRPCReqCreation}, // should fail, missed value
	}
	for _, td := range setTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.BulkSet(td.delete, td.replace, td.update, yms.SRL, 0) // ct set to 0 to avoid confirm timeout
			checkErrGotVSExp(err, td.expErr, t)
		})
	}

}

func TestBulkSetCallBack(t *testing.T) {
	// Get SRL client
	c := helperGetDefClient(t)
	// Test data
	var setTestData = []struct {
		testName string
		delete   []srljrpc.PV
		replace  []srljrpc.PV
		update   []srljrpc.PV
		expErr   error
		cbf      srljrpc.CallBackConfirm
	}{
		{testName: "Set Update w/ confirm timeout and CallBackConfirm true against CANDIDATE datastore with SRL default target",
			delete:  []srljrpc.PV{},
			replace: []srljrpc.PV{},
			update: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("System loopback")},
			},
			expErr: nil,
			cbf: func(req *srljrpc.Request, resp *srljrpc.Response) (bool, error) {
				// For debug purposes only
				// b, err := json.Marshal(req)
				// if err != nil {
				// 	t.Fatal(err)
				// }
				// t.Log(string(b))
				// b, err = json.Marshal(resp)
				// if err != nil {
				// 	t.Fatal(err)
				// }
				// t.Log(string(b))
				return true, nil
			}}, // should succeed w/ commit confirmed via tools
		{testName: "Set Replace w/ confirm timeout and CallBackConfirm false against CANDIDATE datastore with SRL default target",
			delete: []srljrpc.PV{},
			replace: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("System loopback")},
			},
			update: []srljrpc.PV{},
			expErr: nil,
			cbf: func(req *srljrpc.Request, resp *srljrpc.Response) (bool, error) {
				// For debug purposes only
				// b, err := json.Marshal(req)
				// if err != nil {
				// 	t.Fatal(err)
				// }
				// t.Log(string(b))
				// b, err = json.Marshal(resp)
				// if err != nil {
				// 	t.Fatal(err)
				// }
				// t.Log(string(b))
				return false, nil
			}}, // should succeed w/ rollback
	}

	for n, td := range setTestData {
		t.Run(td.testName, func(t *testing.T) {
			// Set reference values, should not fail
			_, err := c.BulkSet([]srljrpc.PV{}, []srljrpc.PV{}, []srljrpc.PV{{"/interface[name=system0]/description", srljrpc.CommandValue("Initial Value")}}, yms.SRL, 0)
			if err != nil {
				t.Fatal(err)
			}

			chResp := make(chan *srljrpc.Response)
			chErr := make(chan error)
			go func() {
				resp, err := c.BulkSetCallBack(td.delete, td.replace, td.update, yms.SRL, 5, 3, td.cbf)

				chResp <- resp
				chErr <- err
			}()

			switch n {
			case 0: // test with confirm timeout and CallBackConfirm returning true.
				var ctRespData []string
				t.Logf("Set Update w/ confirm timeout and CallBackConfirm true against CANDIDATE datastore with SRL default target => Waiting for 6 seconds...")
				time.Sleep(6 * time.Second)
				ctResp, err := c.Get("/interface[name=system0]/description")
				if err != nil {
					t.Fatal(err)
				}
				// Unmarshal response
				err = json.Unmarshal(ctResp.Result, &ctRespData)
				if err != nil {
					t.Fatal(err)
				}
				// Check if values are updated
				if ctRespData[0] != "System loopback" {
					t.Errorf("got: %s, while should be: %s", ctRespData[0], "System loopback")
				}
				resp := <-chResp
				if resp == nil { // should be non-nil, as CallBackConfirm returned true.
					t.Errorf("got: %v, while should be: %s", resp, "non-nil")
					break
				}
				b, err := json.Marshal(resp.Error)
				if err != nil {
					t.Fatal(err)
				}
				t.Log(string(b))
				// t.Log(ctRespData)
			case 1: // test with confirm timeout and CallBackConfirm returning false.
				var ctRespData []string
				t.Logf("Set Replace w/ confirm timeout and CallBackConfirm false against CANDIDATE datastore with SRL default target => Waiting for 6 seconds...")
				time.Sleep(6 * time.Second)
				ctResp, err := c.Get("/interface[name=system0]/description")
				if err != nil {
					t.Fatal(err)
				}
				// Unmarshal response
				err = json.Unmarshal(ctResp.Result, &ctRespData)
				if err != nil {
					t.Fatal(err)
				}
				// Check if values are rolled back
				if ctRespData[0] != "Initial Value" {
					t.Errorf("got: %s, while should be: %s", ctRespData[0], "Initial Value")
				}
				resp := <-chResp
				if resp != nil { // should be nil, as CallBackConfirm returned false.
					t.Errorf("got: %v, while should be: %s", resp, "nil")

				}
			}
			// Getting error from channel.
			err = <-chErr
			checkErrGotVSExp(err, td.expErr, t)
		})
	}

}

func TestBulkDiff(t *testing.T) {
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
			expErr: apierr.ErrClntJSONRPCResp}, // should fail, invalid path
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
			expErr: apierr.ErrClntRPCReqCreation}, // should fail, missed value
	}
	for _, td := range setOCTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.BulkDiff(td.delete, td.replace, td.update, yms.OC)
			checkErrGotVSExp(err, td.expErr, t)
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
			expErr: apierr.ErrClntJSONRPCResp}, // should fail, invalid path
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
			expErr: apierr.ErrClntRPCReqCreation}, // should fail, missed value
	}

	for _, td := range setTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.BulkDiff(td.delete, td.replace, td.update, yms.SRL)
			checkErrGotVSExp(err, td.expErr, t)
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
			pvs:    []srljrpc.PV{{"/interface[name=system0]/invalid:test", srljrpc.CommandValue("")}},
			expErr: apierr.ErrClntJSONRPCResp,
		}, // should fail, invalid path
	}
	for _, td := range getTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.Replace(0, td.pvs...)
			checkErrGotVSExp(err, td.expErr, t)
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
			paths:  []string{"/interface[name=system0]/invalid"},
			expErr: apierr.ErrClntJSONRPCResp,
		}, // should fail, invalid path
	}
	for _, td := range setTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.Delete(0, td.paths...) // ct set to 0 to avoid confirm timeout
			checkErrGotVSExp(err, td.expErr, t)
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
			expErr: apierr.ErrClntJSONRPCResp,
		}, // should fail, invalid path
		{testName: "Validate against CANDIDATE datastore with default target and missed value",
			pvs: []srljrpc.PV{
				{"/interface[name=system0]/description", srljrpc.CommandValue("")}},
			expErr: apierr.ErrClntRPCReqCreation,
		}, // should fail, missed value
	}

	for _, td := range validateTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.Validate(actions.UPDATE, td.pvs...)
			checkErrGotVSExp(err, td.expErr, t)
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
			expErr: apierr.ErrClntJSONRPCResp,
		}, // should fail, invalid path
		{testName: "Set against TOOLS with double value",
			pvs: []srljrpc.PV{
				{"/network-instance[name=default]/protocols/bgp/group[group-name=underlay]/soft-clear/peer-as:65020", srljrpc.CommandValue("65020")}},
			expErr: apierr.ErrClntRPCReqCreation,
		}, // should fail, value specified two times
	}

	for _, td := range toolsTestData {
		t.Run(td.testName, func(t *testing.T) {
			_, err := c.Tools(td.pvs...)
			checkErrGotVSExp(err, td.expErr, t)
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
			checkErrGotVSExp(err, td.expErr, t)
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
				cmds:   []string{"show version", "", "show acl summary"},
				of:     formats.TEXT,
				expErr: apierr.ErrClntRPCReqCreation,
			}, // should fail, empty command
		}
		// CLI with default target using CLI()
		for _, td := range cliTestData {
			t.Run(td.testName, func(t *testing.T) {
				r, err := c.CLI(td.cmds, td.of)
				checkErrGotVSExp(err, td.expErr, t)
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
		uerr := err.(apierr.ClientError).Unwrap().Error()
		t.Fatalf("can't create client: %v", err)
		t.Logf("underlying error: %v", uerr)
	}
	return c
}
