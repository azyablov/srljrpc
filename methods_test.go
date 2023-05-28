//go:build !integration

package srljrpc_test

import (
	"fmt"
	"testing"

	"github.com/azyablov/srljrpc/methods"
)

func TestMethods(t *testing.T) {

	// Verify default method is INVALID_METHOD
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
	}{
		{testName: "Setting method to CLI", method: methods.CLI, expErrSet: nil, expErrGet: nil},
		{testName: "Setting method to GET", method: methods.GET, expErrSet: nil, expErrGet: nil},
		{testName: "Setting method to SET", method: methods.SET, expErrSet: nil, expErrGet: nil},
		{testName: "Setting method to VALIDATE", method: methods.VALIDATE, expErrSet: nil, expErrGet: nil},
		{testName: "Setting method to DIFF", method: methods.DIFF, expErrSet: nil, expErrGet: nil},
		{testName: "Setting method to INVALID_METHOD", method: methods.INVALID_METHOD, expErrSet: fmt.Errorf(methods.SetErrMsg), expErrGet: fmt.Errorf(methods.GetErrMsg)},
		{testName: "Setting method to 100", method: methods.EnumMethods(100), expErrSet: fmt.Errorf(methods.SetErrMsg), expErrGet: fmt.Errorf(methods.GetErrMsg)},
	}
	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			m := methods.Method{}
			errSetMtd := m.SetMethod(td.method)
			switch {
			case errSetMtd == nil && td.expErrSet == nil:
			case errSetMtd != nil && td.expErrSet != nil:
				if errSetMtd.Error() != td.expErrSet.Error() {
					t.Errorf("got: %s, while should be: %s", errSetMtd, td.expErrSet)
				}
			case errSetMtd == nil && td.expErrSet != nil:
				t.Errorf("got: %s, while should be: %s", errSetMtd, td.expErrSet)
			case errSetMtd != nil && td.expErrSet == nil:
				t.Errorf("got: %s, while should be: %s", errSetMtd, td.expErrSet)
			default:
				t.Errorf("got: %s, while should be: %s", errSetMtd, td.expErrSet)
			}

			rm, errGetMtd := m.GetMethod()
			switch {
			case errGetMtd == nil && td.expErrGet == nil:
				// While SetMethod must failing, GetMethod must not get the same result
				if rm == td.method && td.expErrSet != nil {
					t.Errorf("got: %v, while should be: %v", rm, td.method)
				}
				// If SetMethod is ok, then GetMethod must return the same result
				if rm != td.method && td.expErrSet == nil {
					t.Errorf("got: %v, while should be: %v", rm, td.method)
				}
			case errGetMtd != nil && td.expErrGet != nil:
				if errGetMtd.Error() != td.expErrGet.Error() {
					errStr := fmt.Sprintf("got: %s, while should be: %s", errGetMtd, td.expErrGet)
					if rm != methods.INVALID_METHOD {
						errStr = fmt.Sprintf("method expected %v, but got action %v", td.method, rm)
					}
					t.Errorf(errStr)
				}
			case errGetMtd == nil && td.expErrGet != nil:
				t.Errorf("got: %v, while should be: %v", errGetMtd, td.expErrGet)
			case errGetMtd != nil && td.expErrGet == nil:
				t.Errorf("got: %s, while should be: %v", errGetMtd, td.expErrGet)
			default:
				t.Errorf("got: %s, while should be: %s", errGetMtd, td.expErrGet)
			}
		})
	}
}
