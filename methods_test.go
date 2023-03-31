package srljrpc_test

import (
	"fmt"
	"testing"

	"github.com/azyablov/srljrpc/actions"
	"github.com/azyablov/srljrpc/methods"
)

func TestMethods(t *testing.T) {

	// Verify default action is INVALID_ACTION
	m := methods.Method{}
	rm, err := m.GetMethod()
	if err == nil || rm != methods.INVALID_METHOD {
		t.Errorf("Default action should be INVALID_ACTION, but got %v", rm)
	}

	// Table driven tests
	var testData = []struct {
		testName  string
		method    methods.EnumMethods
		expErrSet error
		expErrGet error
		errMsg    string
	}{
		{testName: "Setting method to CLI", method: methods.CLI, expErrSet: nil, expErrGet: nil, errMsg: "method CLI isn't set properly: "},
		{testName: "Setting method to GET", method: methods.GET, expErrSet: nil, expErrGet: nil, errMsg: "method GET isn't set properly: "},
		{testName: "Setting method to SET", method: methods.SET, expErrSet: nil, expErrGet: nil, errMsg: "method SET isn't set properly: "},
		{testName: "Setting method to VALIDATE", method: methods.VALIDATE, expErrSet: nil, expErrGet: nil, errMsg: "method VALIDATE isn't set properly: "},
		{testName: "Setting method to INVALID_METHOD", method: methods.INVALID_METHOD, expErrSet: fmt.Errorf(methods.SetErrMsg), expErrGet: fmt.Errorf(methods.GetErrMsg), errMsg: "method INVALID_METHOD was handled incorrectly: "},
		{testName: "Setting method to 100", method: methods.EnumMethods(100), expErrSet: fmt.Errorf(methods.SetErrMsg), expErrGet: fmt.Errorf(methods.GetErrMsg), errMsg: "fake method 100 was handled incorrectly: "},
	}
	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			m := methods.Method{}
			err := m.SetMethod(td.method)
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

			rm, err := m.GetMethod()
			switch {
			case err == nil && td.expErrGet == nil:
				// while SetMethod must failing, GetMethod must not get the same result
				if rm == td.method && td.expErrSet != nil {
					t.Errorf(td.errMsg+"got %v, while should be %v", rm, td.method)
				}
				// if SetMethod is ok, then GetMethod must return the same result
				if rm != td.method && td.expErrSet == nil {
					t.Errorf(td.errMsg+"got %v, while should be %v", rm, td.method)
				}
			case err != nil && td.expErrGet != nil:
				if err.Error() != td.expErrGet.Error() {
					errStr := fmt.Sprintf(td.errMsg+"got %s, while should be %s", err, td.expErrGet)
					if rm != actions.INVALID_ACTION {
						errStr = fmt.Sprintf(errStr+"action expected %v, but got action %v", td.method, rm)
					}
					t.Errorf(errStr)
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
