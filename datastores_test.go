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
		errMsg    string
	}{
		{testName: "Setting datastore to CANDIDATE", datastore: datastores.CANDIDATE, expErrSet: nil, expErrGet: nil, errMsg: "datastore CANDIDATE isn't set properly: "},
		{testName: "Setting datastore to RUNNING", datastore: datastores.RUNNING, expErrSet: nil, expErrGet: nil, errMsg: "datastore RUNNING isn't set properly: "},
		{testName: "Setting datastore to STATE", datastore: datastores.STATE, expErrSet: nil, expErrGet: nil, errMsg: "datastore STATE isn't set properly: "},
		{testName: "Setting datastore to TOOLS", datastore: datastores.TOOLS, expErrSet: nil, expErrGet: nil, errMsg: "datastore TOOLS isn't set properly: "},
		{testName: "Setting datastore to non existent datastore 100", datastore: datastores.EnumDatastores(100), expErrSet: fmt.Errorf(datastores.SetErrMsg),
			expErrGet: nil, errMsg: "fake datastore 100 was handled incorrectly: "},
	}

	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			d := datastores.Datastore{}
			err := d.SetDatastore(td.datastore)
			switch {
			case err == nil && td.expErrSet == nil:
			case err != nil && td.expErrSet != nil:
				if err.Error() != td.expErrSet.Error() {
					t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErrSet)
				}
			case err == nil && td.expErrSet != nil:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErrSet)
			case err != nil && td.expErrSet == nil:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErrSet)
			default:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErrSet)
			}

			rd, err := d.GetDatastore()
			switch {
			case err == nil && td.expErrGet == nil:
				// while SetDatastore must failing, GetDatastore must not get the same result
				if rd == td.datastore && td.expErrSet != nil {
					t.Errorf(td.errMsg+"got %v, while should be %v", rd, td.datastore)
				}
				// if SetDatastore is ok, then GetDatastore must return the same result
				if rd != td.datastore && td.expErrSet == nil {
					t.Errorf(td.errMsg+"got %v, while should be %v", rd, td.datastore)
				}
			case err != nil && td.expErrGet != nil:
				if err.Error() != td.expErrGet.Error() {
					t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErrGet)
				}
			case err == nil && td.expErrGet != nil:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErrGet)
			case err != nil && td.expErrGet == nil:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErrGet)
			default:
				t.Errorf(td.errMsg+"got %s, while should be %s", err, td.expErrGet)
			}
		})
	}
}
