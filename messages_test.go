package srljrpc_test

import (
	"fmt"
	"testing"

	"github.com/azyablov/srljrpc"
	"github.com/azyablov/srljrpc/actions"
	"github.com/azyablov/srljrpc/datastores"
	"github.com/azyablov/srljrpc/methods"
)

func TestNewRequest(t *testing.T) {
	// GET method testing
	cmdArgs := []struct {
		action actions.EnumActions
		path   string
		value  srljrpc.CommandValue
		opts   []srljrpc.CommandOptions
	}{
		{actions.NONE, "/system/name/host-name", srljrpc.CommandValue(""), nil},
		{actions.NONE, "/system/name/host-name", srljrpc.CommandValue("shouldFail"), nil},
		{actions.DELETE, "/system/name/host-name", srljrpc.CommandValue(""), nil},
		{actions.NONE, "/system/name/host-name", srljrpc.CommandValue(""), []srljrpc.CommandOptions{srljrpc.WithDatastore(datastores.TOOLS)}},
		{actions.NONE, "", srljrpc.CommandValue(""), nil},
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
		testName     string
		cmd          *srljrpc.Command
		expReqErr    error
		expectedJSON string
	}{
		{"Basic GET", cmdResults[0], nil, `{"jsonrpc":"2.0","id":1,"method":"get","params":{"commands":[{"path":"/system/name/host-name"}]}}`},
		{"Basic GET with value", cmdResults[1], fmt.Errorf("value not allowed for method %s", m), ``},
		{"Basic GET with actions", cmdResults[2], fmt.Errorf("action not allowed for method %s", m), ``},
		{"Basic GET with TOOLS datastore}", cmdResults[3], fmt.Errorf("datastore %s not allowed for method %s", "tools", m), ``},
		{"Basic GET with empty path", cmdResults[4], fmt.Errorf("path not found, but should be specified for method %s", m), ``},
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
			b, err := r.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("GET request: %v", string(b))
		})
	}
}

func TestNewCLIRequest(t *testing.T) {
	//TODO
}
