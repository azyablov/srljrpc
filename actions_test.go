//go:build unit

package srljrpc_test

import (
	"fmt"
	"testing"

	"github.com/azyablov/srljrpc/actions"
)

func TestActions(t *testing.T) {

	// Verify default action is INVALID_ACTION
	a := actions.Action{}
	ra, err := a.GetAction()
	if err != nil || ra != actions.NONE {
		t.Errorf("Default action should be NONE, but got %v", ra)
	}

	// Table driven tests
	var testData = []struct {
		testName  string
		action    actions.EnumActions
		expErrSet error
		expErrGet error
		errMsg    string
	}{
		{testName: "Setting action to REPLACE", action: actions.REPLACE, expErrSet: nil, expErrGet: nil, errMsg: "action REPLACE isn't set properly: "},
		{testName: "Setting action to UPDATE", action: actions.UPDATE, expErrSet: nil, expErrGet: nil, errMsg: "action UPDATE isn't set properly: "},
		{testName: "Setting action to DELETE", action: actions.DELETE, expErrSet: nil, expErrGet: nil, errMsg: "action DELETE isn't set properly: "},
		{testName: "Setting action to NONE", action: actions.NONE, expErrSet: nil, expErrGet: nil, errMsg: "action NONE isn't set properly: "},
		{testName: "Setting action to INVALID_ACTION", action: actions.INVALID_ACTION, expErrSet: fmt.Errorf(actions.SetErrMsg), expErrGet: nil, errMsg: "action INVALID_ACTION was handled incorrectly: "},
		{testName: "Setting action to non existent action 100", action: actions.EnumActions(100), expErrSet: fmt.Errorf(actions.SetErrMsg), expErrGet: nil, errMsg: "fake action 100 was handled incorrectly: "},
	}
	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			a := actions.Action{}
			err := a.SetAction(td.action)
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

			ra, err := a.GetAction()
			switch {
			case err == nil && td.expErrGet == nil:
				// while SetAction must failing, GetAction must not get the same result
				if ra == td.action && td.expErrSet != nil {
					t.Errorf(td.errMsg+"got %v, while should be %v", ra, td.action)
				}
				// if SetAction is ok, then GetAction must return the same result
				if ra != td.action && td.expErrSet == nil {
					t.Errorf(td.errMsg+"got %v, while should be %v", ra, td.action)
				}
			case err != nil && td.expErrGet != nil:
				if err.Error() != td.expErrGet.Error() {
					errStr := fmt.Sprintf(td.errMsg+"got %s, while should be %s", err, td.expErrGet)
					if ra != actions.INVALID_ACTION {
						errStr = fmt.Sprintf(errStr+"action expected %v, but got action %v", td.action, ra)
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
