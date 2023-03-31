package srljrpc_test

import (
	"fmt"
	"testing"

	"github.com/azyablov/srljrpc/formats"
)

func TestFormats(t *testing.T) {
	// Verify default format is JSON
	of := formats.OutputFormat{}
	rof, err := of.GetFormat()
	if err != nil || rof != formats.JSON {
		t.Errorf("Default format should be JSON, but got %v", rof)
	}

	// Table driven tests
	var testData = []struct {
		testName  string
		format    formats.EnumOutputFormats
		expErrSet error
		expErrGet error
		errMsg    string
	}{
		{testName: "Setting format to JSON", format: formats.JSON, expErrSet: nil, expErrGet: nil, errMsg: "format JSON isn't set properly: "},
		{testName: "Setting format to XML", format: formats.XML, expErrSet: nil, expErrGet: nil, errMsg: "format XML isn't set properly: "},
		{testName: "Setting format to TABLE", format: formats.TABLE, expErrSet: nil, expErrGet: nil, errMsg: "format TABLE isn't set properly: "},
		{testName: "Setting format to non existent format 100", format: formats.EnumOutputFormats(100), expErrSet: fmt.Errorf(formats.SetErrMsg),
			expErrGet: nil, errMsg: "fake format 100 was handled incorrectly: "},
	}

	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			of := formats.OutputFormat{}
			err := of.SetFormat(td.format)
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

			rof, err := of.GetFormat()
			switch {
			case err == nil && td.expErrGet == nil:
				// while SetFormat must failing, GetFormat must not get the same result
				if rof == td.format && td.expErrSet != nil {
					t.Errorf(td.errMsg+"got %v, while should be %v", rof, td.format)
				}
				// if SetFormat is ok, then GetFormat must return the same result
				if rof != td.format && td.expErrSet == nil {
					t.Errorf(td.errMsg+"got %v, while should be %v", rof, td.format)
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
