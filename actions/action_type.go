package actions

import (
	"fmt"
)

// EnumActions is an enumeration type of supported method actions.
type EnumActions int

//	Valid enumeration EnumActions:
//		REPLACE - Replaces the entire configuration within a specific context with the supplied configuration; equivalent to a delete/update. When the action command is used with the tools datastore, update is the only supported option.
//		UPDATE - Updates a leaf or container with the specified value.
//		DELETE - Deletes a leaf or container. All children beneath the parent are removed from the system.
//
// Additional actions to required for the correct handling of the request:
// INVALID_ACTION - at the time of object creation, the action is not set.
// NONE - used for GET method, where the action must not be specified.
const (
	INVALID_ACTION EnumActions = iota
	REPLACE        EnumActions = iota + 1
	UPDATE
	DELETE
	NONE
)

// Error messages for the Action class.
const (
	GetErrMsg = "action isn't set properly, while should be REPLACE / UPDATE / DELETE or NONE for method GET"
	SetErrMsg = "action provided isn't correct, while should be REPLACE / UPDATE / DELETE or NONE for method GET"
)

// Action use is conditional mandatory; used with the set and validate methods only.
type Action struct {
	Action string `json:"action,omitempty"`
}

// GetAction returns the action type and non nil error if the action is not set properly.
func (a *Action) GetAction() (EnumActions, error) {
	var ra EnumActions
	switch a.Action {
	case "replace":
		ra = REPLACE
	case "update":
		ra = UPDATE
	case "delete":
		ra = DELETE
	case "":
		ra = NONE
	default:
		return ra, fmt.Errorf(GetErrMsg)
	}
	return ra, nil
}

// SetAction sets the action type and non nil error if provided action is not correct.
func (a *Action) SetAction(ra EnumActions) error {
	switch ra {
	case DELETE:
		a.Action = "delete"
	case REPLACE:
		a.Action = "replace"
	case UPDATE:
		a.Action = "update"
	case NONE:
		a.Action = ""
	default:
		return fmt.Errorf(SetErrMsg)
	}
	return nil
}
