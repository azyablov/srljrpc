//go:build unit

package srljrpc_test

import (
	"bytes"
	"errors"
	"html/template"
	"testing"

	"github.com/azyablov/srljrpc"
	"github.com/azyablov/srljrpc/actions"
	"github.com/azyablov/srljrpc/apierr"
	"github.com/azyablov/srljrpc/datastores"
	"github.com/azyablov/srljrpc/formats"
	"github.com/azyablov/srljrpc/methods"
	"github.com/azyablov/srljrpc/yms"
	"github.com/google/go-cmp/cmp"
)

func TestNewRequest_Get(t *testing.T) {
	// GET method testing
	cmdArgs := []struct {
		action actions.EnumActions
		path   string
		value  srljrpc.CommandValue
		opts   []srljrpc.CommandOption
	}{
		{actions.NONE, "/system/name/host-name", srljrpc.CommandValue(""), nil}, // should succeed
		{actions.NONE, "/system/name/host-name", srljrpc.CommandValue(""), []srljrpc.CommandOption{srljrpc.WithDefaults(), srljrpc.WithoutRecursion(), srljrpc.WithDatastore(datastores.STATE)}}, // should succeed
		{actions.NONE, "/system/name/host-name", srljrpc.CommandValue("shouldFail"), nil},                                                                                                        // should fail, bcz of value
		{actions.DELETE, "/system/name/host-name", srljrpc.CommandValue(""), nil},                                                                                                                // should fail, bcz of action
		{actions.NONE, "/system/name/host-name", srljrpc.CommandValue(""), []srljrpc.CommandOption{srljrpc.WithDatastore(datastores.TOOLS)}},                                                     // should fail, bcz of datastore
		{actions.NONE, "", srljrpc.CommandValue(""), nil},                                                                                                                                        // should fail, bcz of empty path
		{actions.NONE, "/system/name/host-name", srljrpc.CommandValue(""), []srljrpc.CommandOption{srljrpc.WithDefaults(), srljrpc.WithoutRecursion(), srljrpc.WithDatastore(datastores.STATE)}}, // command is ok, but should fail due to unsupported datastore TOOLS under request
		{actions.NONE, "/system/name/host-name", srljrpc.CommandValue(""), nil},                                                                                                                  // command is ok, but should fail due to unsupported confirm timeout
	}
	cmdResults := []*srljrpc.Command{}

	for _, ca := range cmdArgs {
		t.Run("NewGETCommand", func(t *testing.T) {
			cmd, err := srljrpc.NewCommand(ca.action, ca.path, ca.value, ca.opts...)
			if err != nil {
				t.Fatal(err)
			}
			cmdResults = append(cmdResults, cmd)
		})
	}

	//m := "get"
	testData := []struct {
		testName  string
		cmd       *srljrpc.Command
		expReqErr error
		tmplJSON  string
		opts      []srljrpc.RequestOption
	}{
		{"Basic GET", cmdResults[0], nil, `{"jsonrpc":"2.0","id":{{.}},"method":"get","params":{"commands":[{"path":"/system/name/host-name"}],"output-format":"json"}}`, []srljrpc.RequestOption{srljrpc.WithOutputFormat(formats.JSON)}},
		{"Basic GET with options", cmdResults[1], nil, `{"jsonrpc":"2.0","id":{{.}},"method":"get","params":{"commands":[{"path":"/system/name/host-name","recursive":false,"include-field-defaults":true,"datastore":"state"}]}}`, []srljrpc.RequestOption{}},
		{"Basic GET with value", cmdResults[2], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic GET with actions", cmdResults[3], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic GET with command TOOLS datastore", cmdResults[4], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic GET with empty path", cmdResults[5], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic GET with request STATE datastore", cmdResults[6], apierr.ErrMsgReqGetDSNotAllowed, `null`, []srljrpc.RequestOption{srljrpc.WithRequestDatastore(datastores.TOOLS)}},
		{"Basic GET", cmdResults[7], apierr.ErrMsgReqSettingConfirmTimeout, `null`, []srljrpc.RequestOption{srljrpc.WithOutputFormat(formats.JSON), srljrpc.WithConfirmTimeout(5)}},
	}

	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			r, err := srljrpc.NewRequest(methods.GET, []*srljrpc.Command{td.cmd}, td.opts...)
			checkErrGotVSExp(err, td.expReqErr, t)
			expJSON := new(bytes.Buffer)
			expJSON.Grow(512)
			if td.tmplJSON == `null` {
				// null in case of expected error
				expJSON.WriteString(`null`)
			} else {
				// creating template
				tmpl, err := template.New("get").Parse(td.tmplJSON)
				if err != nil {
					t.Fatal(err)
				}
				// embedding ID into template
				err = tmpl.Execute(expJSON, r.ID)
				if err != nil {
					t.Fatal(err)
				}
			}
			b, err := r.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			// Uncomment for debugging
			// t.Logf("GET request: %s", string(b))
			// t.Logf("GET expected: %s", expJSON.String())
			out := cmp.Diff(string(b), expJSON.String())
			if out != "" {
				t.Logf(out)
				t.Errorf("got %s, while should be %s", string(b), expJSON.String())
			}
		})
	}
}

func TestNewGetRequest(t *testing.T) {
	//m := "get"
	testData := []struct {
		testName  string
		paths     []string
		rec       bool
		def       bool
		of        formats.EnumOutputFormats
		ds        datastores.EnumDatastores
		expReqErr error
		tmplJSON  string
	}{
		{"GET Request w/ recursion w/o def w/ JSON, w/ RUNNING", []string{"/system/name/host-name", "/interface[name=ethernet-1/1]/description"}, true, false, formats.JSON, datastores.RUNNING, nil,
			`{"jsonrpc":"2.0","id":{{.}},"method":"get","params":{"commands":[{"path":"/system/name/host-name","recursive":false},{"path":"/interface[name=ethernet-1/1]/description","recursive":false}],"output-format":"json","datastore":"running"}}`}, // should succeed
		{"GET Request w/ recursion w/o def w/ TABLE, w/ CANDIDATE", []string{"/system/name/host-name", "/interface[name=ethernet-1/1]/description"}, false, true, formats.TABLE, datastores.CANDIDATE, nil,
			`{"jsonrpc":"2.0","id":{{.}},"method":"get","params":{"commands":[{"path":"/system/name/host-name","include-field-defaults":true},{"path":"/interface[name=ethernet-1/1]/description","include-field-defaults":true}],"output-format":"table","datastore":"candidate"}}`}, // should succeed
		{"GET Request w/ recursion w/o def w/ TEXT, w/ STATE", []string{"/system/name/host-name", "/interface[name=ethernet-1/1]/description"}, false, true, formats.TEXT, datastores.STATE, nil,
			`{"jsonrpc":"2.0","id":{{.}},"method":"get","params":{"commands":[{"path":"/system/name/host-name","include-field-defaults":true},{"path":"/interface[name=ethernet-1/1]/description","include-field-defaults":true}],"output-format":"text","datastore":"state"}}`}, // should succeed
		{"GET Request w/ recursion w/o def w/ JSON, w/ TOOLS", []string{"/system/name/host-name", "/interface[name=ethernet-1/1]/description"}, false, false, formats.JSON, datastores.TOOLS, apierr.ErrMsgReqGetDSNotAllowed,
			"null"}, // should fail, bcz of unsupported datastore TOOLS
	}

	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			r, err := srljrpc.NewGetRequest(td.paths, td.rec, td.def, td.of, td.ds)
			checkErrGotVSExp(err, td.expReqErr, t)
			expJSON := new(bytes.Buffer)
			expJSON.Grow(512)
			if td.tmplJSON == `null` {
				// null in case of expected error
				expJSON.WriteString(`null`)
			} else {
				// creating template
				tmpl, err := template.New("get").Parse(td.tmplJSON)
				if err != nil {
					t.Fatal(err)
				}
				// embedding ID into template
				err = tmpl.Execute(expJSON, r.ID)
				if err != nil {
					t.Fatal(err)
				}
			}
			b, err := r.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			// Uncomment for debugging
			// t.Logf("GET request: %s", string(b))
			// t.Logf("GET expected: %s", expJSON.String())
			out := cmp.Diff(string(b), expJSON.String())
			if out != "" {
				t.Logf(out)
				t.Errorf("got %s, while should be %s", string(b), expJSON.String())
			}
		})
	}
}

func TestNewRequest_Set(t *testing.T) {
	// SET method testing
	cmdArgs := []struct {
		action actions.EnumActions
		path   string
		value  srljrpc.CommandValue
		opts   []srljrpc.CommandOption
	}{
		{actions.UPDATE, "/system/name/host-name", srljrpc.CommandValue("SetUpdate"), []srljrpc.CommandOption{}},                                              // should succeed
		{actions.REPLACE, "/system/name/host-name", srljrpc.CommandValue("SetReplace"), []srljrpc.CommandOption{}},                                            // should succeed
		{actions.DELETE, "/system/name/host-name", srljrpc.CommandValue("SetDelete"), []srljrpc.CommandOption{}},                                              // should fail cause of value not allowed
		{actions.DELETE, "/system/name/host-name", srljrpc.CommandValue(""), []srljrpc.CommandOption{srljrpc.WithDatastore(datastores.RUNNING)}},              // should be failing due to unsupported datastore by SET
		{actions.NONE, "/system/name/host-name", srljrpc.CommandValue("test"), []srljrpc.CommandOption{}},                                                     // should be failing due to unsupported action NONE by SET
		{actions.UPDATE, "", srljrpc.CommandValue("test"), []srljrpc.CommandOption{}},                                                                         // should be failing due to empty path
		{actions.UPDATE, "/system/name/host-name", srljrpc.CommandValue(""), []srljrpc.CommandOption{}},                                                       // should be failing due to empty value
		{actions.UPDATE, "/system/name/host-name:test", srljrpc.CommandValue(""), []srljrpc.CommandOption{}},                                                  // should succeed, bcz of :test value specified as part of path
		{actions.REPLACE, "/system/name/host-name:test", srljrpc.CommandValue("TEST"), []srljrpc.CommandOption{}},                                             // should fail, bcz of :test value specified as part of path and value is not empty
		{actions.REPLACE, "/system/name/host-name:test:TEST", srljrpc.CommandValue(""), []srljrpc.CommandOption{}},                                            // should succeed, bcz of :test:TEST value specified as part of path
		{actions.UPDATE, "/system/name/host-name", srljrpc.CommandValue("SetUpdate"), []srljrpc.CommandOption{srljrpc.WithDatastore(datastores.TOOLS)}},       // should fail because of command lvl datastore specified.
		{actions.REPLACE, "/system/name/host-name", srljrpc.CommandValue("SetReplace"), []srljrpc.CommandOption{srljrpc.WithDatastore(datastores.CANDIDATE)}}, // should fail because of command lvl datastore specified.
		{actions.UPDATE, "/system/name/host-name", srljrpc.CommandValue("SetUpdateTEXTSRL"), []srljrpc.CommandOption{}},                                       // should succeed
		{actions.REPLACE, "/system/name/host-name", srljrpc.CommandValue("SetReplaceRUNNING"), []srljrpc.CommandOption{}},                                     // should fail because of RUNNING datastore specified.
		{actions.UPDATE, "/system/name/host-name", srljrpc.CommandValue("SetUpdateTEXTSRL"), []srljrpc.CommandOption{}},                                       // should succeed
		{actions.DELETE, "/system/name/host-name", srljrpc.CommandValue(""), []srljrpc.CommandOption{}},                                                       // should succeed
	}
	cmdResults := []*srljrpc.Command{}

	for _, ca := range cmdArgs {
		t.Run("NewSETCommand", func(t *testing.T) {
			cmd, err := srljrpc.NewCommand(ca.action, ca.path, ca.value, ca.opts...)
			if err != nil {
				t.Fatal(err)
			}
			cmdResults = append(cmdResults, cmd)
		})
	}

	//m := "set"
	testData := []struct {
		testName  string
		cmd       *srljrpc.Command
		expReqErr error
		tmplJSON  string
		opts      []srljrpc.RequestOption
	}{
		{"Basic SET UPDATE", cmdResults[0], nil, `{"jsonrpc":"2.0","id":{{.}},"method":"set","params":{"commands":[{"path":"/system/name/host-name","value":"SetUpdate","action":"update"}]}}`, []srljrpc.RequestOption{}},
		{"Basic SET REPLACE", cmdResults[1], nil, `{"jsonrpc":"2.0","id":{{.}},"method":"set","params":{"commands":[{"path":"/system/name/host-name","value":"SetReplace","action":"replace"}]}}`, []srljrpc.RequestOption{}},
		{"Basic SET DELETE", cmdResults[2], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic SET with unsupported command lvl datastore RUNNING}", cmdResults[3], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic SET without action", cmdResults[4], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic SET with empty path", cmdResults[5], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic SET with empty value", cmdResults[6], apierr.ErrMsgDSCandidateUpdateNoValue, `null`, []srljrpc.RequestOption{srljrpc.WithRequestDatastore(datastores.CANDIDATE)}},
		{"Basic SET with k:v path", cmdResults[7], nil, `{"jsonrpc":"2.0","id":{{.}},"method":"set","params":{"commands":[{"path":"/system/name/host-name:test","action":"update"}]}}`, []srljrpc.RequestOption{}},
		{"Basic SET with k:v path and value", cmdResults[8], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic SET with incorrect k:v path", cmdResults[9], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic SET with unsupported command lvl datastore TOOLS", cmdResults[10], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic SET with unsupported command lvl datastore CANDIDATE", cmdResults[11], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic SET UPDATE output format TEXT and ym SRL ", cmdResults[12], nil,
			`{"jsonrpc":"2.0","id":{{.}},"method":"set","params":{"commands":[{"path":"/system/name/host-name","value":"SetUpdateTEXTSRL","action":"update"}],"output-format":"text","yang-models":"srl"}}`,
			[]srljrpc.RequestOption{srljrpc.WithOutputFormat(formats.TEXT), srljrpc.WithYmType(yms.SRL)}},
		{"Basic SET REPLACE w/ RUNNING datastore", cmdResults[13], apierr.ErrMsgDSToolsCandidateSetOnly, `null`, []srljrpc.RequestOption{srljrpc.WithRequestDatastore(datastores.RUNNING)}},
		{"Basic SET UPDATE output format TEXT, datastore TOOLS and ym OC", cmdResults[14], nil,
			`{"jsonrpc":"2.0","id":{{.}},"method":"set","params":{"commands":[{"path":"/system/name/host-name","value":"SetUpdateTEXTSRL","action":"update"}],"output-format":"table","datastore":"tools","yang-models":"oc"}}`,
			[]srljrpc.RequestOption{srljrpc.WithOutputFormat(formats.TABLE), srljrpc.WithYmType(yms.OC), srljrpc.WithRequestDatastore(datastores.TOOLS)}},
		{"Basic SET DELETE output format TEXT with confirm timeout, datastore CANDIDATE and ym OC", cmdResults[15], nil,
			`{"jsonrpc":"2.0","id":{{.}},"method":"set","params":{"commands":[{"path":"/system/name/host-name","action":"delete"}],"output-format":"table","datastore":"candidate","yang-models":"oc","confirm-timeout":22}}`,
			[]srljrpc.RequestOption{srljrpc.WithOutputFormat(formats.TABLE), srljrpc.WithYmType(yms.OC), srljrpc.WithRequestDatastore(datastores.CANDIDATE), srljrpc.WithConfirmTimeout(22)}}, // with new confirm timeout
	}

	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			r, err := srljrpc.NewRequest(methods.SET, []*srljrpc.Command{td.cmd}, td.opts...)
			checkErrGotVSExp(err, td.expReqErr, t)
			expJSON := new(bytes.Buffer)
			expJSON.Grow(512)
			if td.tmplJSON == `null` {
				// null in case of expected error
				expJSON.WriteString(`null`)
			} else {
				// creating template
				tmpl, err := template.New("set").Parse(td.tmplJSON)
				if err != nil {
					t.Fatal(err)
				}
				// embedding ID into template
				err = tmpl.Execute(expJSON, r.ID)
				if err != nil {
					t.Fatal(err)
				}
			}
			b, err := r.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			// Uncomment for debugging
			// t.Logf("SET request: %s", string(b))
			// t.Logf("SET expected: %s", expJSON.String())
			out := cmp.Diff(string(b), expJSON.String())
			if out != "" {
				t.Logf(out)
				t.Errorf("got %s, while should be %s", string(b), expJSON.String())
			}
		})
	}
}

func TestNewSetRequest(t *testing.T) {
	testData := []struct {
		testName  string
		delete    []srljrpc.PV
		replace   []srljrpc.PV
		update    []srljrpc.PV
		ym        yms.EnumYmType
		of        formats.EnumOutputFormats
		ds        datastores.EnumDatastores
		ct        int
		expReqErr error
		tmplJSON  string
	}{
		{"SET Request w/ SRL w/ JSON w/ CANDIDATE",
			[]srljrpc.PV{{"/system/name/host-name", srljrpc.CommandValue("Delete")}},
			[]srljrpc.PV{{"/system/name/host-name", srljrpc.CommandValue("Replace")}},
			[]srljrpc.PV{{"/system/name/host-name", srljrpc.CommandValue("Update")}}, yms.SRL, formats.JSON, datastores.CANDIDATE, 0, nil,
			`{"jsonrpc":"2.0","id":{{.}},"method":"set","params":{"commands":[{"path":"/system/name/host-name","action":"delete"},{"path":"/system/name/host-name","value":"Replace","action":"replace"},{"path":"/system/name/host-name","value":"Update","action":"update"}],"output-format":"json","datastore":"candidate","yang-models":"srl"}}`}, // should succeed
		{"SET Request w/ OC w/ TEXT w/ TOOLS",
			[]srljrpc.PV{{"/system/name/host-name", srljrpc.CommandValue("Delete")}},
			[]srljrpc.PV{{"/system/name/host-name", srljrpc.CommandValue("Replace")}},
			[]srljrpc.PV{{"/network-instance[name=default]/protocols/bgp/neighbor[peer-address=100.24.11.1]/reset-peer", srljrpc.CommandValue("Update")}}, yms.OC, formats.TEXT, datastores.TOOLS, 0,
			apierr.ErrMsgSetNotAllowedActForTools,
			`null`}, // should fail, bcz of unsupported datastore TOOLS
		{"SET Request w/ SRL w/ TABLE w/ RUNNING",
			[]srljrpc.PV{{"/system/name/host-name", srljrpc.CommandValue("Delete")}},
			[]srljrpc.PV{{"/system/name/host-name", srljrpc.CommandValue("Replace")}},
			[]srljrpc.PV{{"/system/name/host-name", srljrpc.CommandValue("Update")}}, yms.SRL, formats.TABLE, datastores.RUNNING, 0,
			apierr.ErrMsgDSToolsCandidateSetOnly,
			`null`}, // should fail, bcz of unsupported datastore RUNNING
		{"SET Request w/ OC w/ TEXT w/ CANDIDATE",
			[]srljrpc.PV{},
			[]srljrpc.PV{},
			[]srljrpc.PV{{"/system/name/host-name", srljrpc.CommandValue("")}}, yms.SRL, formats.JSON, datastores.CANDIDATE, 0,
			apierr.ErrMsgDSCandidateUpdateNoValue,
			`null`}, // should fail, bcz UPDATE action should have value for CANDIDATE datastore
		{"SET Request w/ SRL w/ JSON w/ TOOLS",
			[]srljrpc.PV{},
			[]srljrpc.PV{},
			[]srljrpc.PV{{"/network-instance[name=default]/protocols/bgp/neighbor[peer-address=100.24.11.1]/reset-peer", srljrpc.CommandValue("")}}, yms.SRL, formats.JSON, datastores.TOOLS, 22, nil,
			`{"jsonrpc":"2.0","id":{{.}},"method":"set","params":{"commands":[{"path":"/network-instance[name=default]/protocols/bgp/neighbor[peer-address=100.24.11.1]/reset-peer","action":"update"}],"output-format":"json","datastore":"tools","yang-models":"srl","confirm-timeout":22}}`},
		// should succeed, bcz UPDATE action value is optional for TOOLS datastore. With new confirm timeout.
	}

	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			r, err := srljrpc.NewSetRequest(td.delete, td.replace, td.update, td.ym, td.of, td.ds, td.ct)
			checkErrGotVSExp(err, td.expReqErr, t)
			expJSON := new(bytes.Buffer)
			expJSON.Grow(512)
			if td.tmplJSON == `null` {
				// null in case of expected error
				expJSON.WriteString(`null`)
			} else {
				// creating template
				tmpl, err := template.New("get").Parse(td.tmplJSON)
				if err != nil {
					t.Fatal(err)
				}
				// embedding ID into template
				err = tmpl.Execute(expJSON, r.ID)
				if err != nil {
					t.Fatal(err)
				}
			}
			b, err := r.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			// Uncomment for debugging
			//t.Logf("SET request: %s", string(b))
			// t.Logf("SET expected: %s", expJSON.String())
			out := cmp.Diff(string(b), expJSON.String())
			if out != "" {
				t.Logf(out)
				t.Errorf("got %s, while should be %s", string(b), expJSON.String())
			}
		})
	}
}

func TestNewRequest_Validate(t *testing.T) {
	// VALIDATE method testing
	cmdArgs := []struct {
		action actions.EnumActions
		path   string
		value  srljrpc.CommandValue
		opts   []srljrpc.CommandOption
	}{
		{actions.UPDATE, "/system/name/host-name", srljrpc.CommandValue("SetUpdate"), []srljrpc.CommandOption{}},                                              // should succeed
		{actions.REPLACE, "/system/name/host-name", srljrpc.CommandValue("SetReplace"), []srljrpc.CommandOption{}},                                            // should succeed
		{actions.DELETE, "/system/name/host-name", srljrpc.CommandValue("SetDelete"), []srljrpc.CommandOption{}},                                              // should fail cause of value not allowed
		{actions.DELETE, "/system/name/host-name", srljrpc.CommandValue(""), []srljrpc.CommandOption{srljrpc.WithDatastore(datastores.RUNNING)}},              // should be failing due to unsupported datastore by SET
		{actions.NONE, "/system/name/host-name", srljrpc.CommandValue("test"), []srljrpc.CommandOption{}},                                                     // should be failing due to unsupported action NONE by SET
		{actions.UPDATE, "", srljrpc.CommandValue("test"), []srljrpc.CommandOption{}},                                                                         // should be failing due to empty path
		{actions.UPDATE, "/system/name/host-name", srljrpc.CommandValue(""), []srljrpc.CommandOption{}},                                                       // should be failing due to empty value
		{actions.UPDATE, "/system/name/host-name:test", srljrpc.CommandValue(""), []srljrpc.CommandOption{}},                                                  // should succeed, bcz of :test value specified as part of path
		{actions.REPLACE, "/system/name/host-name:test", srljrpc.CommandValue("TEST"), []srljrpc.CommandOption{}},                                             // should fail, bcz of :test value specified as part of path and value is not empty
		{actions.REPLACE, "/system/name/host-name:test:TEST", srljrpc.CommandValue(""), []srljrpc.CommandOption{}},                                            // should succeed, bcz of :test:TEST value specified as part of path
		{actions.UPDATE, "/system/name/host-name", srljrpc.CommandValue("SetUpdate"), []srljrpc.CommandOption{srljrpc.WithDatastore(datastores.TOOLS)}},       // should fail because of command lvl datastore specified.
		{actions.REPLACE, "/system/name/host-name", srljrpc.CommandValue("SetReplace"), []srljrpc.CommandOption{srljrpc.WithDatastore(datastores.CANDIDATE)}}, // should fail because of command lvl datastore specified.
		{actions.UPDATE, "/system/name/host-name", srljrpc.CommandValue("SetUpdateTEXTSRL"), []srljrpc.CommandOption{}},                                       // should succeed
		{actions.REPLACE, "/system/name/host-name", srljrpc.CommandValue("SetReplaceRUNNING"), []srljrpc.CommandOption{}},                                     // should fail because of RUNNING datastore specified.
		{actions.UPDATE, "/system/name/host-name", srljrpc.CommandValue("SetUpdateTEXTSRL"), []srljrpc.CommandOption{}},                                       // should succeed
	}
	cmdResults := []*srljrpc.Command{}

	for _, ca := range cmdArgs {
		t.Run("NewVALIDATECommand", func(t *testing.T) {
			cmd, err := srljrpc.NewCommand(ca.action, ca.path, ca.value, ca.opts...)
			if err != nil {
				t.Fatal(err)
			}
			cmdResults = append(cmdResults, cmd)
		})
	}

	//m := "validate"
	testData := []struct {
		testName  string
		cmd       *srljrpc.Command
		expReqErr error
		tmplJSON  string
		opts      []srljrpc.RequestOption
	}{
		{"Basic VALIDATE UPDATE", cmdResults[0], nil, `{"jsonrpc":"2.0","id":{{.}},"method":"validate","params":{"commands":[{"path":"/system/name/host-name","value":"SetUpdate","action":"update"}]}}`, []srljrpc.RequestOption{}},
		{"Basic VALIDATE REPLACE", cmdResults[1], nil, `{"jsonrpc":"2.0","id":{{.}},"method":"validate","params":{"commands":[{"path":"/system/name/host-name","value":"SetReplace","action":"replace"}]}}`, []srljrpc.RequestOption{}},
		{"Basic VALIDATE DELETE", cmdResults[2], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic VALIDATE with unsupported command lvl datastore RUNNING}", cmdResults[3], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic VALIDATE without action", cmdResults[4], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic VALIDATE with empty path", cmdResults[5], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic VALIDATE with empty value", cmdResults[6], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic VALIDATE with k:v path", cmdResults[7], nil, `{"jsonrpc":"2.0","id":{{.}},"method":"validate","params":{"commands":[{"path":"/system/name/host-name:test","action":"update"}]}}`, []srljrpc.RequestOption{}},
		{"Basic VALIDATE with k:v path and value", cmdResults[8], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic VALIDATE with incorrect k:v path", cmdResults[9], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic VALIDATE with unsupported command lvl datastore TOOLS", cmdResults[10], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic VALIDATE with unsupported command lvl datastore CANDIDATE", cmdResults[11], apierr.ErrMsgReqAddingCmds, `null`, []srljrpc.RequestOption{}},
		{"Basic VALIDATE UPDATE output format TEXT and ym SRL", cmdResults[12], nil, `{"jsonrpc":"2.0","id":{{.}},"method":"validate","params":{"commands":[{"path":"/system/name/host-name","value":"SetUpdateTEXTSRL","action":"update"}],"output-format":"text","yang-models":"srl"}}`,
			[]srljrpc.RequestOption{srljrpc.WithOutputFormat(formats.TEXT), srljrpc.WithYmType(yms.SRL)}},
		{"Basic VALIDATE REPLACE w/ RUNNING datastore", cmdResults[13], apierr.ErrMsgDSCandidateValidateOnly, `null`, []srljrpc.RequestOption{srljrpc.WithRequestDatastore(datastores.RUNNING)}},
		{"Basic VALIDATE UPDATE output format TEXT, datastore TOOLS and ym OC", cmdResults[14], nil,
			`{"jsonrpc":"2.0","id":{{.}},"method":"validate","params":{"commands":[{"path":"/system/name/host-name","value":"SetUpdateTEXTSRL","action":"update"}],"output-format":"table","datastore":"candidate","yang-models":"oc"}}`,
			[]srljrpc.RequestOption{srljrpc.WithOutputFormat(formats.TABLE), srljrpc.WithYmType(yms.OC), srljrpc.WithRequestDatastore(datastores.CANDIDATE)}},
	}

	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			r, err := srljrpc.NewRequest(methods.VALIDATE, []*srljrpc.Command{td.cmd}, td.opts...)
			checkErrGotVSExp(err, td.expReqErr, t)
			expJSON := new(bytes.Buffer)
			expJSON.Grow(512)
			if td.tmplJSON == `null` {
				// null in case of expected error
				expJSON.WriteString(`null`)
			} else {
				// creating template
				tmpl, err := template.New("set").Parse(td.tmplJSON)
				if err != nil {
					t.Fatal(err)
				}
				// embedding ID into template
				err = tmpl.Execute(expJSON, r.ID)
				if err != nil {
					t.Fatal(err)
				}
			}
			b, err := r.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			// Uncomment for debugging
			// t.Logf("VALIDATE request: %s", string(b))
			// t.Logf("VALIDATE expected: %s", expJSON.String())
			out := cmp.Diff(string(b), expJSON.String())
			if out != "" {
				t.Logf(out)
				t.Errorf("got %s, while should be %s", string(b), expJSON.String())
			}
		})
	}
}

func TestNewValidateRequest(t *testing.T) {
	//m := "validate"
	testData := []struct {
		testName  string
		delete    []srljrpc.PV
		replace   []srljrpc.PV
		update    []srljrpc.PV
		ym        yms.EnumYmType
		of        formats.EnumOutputFormats
		ds        datastores.EnumDatastores
		expReqErr error
		tmplJSON  string
	}{
		{"VALIDATE Request w/ SRL w/ JSON w/ CANDIDATE",
			[]srljrpc.PV{{"/system/name/host-name", srljrpc.CommandValue("Delete")}},
			[]srljrpc.PV{{"/system/name/host-name", srljrpc.CommandValue("Replace")}},
			[]srljrpc.PV{{"/system/name/host-name", srljrpc.CommandValue("Update")}}, yms.SRL, formats.JSON, datastores.CANDIDATE, nil,
			`{"jsonrpc":"2.0","id":{{.}},"method":"validate","params":{"commands":[{"path":"/system/name/host-name","action":"delete"},{"path":"/system/name/host-name","value":"Replace","action":"replace"},{"path":"/system/name/host-name","value":"Update","action":"update"}],"output-format":"json","datastore":"candidate","yang-models":"srl"}}`}, // should succeed
		{"VALIDATE Request w/ OC w/ TEXT w/ TOOLS",
			[]srljrpc.PV{{"/system/name/host-name", srljrpc.CommandValue("Delete")}},
			[]srljrpc.PV{{"/system/name/host-name", srljrpc.CommandValue("Replace")}},
			[]srljrpc.PV{{"/system/name/host-name", srljrpc.CommandValue("Update")}}, yms.OC, formats.TEXT, datastores.TOOLS, apierr.ErrMsgDSCandidateValidateOnly,
			`null`}, // should fail, bcz of unsupported datastore TOOLS
		{"VALIDATE Request w/ SRL w/ TABLE w/ CANDIDATE",
			[]srljrpc.PV{{"/system/name/host-name", srljrpc.CommandValue("Delete")}},
			[]srljrpc.PV{{"/system/name/host-name", srljrpc.CommandValue("Replace")}},
			[]srljrpc.PV{{"/system/name/host-name", srljrpc.CommandValue("Update")}}, yms.OC, formats.TABLE, datastores.CANDIDATE, nil,
			`{"jsonrpc":"2.0","id":{{.}},"method":"validate","params":{"commands":[{"path":"/system/name/host-name","action":"delete"},{"path":"/system/name/host-name","value":"Replace","action":"replace"},{"path":"/system/name/host-name","value":"Update","action":"update"}],"output-format":"table","datastore":"candidate","yang-models":"oc"}}`}, // should fail, bcz of unsupported datastore RUNNING

	}

	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			r, err := srljrpc.NewValidateRequest(td.delete, td.replace, td.update, td.ym, td.of, td.ds)
			checkErrGotVSExp(err, td.expReqErr, t)
			expJSON := new(bytes.Buffer)
			expJSON.Grow(512)
			if td.tmplJSON == `null` {
				// null in case of expected error
				expJSON.WriteString(`null`)
			} else {
				// creating template
				tmpl, err := template.New("get").Parse(td.tmplJSON)
				if err != nil {
					t.Fatal(err)
				}
				// embedding ID into template
				err = tmpl.Execute(expJSON, r.ID)
				if err != nil {
					t.Fatal(err)
				}
			}
			b, err := r.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			// Uncomment for debugging
			// t.Logf("SET request: %s", string(b))
			// t.Logf("SET expected: %s", expJSON.String())
			out := cmp.Diff(string(b), expJSON.String())
			if out != "" {
				t.Logf(out)
				t.Errorf("got %s, while should be %s", string(b), expJSON.String())
			}
		})
	}
}

func TestNewCLIRequest(t *testing.T) {
	testData := []struct {
		testName  string
		cmds      []string
		tmplJSON  string
		of        formats.EnumOutputFormats
		expReqErr error
	}{
		{"CLI Request with JSON output format", []string{"show version", "show system lldp neighbor", "info from state network-instance default interface system0.0"}, `{"jsonrpc":"2.0","id":{{.}},"method":"cli","params":{"commands":["show version","show system lldp neighbor","info from state network-instance default interface system0.0"],"output-format":"json"}}`, formats.JSON, nil},
		{"CLI Request with TEXT output format", []string{"show version", "show system lldp neighbor", "info from state network-instance default interface system0.0"}, `{"jsonrpc":"2.0","id":{{.}},"method":"cli","params":{"commands":["show version","show system lldp neighbor","info from state network-instance default interface system0.0"],"output-format":"text"}}`, formats.TEXT, nil},
		{"CLI Request with TABLE output format", []string{"show version", "show system lldp neighbor", "info from state network-instance default interface system0.0"}, `{"jsonrpc":"2.0","id":{{.}},"method":"cli","params":{"commands":["show version","show system lldp neighbor","info from state network-instance default interface system0.0"],"output-format":"table"}}`, formats.TABLE, nil},
		{"CLI Request with empty command", []string{"show version", "", "info from state network-instance default interface system0.0"}, `null`, formats.TABLE, apierr.ErrMsgCLIAddingCmdsInReq},
		{"CLI Request with fake(foo) output format", []string{"show version", "show system lldp neighbor", "info from state network-instance default interface system0.0"}, `null`, formats.EnumOutputFormats("foo"), apierr.ErrMsgCLISettingOutFormat},
	}
	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			r, err := srljrpc.NewCLIRequest(td.cmds, td.of)
			checkErrGotVSExp(err, td.expReqErr, t)
			b, err := r.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			expJSON := new(bytes.Buffer)
			expJSON.Grow(512)
			if td.tmplJSON == `null` {
				// null in case of expected error
				expJSON.WriteString(`null`)
			} else {
				// creating template
				tmpl, err := template.New("set").Parse(td.tmplJSON)
				if err != nil {
					t.Fatal(err)
				}
				// embedding ID into template
				err = tmpl.Execute(expJSON, r.ID)
				if err != nil {
					t.Fatal(err)
				}
			}
			// Uncomment for debugging
			// t.Logf("CLI request: %s", string(b))
			// t.Logf("CLI expected: %s", expJSON.String())
			out := cmp.Diff(string(b), expJSON.String())
			if out != "" {
				t.Logf(out)
				t.Errorf("got %s, while should be %s", string(b), expJSON.String())
			}
		})
	}
}

func TestNewRequest_Diff(t *testing.T) {
	// Diff method testing
	cmdArgs := []struct {
		action actions.EnumActions
		path   string
		value  srljrpc.CommandValue
		opts   []srljrpc.CommandOption
	}{
		{actions.UPDATE, "/interface[name=ethernet-1/1]/description", srljrpc.CommandValue("DiffUpdate"), []srljrpc.CommandOption{}},                                                                      // #1 should succeed
		{actions.REPLACE, "/interface[name=ethernet-1/1]/description", srljrpc.CommandValue("DiffReplace"), []srljrpc.CommandOption{}},                                                                    // #2 should succeed
		{actions.DELETE, "/interface[name=ethernet-1/1]/description", srljrpc.CommandValue(""), []srljrpc.CommandOption{}},                                                                                // #3 should succeed
		{actions.DELETE, "/interface[name=ethernet-1/1]/description", srljrpc.CommandValue("DiffDelete"), []srljrpc.CommandOption{}},                                                                      // #4 should fail cause of value not allowed
		{actions.NONE, "/interface[name=ethernet-1/1]/description", srljrpc.CommandValue("test"), []srljrpc.CommandOption{}},                                                                              // #5 should fail due to unsupported action NONE by DIFF
		{actions.UPDATE, "/interface[name=ethernet-1/1]/description", srljrpc.CommandValue("DiffUpdate_with_CANDIDATE"), []srljrpc.CommandOption{srljrpc.WithDatastore(datastores.CANDIDATE)}},            // #6 should succeed with correct datastore for DIFF
		{actions.UPDATE, "", srljrpc.CommandValue("test"), []srljrpc.CommandOption{}},                                                                                                                     // #7 should fail due to empty path
		{actions.UPDATE, "/interface[name=ethernet-1/1]/description", srljrpc.CommandValue(""), []srljrpc.CommandOption{}},                                                                                // #8 should fail due to empty value
		{actions.UPDATE, "/interface[name=ethernet-1/1]/description:DiffUpdate_test", srljrpc.CommandValue(""), []srljrpc.CommandOption{srljrpc.WithAddPathKeywords([]byte(`{"name": "ethernet-1/1"}`))}}, // #9 should succeed, bcz of :test value specified as part of path + check for AddPathKeywords
		{actions.REPLACE, "/interface[name=ethernet-1/1]/description:DiffReplace_test", srljrpc.CommandValue("DiffReplace_TEST"), []srljrpc.CommandOption{}},                                              // #10 should fail, bcz of :test value specified as part of path and value is not empty
		{actions.REPLACE, "/interface[name=ethernet-1/1]/description:DiffReplace_test", srljrpc.CommandValue(""), []srljrpc.CommandOption{srljrpc.WithDefaults(), srljrpc.WithoutRecursion()}},            // #11 should succeed
	}
	cmdResults := []*srljrpc.Command{}

	for _, ca := range cmdArgs {
		t.Run("NewDIFFCommand", func(t *testing.T) {
			cmd, err := srljrpc.NewCommand(ca.action, ca.path, ca.value, ca.opts...)
			if err != nil {
				t.Fatal(err)
			}
			cmdResults = append(cmdResults, cmd)
		})
	}
	m := "diff"
	testData := []struct {
		testName  string
		cmd       *srljrpc.Command
		expReqErr error
		tmplJSON  string
	}{
		{"Basic DIFF UPDATE w/o datastore [exp. SUCC]", cmdResults[0], nil,
			`{"jsonrpc":"2.0","id":{{.}},"method":"diff","params":{"commands":[{"path":"/interface[name=ethernet-1/1]/description","value":"DiffUpdate","action":"update"}]}}`}, // #1
		{"Basic DIFF REPLACE w/o datastore [exp. SUCC]", cmdResults[1], nil,
			`{"jsonrpc":"2.0","id":{{.}},"method":"diff","params":{"commands":[{"path":"/interface[name=ethernet-1/1]/description","value":"DiffReplace","action":"replace"}]}}`}, // #2
		{"Basic DIFF DELETE w/o datastore [exp. SUCC]", cmdResults[2], nil,
			`{"jsonrpc":"2.0","id":{{.}},"method":"diff","params":{"commands":[{"path":"/interface[name=ethernet-1/1]/description","action":"delete"}],"yang-models":"oc"}}`}, // #3
		{"Basic DIFF DELETE with value [exp. FAIL]", cmdResults[3], apierr.ErrMsgReqAddingCmds, `null`},             // #4
		{"Basic DIFF without action [exp. FAIL]", cmdResults[4], apierr.ErrMsgReqAddingCmds, `null`},                // #5
		{"Basic DIFF UPDATE with correct datastore [exp. FAIL]", cmdResults[5], apierr.ErrMsgReqAddingCmds, `null`}, // #6
		{"Basic DIFF UPDATE with empty path [exp. FAIL]", cmdResults[6], apierr.ErrMsgReqAddingCmds, `null`},        // #7
		{"Basic DIFF UPDATE with empty value [exp. FAIL]", cmdResults[7], apierr.ErrMsgReqAddingCmds, `null`},       // #8
		{"Basic DIFF UPDATE with k:v path and value and path keywords [exp. SUCC]", cmdResults[8], nil,
			`{"jsonrpc":"2.0","id":{{.}},"method":"diff","params":{"commands":[{"path":"/interface[name=ethernet-1/1]/description:DiffUpdate_test","path-keywords":{"name":"ethernet-1/1"},"action":"update"}],"yang-models":"oc"}}`}, // #9
		{"Basic DIFF REPLACE with :test value specified as part of path and value is not empty [exp. FAIL]", cmdResults[9], apierr.ErrMsgReqAddingCmds, `null`}, // #10
		{"Basic DIFF REPLACE with w defaults and w/o recursion [exp. SUCC]", cmdResults[10], nil,
			`{"jsonrpc":"2.0","id":{{.}},"method":"diff","params":{"commands":[{"path":"/interface[name=ethernet-1/1]/description:DiffReplace_test","recursive":false,"include-field-defaults":true,"action":"replace"}],"yang-models":"srl"}}`}, // #11
	}
	for i, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			opts := []srljrpc.RequestOption{}
			if i == 2 || i == 8 {
				opts = append(opts, srljrpc.WithYmType(yms.OC)) // for the tests #3&9 specify OC YANG model type
			}
			if i == 10 {
				opts = append(opts, srljrpc.WithYmType(yms.SRL)) // for the tests #11 specify SRL YANG model type explicitly
			}

			r, err := srljrpc.NewRequest(methods.DIFF, []*srljrpc.Command{td.cmd}, opts...)
			checkErrGotVSExp(err, td.expReqErr, t)
			expJSON := new(bytes.Buffer)
			expJSON.Grow(512)
			if td.tmplJSON == `null` {
				// null in case of expected error
				expJSON.WriteString(`null`)
			} else {
				// creating template
				tmpl, err := template.New(m).Parse(td.tmplJSON)
				if err != nil {
					t.Fatal(err)
				}
				// embedding ID into template
				err = tmpl.Execute(expJSON, r.ID)
				if err != nil {
					t.Fatal(err)
				}
			}
			b, err := r.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			// Uncomment for debugging
			// t.Logf("SET request: %s", string(b))
			// t.Logf("SET expected: %s", expJSON.String())
			out := cmp.Diff(string(b), expJSON.String())
			//t.Logf("SET request: %s", string(b))
			if out != "" {
				t.Logf(out)
				t.Errorf("got: %s,\n while should be: %s", string(b), expJSON.String())
			}
		})
	}
}

func TestNewDiffRequest(t *testing.T) {
	//m := "diff"
	testData := []struct {
		testName  string
		delete    []srljrpc.PV
		replace   []srljrpc.PV
		update    []srljrpc.PV
		ym        yms.EnumYmType
		of        formats.EnumOutputFormats
		ds        datastores.EnumDatastores
		expReqErr error
		tmplJSON  string
	}{
		{"VALIDATE Request w/ SRL w/ JSON w/ CANDIDATE",
			[]srljrpc.PV{{"/interface[name=mgmt0]/description", srljrpc.CommandValue("ValueDelete_should_not_trigger_error")}},
			[]srljrpc.PV{{"/interface[name=ethernet-1/1]/subinterface[index=1]/description:MAC-VRF 1 + REPLACED", srljrpc.CommandValue("")}},
			[]srljrpc.PV{{"/interface[name=system0]/description:UPDATED", srljrpc.CommandValue("")}}, yms.SRL, formats.JSON, datastores.CANDIDATE, nil,
			`{"jsonrpc":"2.0","id":{{.}},"method":"diff","params":{"commands":[{"path":"/interface[name=mgmt0]/description","action":"delete"},{"path":"/interface[name=ethernet-1/1]/subinterface[index=1]/description:MAC-VRF 1 + REPLACED","action":"replace"},{"path":"/interface[name=system0]/description:UPDATED","action":"update"}],"output-format":"json","datastore":"candidate","yang-models":"srl"}}`}, // should succeed
		{"VALIDATE Request w/ OC w/ TEXT w/ RUNNING",
			[]srljrpc.PV{{"/interfaces/interface[name=mgmt0]/subinterfaces/subinterface[index=0]", srljrpc.CommandValue("ValueDelete_should_not_trigger_error")}},
			[]srljrpc.PV{{"/interfaces/interface[name=ethernet-1/1]/subinterfaces/subinterface[index=0]/config/description:diff oc test w/o underscore", srljrpc.CommandValue("")}},
			[]srljrpc.PV{{"/interfaces/interface[name=system0]/config/description:UPDATED", srljrpc.CommandValue("")}}, yms.OC, formats.TEXT, datastores.RUNNING, apierr.ErrMsgDSCandidateDiffOnly,
			`null`}, // should fail, bcz of unsupported datastore RUNNING
		{"VALIDATE Request w/ SRL w/ TABLE w/ CANDIDATE",
			[]srljrpc.PV{{"/interfaces/interface[name=mgmt0]/subinterfaces/subinterface[index=0]", srljrpc.CommandValue("ValueDelete_should_not_trigger_error")}},
			[]srljrpc.PV{{"/interfaces/interface[name=ethernet-1/1]/subinterfaces/subinterface[index=0]/config/description:diff oc test w/o underscore", srljrpc.CommandValue("")}},
			[]srljrpc.PV{{"/interfaces/interface[name=system0]/config/description:UPDATED", srljrpc.CommandValue("")}}, yms.OC, formats.TABLE, datastores.CANDIDATE, nil,
			`{"jsonrpc":"2.0","id":{{.}},"method":"diff","params":{"commands":[{"path":"/interfaces/interface[name=mgmt0]/subinterfaces/subinterface[index=0]","action":"delete"},{"path":"/interfaces/interface[name=ethernet-1/1]/subinterfaces/subinterface[index=0]/config/description:diff oc test w/o underscore","action":"replace"},{"path":"/interfaces/interface[name=system0]/config/description:UPDATED","action":"update"}],"output-format":"table","datastore":"candidate","yang-models":"oc"}}`}, // should succeed

	}

	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			r, err := srljrpc.NewDiffRequest(td.delete, td.replace, td.update, td.ym, td.of, td.ds)
			checkErrGotVSExp(err, td.expReqErr, t)
			expJSON := new(bytes.Buffer)
			expJSON.Grow(512)
			if td.tmplJSON == `null` {
				// null in case of expected error
				expJSON.WriteString(`null`)
			} else {
				// creating template
				tmpl, err := template.New("get").Parse(td.tmplJSON)
				if err != nil {
					t.Fatal(err)
				}
				// embedding ID into template
				err = tmpl.Execute(expJSON, r.ID)
				if err != nil {
					t.Fatal(err)
				}
			}
			b, err := r.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			// Uncomment for debugging
			// t.Logf("SET request: %s", string(b))
			// t.Logf("SET expected: %s", expJSON.String())
			out := cmp.Diff(string(b), expJSON.String())
			if out != "" {
				t.Logf(out)
				t.Errorf("got %s, while should be %s", string(b), expJSON.String())
			}
		})
	}
}

func checkErrGotVSExp(err error, expErr error, t *testing.T) {
	switch {
	case err == nil && expErr == nil:
	case err != nil && expErr != nil:
		if !errors.Is(err, expErr) {
			t.Errorf("got: [%s], while should be: [%s]", err, expErr)
		}
	case err == nil && expErr != nil:
		t.Errorf("got: [%v], while should be: [%s]", err, expErr)
	case err != nil && expErr == nil:
		t.Errorf("got: [%s], while should be: [%v]", err, expErr)
	default:
		t.Errorf("unexpected error - got: [%s], while should be: [%s]", err, expErr)
	}
}
