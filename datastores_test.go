//go:build !integration

package srljrpc_test

import (
	"fmt"
	"testing"

	"github.com/azyablov/srljrpc/datastores"
)

func TestDatastores(t *testing.T) {
	// Verify default datastore is CANDIDATE
	d := datastores.Datastore{}
	rd, err := d.GetDatastore()
	if err != nil || rd != datastores.CANDIDATE {
		t.Errorf("Default datastore should be CANDIDATE, but got %v", rd)
	}

	// Table driven tests
	var testData = []struct {
		testName  string
		datastore datastores.EnumDatastores
		expErrSet error
		expErrGet error
	}{
		{testName: "Setting datastore to CANDIDATE", datastore: datastores.CANDIDATE, expErrSet: nil, expErrGet: nil}, // should succeed
		{testName: "Setting datastore to RUNNING", datastore: datastores.RUNNING, expErrSet: nil, expErrGet: nil},     // should succeed
		{testName: "Setting datastore to STATE", datastore: datastores.STATE, expErrSet: nil, expErrGet: nil},         // should succeed
		{testName: "Setting datastore to TOOLS", datastore: datastores.TOOLS, expErrSet: nil, expErrGet: nil},         // should succeed
		{testName: "Setting datastore to non existent datastore 100", datastore: datastores.EnumDatastores(100), expErrSet: fmt.Errorf(datastores.SetErrMsg), // should fail, unsupported datastore
			expErrGet: nil},
	}

	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			d := datastores.Datastore{}
			errSetDs := d.SetDatastore(td.datastore)
			switch {
			case errSetDs == nil && td.expErrSet == nil:
			case errSetDs != nil && td.expErrSet != nil:
				if errSetDs.Error() != td.expErrSet.Error() {
					t.Errorf("got: %s, while should be: %s", errSetDs, td.expErrSet)
				}
			case errSetDs == nil && td.expErrSet != nil:
				t.Errorf("got: %v, while should be: %s", errSetDs, td.expErrSet)
			case errSetDs != nil && td.expErrSet == nil:
				t.Errorf("got: %s, while should be: %v", errSetDs, td.expErrSet)
			default:
				t.Errorf("got %s, while should be %s", errSetDs, td.expErrSet)
			}

			rd, errGetDs := d.GetDatastore()
			switch {
			case errGetDs == nil && td.expErrGet == nil:
				// while SetDatastore must failing, GetDatastore must not get the same result
				if rd == td.datastore && td.expErrSet != nil {
					t.Errorf("got %v, while should be %v", rd, td.datastore)
				}
				// if SetDatastore is ok, then GetDatastore must return the same result
				if rd != td.datastore && td.expErrSet == nil {
					t.Errorf("got %v, while should be %v", rd, td.datastore)
				}
			case errGetDs != nil && td.expErrGet != nil:
				if errGetDs.Error() != td.expErrGet.Error() {
					t.Errorf("got %s, while should be %s", errGetDs, td.expErrGet)
				}
			case errGetDs == nil && td.expErrGet != nil:
				t.Errorf("got %s, while should be %s", errGetDs, td.expErrGet)
			case errGetDs != nil && td.expErrGet == nil:
				t.Errorf("got %s, while should be %s", errGetDs, td.expErrGet)
			default:
				t.Errorf("got %s, while should be %s", errGetDs, td.expErrGet)
			}
		})
	}
}
