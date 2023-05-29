//go:build !integration

package srljrpc_test

import (
	"fmt"
	"testing"

	"github.com/azyablov/srljrpc/yms"
)

func TestYms(t *testing.T) {
	// Verify default yang models is SRL
	ym := yms.YmType{}
	rym, err := ym.GetYmType()
	// Check if we got an error
	if err != nil {
		t.Errorf("Default yang models should be SRL, but got an error: %s", err)
	}
	// Check if we got the right yang models
	if rym != yms.SRL {
		t.Errorf("Default yang models should be SRL, but got %v", rym)
	}

	// Table driven tests
	var testData = []struct {
		testName  string
		yms       yms.EnumYmType
		expErrSet error
		expErrGet error
		errMsg    string
	}{
		{testName: "Setting yang models to SRL", yms: yms.SRL, expErrSet: nil, expErrGet: nil, errMsg: "yang models SRL isn't set properly: "},
		{testName: "Setting yang models to OC", yms: yms.OC, expErrSet: nil, expErrGet: nil, errMsg: "yang models OC isn't set properly: "},
		{testName: "Setting yang models to non existent option 100", yms: yms.EnumYmType(100), expErrSet: fmt.Errorf(yms.SetErrMsg), expErrGet: nil, errMsg: "fake yang models option 100 was handled incorrectly: "},
	}

	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			y := yms.YmType{}
			errSetYm := y.SetYmType(td.yms)
			switch {
			case errSetYm == nil && td.expErrSet == nil:
			case errSetYm != nil && td.expErrSet != nil:
				if errSetYm.Error() != td.expErrSet.Error() {
					t.Errorf(td.errMsg+"got %s, while should be %s", errSetYm, td.expErrSet)
				}
			case errSetYm == nil && td.expErrSet != nil:
				t.Errorf(td.errMsg+"got %s, while should be %s", errSetYm, td.expErrSet)
			case errSetYm != nil && td.expErrSet == nil:
				t.Errorf(td.errMsg+"got %s, while should be %s", errSetYm, td.expErrSet)
			default:
				t.Errorf(td.errMsg+"got %s, while should be %s", errSetYm, td.expErrSet)
			}

			ry, errGetYm := y.GetYmType()
			switch {
			case errGetYm == nil && td.expErrGet == nil:
				// while SetYmType must failing, GetYmType must not get the same result
				if ry == td.yms && td.expErrSet != nil {
					t.Errorf(td.errMsg+"got %v, while should be %v", ry, td.yms)
				}
				// if SetYmType is ok, then GetYmType must return the same result
				if ry != td.yms && td.expErrSet == nil {
					t.Errorf(td.errMsg+"got %v, while should be %v", ry, td.yms)
				}
			case errGetYm != nil && td.expErrGet != nil:
				if errGetYm.Error() != td.expErrGet.Error() {
					t.Errorf(td.errMsg+"got %s, while should be %s", errGetYm, td.expErrGet)
				}
			case errGetYm == nil && td.expErrGet != nil:
				t.Errorf(td.errMsg+"got %s, while should be %s", errGetYm, td.expErrGet)
			case errGetYm != nil && td.expErrGet == nil:
				t.Errorf(td.errMsg+"got %s, while should be %s", errGetYm, td.expErrGet)
			default:
				t.Errorf(td.errMsg+"got %s, while should be %s", errGetYm, td.expErrGet)
			}
		})
	}
}
