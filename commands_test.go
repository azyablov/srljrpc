package srljrpc

import (
	"encoding/json"
	"testing"

	"github.com/azyablov/srljrpc/actions"
	"github.com/azyablov/srljrpc/datastores"
	"github.com/google/go-cmp/cmp"
)

func Test_rawSetCommand(t *testing.T) {
	var mockSetCmd = &Command{
		Path:      "/interface[name={name}]/description",
		Value:     "This is a test.",
		Action:    &actions.Action{},
		Datastore: &datastores.Datastore{},
	}
	var expected = []byte(`{
T:  "path": "/interface[name={name}]/description",
T:  "value": "This is a test.",
T:  "path-keywords": {
T:    "name": "mgmt0"
T:  },
T:  "include-field-defaults": true,
T:  "action": "update",
T:  "datastore": "candidate"
T:}`)

	err := mockSetCmd.withDatastore(datastores.CANDIDATE)
	if err != nil {
		t.Fatal(err)
	}

	mockSetCmd.withoutRecursion()
	mockSetCmd.withDefaults()

	err = mockSetCmd.withPathKeywords([]byte(`{"name": "mgmt0"}`))
	if err != nil {
		t.Fatal(err)
	}

	err = mockSetCmd.SetAction(actions.UPDATE)
	if err != nil {
		t.Fatal(err)
	}

	b, err := json.MarshalIndent(mockSetCmd, "T:", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if out := cmp.Diff(string(b), string(expected)); out != "" {
		t.Log(out)
		t.Fatalf("expected: %s, got: %s", string(expected), string(b))
	}
}

func Test_rawGetCommand(t *testing.T) {
	var mockGetCmd = &Command{
		Path:      "/interface[name={name}]/description",
		Datastore: &datastores.Datastore{},
	}
	var expected = []byte(`{
T:  "path": "/interface[name={name}]/description",
T:  "path-keywords": {
T:    "name": "mgmt0"
T:  },
T:  "datastore": "running"
T:}`)

	err := mockGetCmd.withDatastore(datastores.RUNNING)
	if err != nil {
		t.Fatal(err)
	}

	err = mockGetCmd.SetAction(actions.NONE)
	t.Logf("expected error: %s", err)
	if err == nil {
		t.Fatalf("expected error '%s', got nil", actions.NoneErrMsg)
	}

	err = mockGetCmd.withPathKeywords([]byte(`{"name": "mgmt0"}`))
	if err != nil {
		t.Fatal(err)
	}

	b, err := json.MarshalIndent(mockGetCmd, "T:", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if out := cmp.Diff(string(b), string(expected)); out != "" {
		t.Fatalf("\nexpected: %s, \ngot: %s", string(expected), string(b))
	}
}

func Test_withPathKeyword(t *testing.T) {
	var mockGetCmd = &Command{
		Path:      "/interface[name={name}]/description",
		Datastore: &datastores.Datastore{},
	}

	err := mockGetCmd.withPathKeywords([]byte(`{"name": "mgmt0", "name1": "mgmt1", "name2" "mgmt2"}`))
	t.Logf("expected error: %s", err)
	if err == nil {
		t.Fatalf("expected error 'failed to unmarshal path-keywords', got nil")
	}
}
