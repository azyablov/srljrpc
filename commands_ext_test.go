package srljrpc_test

import (
	"encoding/json"
	"testing"

	"github.com/azyablov/srljrpc"
	"github.com/azyablov/srljrpc/actions"
	"github.com/azyablov/srljrpc/datastores"
	"github.com/google/go-cmp/cmp"
)

func TestNewCommand(t *testing.T) {
	// checking for the system version and hostname
	strOpts := []string{"", "WithoutRecursion", "WithDefaults", "WithAddPathKeywords", "WithDatastore"}
	cOpts := []srljrpc.CommandOptions{nil, srljrpc.WithoutRecursion(), srljrpc.WithDefaults(), srljrpc.WithAddPathKeywords(json.RawMessage(`{"name": "mgmt0"}`)), srljrpc.WithDatastore(datastores.CANDIDATE)}
	expectedJSON := [][]string{
		{`{"path":"/system/name/host-name","recursive":true}`, `{"path":"/system/name/host-name","value":"test delete","recursive":true,"action":"delete"}`, `{"path":"/system/name/host-name","value":"test update","recursive":true,"action":"update"}`, `{"path":"/system/name/host-name","value":"test replace","recursive":true,"action":"replace"}`},
		{`{"path":"/system/name/host-name"}`, `{"path":"/system/name/host-name","value":"test delete","action":"delete"}`, `{"path":"/system/name/host-name","value":"test update","action":"update"}`, `{"path":"/system/name/host-name","value":"test replace","action":"replace"}`},
		{`{"path":"/system/name/host-name","include-field-defaults":true}`, `{"path":"/system/name/host-name","value":"test delete","include-field-defaults":true,"action":"delete"}`, `{"path":"/system/name/host-name","value":"test update","include-field-defaults":true,"action":"update"}`, `{"path":"/system/name/host-name","value":"test replace","include-field-defaults":true,"action":"replace"}`},
		{`{"path":"/system/name/host-name","path-keywords":{"name":"mgmt0"},"include-field-defaults":true}`, `{"path":"/system/name/host-name","value":"test delete","path-keywords":{"name":"mgmt0"},"include-field-defaults":true,"action":"delete"}`, `{"path":"/system/name/host-name","value":"test update","path-keywords":{"name":"mgmt0"},"include-field-defaults":true,"action":"update"}`, `{"path":"/system/name/host-name","value":"test replace","path-keywords":{"name":"mgmt0"},"include-field-defaults":true,"action":"replace"}`},
	}
	o := strOpts[0]
	for i := 1; i < len(cOpts); i++ {
		// Table driven tests
		testData := []struct {
			testName string
			action   actions.EnumActions
			value    srljrpc.CommandValue
			opts     []srljrpc.CommandOptions
			errExp   error
			expJSON  string
		}{
			{"NONE" + o, actions.NONE, srljrpc.CommandValue(""), cOpts[:i], nil, expectedJSON[i-1][0]},
			{"DELETE" + o, actions.DELETE, srljrpc.CommandValue("test delete"), cOpts[:i], nil, expectedJSON[i-1][1]},
			{"UPDATE" + o, actions.UPDATE, srljrpc.CommandValue("test update"), cOpts[:i], nil, expectedJSON[i-1][2]},
			{"REPLACE" + o, actions.REPLACE, srljrpc.CommandValue("test replace"), cOpts[:i], nil, expectedJSON[i-1][3]},
		}

		for _, td := range testData {
			t.Run(td.testName, func(t *testing.T) {
				cmd, err := srljrpc.NewCommand(td.action, "/system/name/host-name", td.value, td.opts...)
				if err != nil {
					t.Fatal(err)
				}
				b, err := json.Marshal(cmd)
				if err != nil {
					t.Fatal(err)
				}
				t.Log(string(b))
				cmp.Diff(string(b), td.expJSON)
				if out := cmp.Diff(string(b), td.expJSON); out != "" {
					t.Fatalf("\nexpected: %s, \ngot: %s", string(td.expJSON), string(b))
				}
			})
		}
		o = o + " " + strOpts[i]
	}

}
