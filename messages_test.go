package srljrpc_test

import (
	"bytes"
	"fmt"
	"html/template"
	"testing"

	"github.com/azyablov/srljrpc"
	"github.com/azyablov/srljrpc/actions"
	"github.com/azyablov/srljrpc/datastores"
	"github.com/azyablov/srljrpc/formats"
	"github.com/azyablov/srljrpc/methods"
	"github.com/google/go-cmp/cmp"
)

func TestNewRequest_Get(t *testing.T) {
	// GET method testing
	cmdArgs := []struct {
		action actions.EnumActions
		path   string
		value  srljrpc.CommandValue
		opts   []srljrpc.CommandOptions
	}{
		{actions.NONE, "/system/name/host-name", srljrpc.CommandValue(""), nil}, // should succeed
		{actions.NONE, "/system/name/host-name", srljrpc.CommandValue(""), []srljrpc.CommandOptions{srljrpc.WithDefaults(), srljrpc.WithoutRecursion(), srljrpc.WithDatastore(datastores.STATE)}}, // should succeed
		{actions.NONE, "/system/name/host-name", srljrpc.CommandValue("shouldFail"), nil},                                                                                                         // should fail, bcz of value
		{actions.DELETE, "/system/name/host-name", srljrpc.CommandValue(""), nil},                                                                                                                 // should fail, bcz of action
		{actions.NONE, "/system/name/host-name", srljrpc.CommandValue(""), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.TOOLS)}},                                                     // should fail, bcz of datastore
		{actions.NONE, "", srljrpc.CommandValue(""), nil},                                                                                                                                         // should fail, bcz of empty path
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

	m := "get"
	testData := []struct {
		testName  string
		cmd       *srljrpc.Command
		expReqErr error
		tmplJSON  string
	}{
		{"Basic GET", cmdResults[0], nil, `{"jsonrpc":"2.0","id":{{.}},"method":"get","params":{"commands":[{"path":"/system/name/host-name"}]}}`},
		{"Basic GET with options", cmdResults[1], nil, `{"jsonrpc":"2.0","id":{{.}},"method":"get","params":{"commands":[{"path":"/system/name/host-name","recursive":false,"include-field-defaults":true,"datastore":"state"}]}}`},
		{"Basic GET with value", cmdResults[2], fmt.Errorf("value not allowed for method %s", m), `null`},
		{"Basic GET with actions", cmdResults[3], fmt.Errorf("action not allowed for method %s", m), `null`},
		{"Basic GET with TOOLS datastore}", cmdResults[4], fmt.Errorf("datastore %s not allowed for method %s", "tools", m), `null`},
		{"Basic GET with empty path", cmdResults[5], fmt.Errorf("path not found, but should be specified for method %s", m), `null`},
	}

	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			r, err := srljrpc.NewRequest(methods.GET, []*srljrpc.Command{td.cmd})
			switch {
			case err == nil && td.expReqErr == nil:
			case err != nil && td.expReqErr != nil:
				if err.Error() != td.expReqErr.Error() {
					t.Errorf("got %s, while should be %s", err, td.expReqErr)
				}
			case err == nil && td.expReqErr != nil:
				t.Errorf("got %s, while should be %s", err, td.expReqErr)
			case err != nil && td.expReqErr == nil:
				t.Errorf("got %s, while should be %s", err, td.expReqErr)
			default:
				t.Errorf("got %s, while should be %s", err, td.expReqErr)
			}
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
		opts   []srljrpc.CommandOptions
	}{
		{actions.UPDATE, "/system/name/host-name", srljrpc.CommandValue("SetUpdate"), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.TOOLS)}},                                                         // should succeed
		{actions.REPLACE, "/system/name/host-name", srljrpc.CommandValue("SetReplace"), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.CANDIDATE)}},                                                   // should succeed
		{actions.DELETE, "/system/name/host-name", srljrpc.CommandValue("SetDelete"), []srljrpc.CommandOptions{srljrpc.WithDefaults(), srljrpc.WithoutRecursion(), srljrpc.WithDatastore(datastores.CANDIDATE)}}, // should succeed
		{actions.DELETE, "/system/name/host-name", srljrpc.CommandValue(""), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.RUNNING)}},                                                                // should be failing due to unsupported datastore by SET
		{actions.NONE, "/system/name/host-name", srljrpc.CommandValue("test"), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.TOOLS)}},                                                                // should be failing due to unsupported action by SET
		{actions.UPDATE, "", srljrpc.CommandValue(""), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.CANDIDATE)}},                                                                                    // should be failing due to empty path
		{actions.UPDATE, "/system/name/host-name", srljrpc.CommandValue(""), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.CANDIDATE)}},                                                              // should be failing due to empty value
		{actions.UPDATE, "/system/name/host-name:test", srljrpc.CommandValue(""), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.CANDIDATE)}},                                                         // should not fail, bcz of :test value specified as part of path
		{actions.REPLACE, "/system/name/host-name:test", srljrpc.CommandValue("TEST"), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.CANDIDATE)}},                                                    // should fail, bcz of :test value specified as part of path and value is not empty
		{actions.REPLACE, "/system/name/host-name:test:TEST", srljrpc.CommandValue(""), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.CANDIDATE)}},                                                   // should not fail, bcz of :test:TEST value specified as part of path
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

	m := "set"
	testData := []struct {
		testName  string
		cmd       *srljrpc.Command
		expReqErr error
		tmplJSON  string
	}{
		{"Basic SET with TOOLS datastore", cmdResults[0], nil, `{"jsonrpc":"2.0","id":{{.}},"method":"set","params":{"commands":[{"path":"/system/name/host-name","value":"SetUpdate","action":"update","datastore":"tools"}]}}`},
		{"Basic SET with CANDIDATE datastore", cmdResults[1], nil, `{"jsonrpc":"2.0","id":{{.}},"method":"set","params":{"commands":[{"path":"/system/name/host-name","value":"SetReplace","action":"replace","datastore":"candidate"}]}}`},
		{"Basic SET with options", cmdResults[2], nil, `{"jsonrpc":"2.0","id":{{.}},"method":"set","params":{"commands":[{"path":"/system/name/host-name","value":"SetDelete","recursive":false,"include-field-defaults":true,"action":"delete","datastore":"candidate"}]}}`},
		{"Basic SET with unsupported datastore RUNNING}", cmdResults[3], fmt.Errorf("datastore running not allowed for method %s", m), `null`},
		{"Basic SET without action", cmdResults[4], fmt.Errorf("action not found, but should be specified for method %s", m), `null`},
		{"Basic SET with empty path", cmdResults[5], fmt.Errorf("path not found, but should be specified for method %s", m), `null`},
		{"Basic SET with empty value", cmdResults[6], fmt.Errorf("value isn't specified or not found in the path for method %s", m), `null`},
		{"Basic SET with k:v path", cmdResults[7], nil, `{"jsonrpc":"2.0","id":{{.}},"method":"set","params":{"commands":[{"path":"/system/name/host-name:test","action":"update","datastore":"candidate"}]}}`},
		{"Basic SET with k:v path and value", cmdResults[8], fmt.Errorf("value specified in the path and as a separate value for method %s", m), `null`},
		{"Basic SET with incorrect k:v path", cmdResults[9], fmt.Errorf("invalid k:v path specification for method %s", m), `null`},
	}

	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			r, err := srljrpc.NewRequest(methods.SET, []*srljrpc.Command{td.cmd})
			switch {
			case err == nil && td.expReqErr == nil:
			case err != nil && td.expReqErr != nil:
				if err.Error() != td.expReqErr.Error() {
					t.Errorf("got %s, while should be %s", err, td.expReqErr)
				}
			case err == nil && td.expReqErr != nil:
				t.Errorf("got %s, while should be %s", err, td.expReqErr)
			case err != nil && td.expReqErr == nil:
				t.Errorf("got %s, while should be %s", err, td.expReqErr)
			default:
				t.Errorf("got %s, while should be %s", err, td.expReqErr)
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

func TestNewRequest_Validate(t *testing.T) {
	// VALIDATE method testing
	// SET method testing
	cmdArgs := []struct {
		action actions.EnumActions
		path   string
		value  srljrpc.CommandValue
		opts   []srljrpc.CommandOptions
	}{
		{actions.UPDATE, "/system/name/host-name", srljrpc.CommandValue("ValidateUpdate"), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.CANDIDATE)}},                                                     // should succeed
		{actions.REPLACE, "/system/name/host-name", srljrpc.CommandValue("ValidateReplace"), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.CANDIDATE)}},                                                   // should succeed
		{actions.DELETE, "/system/name/host-name", srljrpc.CommandValue("ValidateDelete"), []srljrpc.CommandOptions{srljrpc.WithDefaults(), srljrpc.WithoutRecursion(), srljrpc.WithDatastore(datastores.CANDIDATE)}}, // should succeed
		{actions.DELETE, "/system/name/host-name", srljrpc.CommandValue(""), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.RUNNING)}},                                                                     // should be failing due to unsupported datastore by VALIDATE
		{actions.NONE, "/system/name/host-name", srljrpc.CommandValue("test"), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.CANDIDATE)}},                                                                 // should be failing due to unsupported action by VALIDATE
		{actions.UPDATE, "", srljrpc.CommandValue(""), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.CANDIDATE)}},                                                                                         // should be failing due to empty path
		{actions.UPDATE, "/system/name/host-name", srljrpc.CommandValue(""), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.CANDIDATE)}},                                                                   // should be failing due to empty value
		{actions.UPDATE, "/system/name/host-name:test", srljrpc.CommandValue(""), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.CANDIDATE)}},                                                              // should not fail, bcz of :test value specified as part of path
		{actions.REPLACE, "/system/name/host-name:test", srljrpc.CommandValue("TEST"), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.CANDIDATE)}},                                                         // should fail, bcz of :test value specified as part of path and value is not empty
		{actions.REPLACE, "/system/name/host-name:test:TEST", srljrpc.CommandValue(""), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.CANDIDATE)}},                                                        // should not fail, bcz of :test:TEST value specified as part of path
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

	m := "validate"
	testData := []struct {
		testName  string
		cmd       *srljrpc.Command
		expReqErr error
		tmplJSON  string
	}{
		{"Basic VALIDATE with TOOLS datastore", cmdResults[0], nil, `{"jsonrpc":"2.0","id":{{.}},"method":"validate","params":{"commands":[{"path":"/system/name/host-name","value":"ValidateUpdate","action":"update","datastore":"candidate"}]}}`},
		{"Basic VALIDATE with CANDIDATE datastore", cmdResults[1], nil, `{"jsonrpc":"2.0","id":{{.}},"method":"validate","params":{"commands":[{"path":"/system/name/host-name","value":"ValidateReplace","action":"replace","datastore":"candidate"}]}}`},
		{"Basic VALIDATE with options", cmdResults[2], nil, `{"jsonrpc":"2.0","id":{{.}},"method":"validate","params":{"commands":[{"path":"/system/name/host-name","value":"ValidateDelete","recursive":false,"include-field-defaults":true,"action":"delete","datastore":"candidate"}]}}`},
		{"Basic VALIDATE with unsupported datastore RUNNING}", cmdResults[3], fmt.Errorf("datastore running not allowed for method %s", m), `null`},
		{"Basic VALIDATE without action", cmdResults[4], fmt.Errorf("action not found, but should be specified for method %s", m), `null`},
		{"Basic VALIDATE with empty path", cmdResults[5], fmt.Errorf("path not found, but should be specified for method %s", m), `null`},
		{"Basic VALIDATE with empty value", cmdResults[6], fmt.Errorf("value isn't specified or not found in the path for method %s", m), `null`},
		{"Basic VALIDATE with k:v path", cmdResults[7], nil, `{"jsonrpc":"2.0","id":{{.}},"method":"validate","params":{"commands":[{"path":"/system/name/host-name:test","action":"update","datastore":"candidate"}]}}`},
		{"Basic VALIDATE with k:v path and value", cmdResults[8], fmt.Errorf("value specified in the path and as a separate value for method %s", m), `null`},
		{"Basic VALIDATE with incorrect k:v path", cmdResults[9], fmt.Errorf("invalid k:v path specification for method %s", m), `null`},
	}

	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			r, err := srljrpc.NewRequest(methods.VALIDATE, []*srljrpc.Command{td.cmd})
			switch {
			case err == nil && td.expReqErr == nil:
			case err != nil && td.expReqErr != nil:
				if err.Error() != td.expReqErr.Error() {
					t.Errorf("got %s, while should be %s", err, td.expReqErr)
				}
			case err == nil && td.expReqErr != nil:
				t.Errorf("got %s, while should be %s", err, td.expReqErr)
			case err != nil && td.expReqErr == nil:
				t.Errorf("got %s, while should be %s", err, td.expReqErr)
			default:
				t.Errorf("got %s, while should be %s", err, td.expReqErr)
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

func TestNewCLIRequest(t *testing.T) {
	testData := []struct {
		testName string
		cmds     []string
		tmplJSON string
		of       formats.EnumOutputFormats
		expErr   error
	}{
		{"CLI Request with JSON output format", []string{"show version", "show system lldp neighbor", "info from state network-instance default interface system0.0"}, `{"jsonrpc":"2.0","id":{{.}},"method":"cli","params":{"commands":["show version","show system lldp neighbor","info from state network-instance default interface system0.0"],"output-format":"json"}}`, formats.JSON, nil},
		{"CLI Request with TEXT output format", []string{"show version", "show system lldp neighbor", "info from state network-instance default interface system0.0"}, `{"jsonrpc":"2.0","id":{{.}},"method":"cli","params":{"commands":["show version","show system lldp neighbor","info from state network-instance default interface system0.0"],"output-format":"text"}}`, formats.TEXT, nil},
		{"CLI Request with TABLE output format", []string{"show version", "show system lldp neighbor", "info from state network-instance default interface system0.0"}, `{"jsonrpc":"2.0","id":{{.}},"method":"cli","params":{"commands":["show version","show system lldp neighbor","info from state network-instance default interface system0.0"],"output-format":"table"}}`, formats.TABLE, nil},
		{"CLI Request with empty command", []string{"show version", "", "info from state network-instance default interface system0.0"}, `null`, formats.TABLE, fmt.Errorf("empty commands are not allowed")},
		{"CLI Request with fake(100) output format", []string{"show version", "show system lldp neighbor", "info from state network-instance default interface system0.0"}, `null`, formats.EnumOutputFormats(100), fmt.Errorf(formats.SetErrMsg)},
	}
	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			r, err := srljrpc.NewCLIRequest(td.cmds, td.of)
			if err != nil {
				t.Logf("NewCLIRequest: %s", err)
			}
			switch {
			case err == nil && td.expErr == nil:
			case err != nil && td.expErr != nil:
				if err.Error() != td.expErr.Error() {
					t.Errorf("got %s, while should be %s", err, td.expErr)
				}
			case err == nil && td.expErr != nil:
				t.Errorf("got %s, while should be %s", err, td.expErr)
			case err != nil && td.expErr == nil:
				t.Errorf("got %s, while should be %s", err, td.expErr)
			default:
				t.Errorf("got %s, while should be %s", err, td.expErr)
			}
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
