package actions

import (
	"fmt"
)

type EnumActions int

//	class EnumActions {
//		<<enumeration>>
//		note "Replaces the entire configuration within a specific context with the supplied configuration; equivalent to a delete/update. When the action command is used with the tools datastore, update is the only supported option."
//		REPLACE
//		note "Updates a leaf or container with the specified value."
//		UPDATE
//		note "Deletes a leaf or container. All children beneath the parent are removed from the system."
//		DELETE
//	}
//
// EnumActions "1" --o Action: OneOf
const (
	INVALID_ACTION             = iota
	REPLACE        EnumActions = iota + 1
	UPDATE
	DELETE
)

// note for action "Conditional mandatory; used with the set and validate methods."
//
//	class Action {
//		<<element>>
//		~GetAction(): EnumActions
//		~SetAction(a: EnumActions): error
//		+string Action
//	}
type Action struct {
	Action string `json:"action"`
}

func (a *Action) GetAction() (EnumActions, error) {
	var ra EnumActions
	switch a.Action {
	case "replace":
		ra = REPLACE
		break
	case "update":
		ra = UPDATE
		break
	case "delete":
		ra = DELETE
		break
	default:
		return ra, fmt.Errorf("action isn't set properly, while should be REPLACE / UPDATE / DELETE")
	}
	return ra, nil
}

func (a *Action) SetAction(ra EnumActions) error {
	switch ra {
	case DELETE:
		a.Action = "delete"
		break
	case REPLACE:
		a.Action = "replace"
		break
	case UPDATE:
		a.Action = "update"
		break
	default:
		return fmt.Errorf("action provided isn't correct, while should be REPLACE / UPDATE / DELETE")
	}
	return nil
}
