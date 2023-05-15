package datastores

import "fmt"

// EnumDatastores is an enumeration type of the datastore types.
type EnumDatastores int

//	Valid enumeration EnumDatastores:
//		CANDIDATE - Used to change the configuration of the system with the get, set, and validate methods; default datastore is used if the datastore parameter is not provided.
//		RUNNING - Used to retrieve the active configuration with the get method.
//		STATE - Used to retrieve the running (active) configuration along with the operational state.
//		TOOLS - Used to perform operational tasks on the system; only supported with the update action command and the set method.
//
// By default we use CANDIDATE datastore, which is empty string OR "candidate" string in case specified explicitly.
const (
	CANDIDATE EnumDatastores = iota
	RUNNING
	STATE
	TOOLS
)

// Error messages for the Datastore class.
const (
	GetErrMsg = "datastore isn't set properly, while should be CANDIDATE / RUNNING / STATE / TOOLS"
	SetErrMsg = "datastore provided isn't correct, while should be CANDIDATE / RUNNING / STATE / TOOLS"
)

// Datastore class implementation inherited directly by Command (for GET only!)
// and Params (for SET/VALIDATE and GET in case Datastore applied for all commands underneath).
// Datastore is optional; selects the datastore to perform the method against. CANDIDATE datastore is used if the datastore parameter is not provided.
type Datastore struct {
	Datastore string `json:"datastore,omitempty"`
}

// GetDatastore returns the datastore type and non nil error if the datastore is not set properly.
func (d *Datastore) GetDatastore() (EnumDatastores, error) {
	var rd EnumDatastores
	switch d.Datastore {
	case "candidate":
		rd = CANDIDATE
	case "running":
		rd = RUNNING
	case "state":
		rd = STATE
	case "tools":
		rd = TOOLS
	case "":
		rd = CANDIDATE
	default:
		return rd, fmt.Errorf(GetErrMsg)
	}
	return rd, nil
}

// SetDatastore sets the datastore type and non nil error if provided datastore is not correct.
func (d *Datastore) SetDatastore(rd EnumDatastores) error {
	switch rd {
	case CANDIDATE:
		d.Datastore = "candidate"
	case RUNNING:
		d.Datastore = "running"
	case STATE:
		d.Datastore = "state"
	case TOOLS:
		d.Datastore = "tools"
	default:
		return fmt.Errorf(SetErrMsg)
	}
	return nil
}

// Additional implementation helper method to get datastore name as string.
func (d *Datastore) DatastoreName() string {
	if d.Datastore == "" {
		return "candidate"
	}
	return d.Datastore
}

// Additional implementation helper method to clean datastore name, used for Command in case of SET/VALIDATE to clean datastore name for all commands underneath.
func (d *Datastore) CleanDatastore() {
	d.Datastore = ""
}

// Additional implementation helper method to check if datastore is default, used for Command in case of GET to check if datastore is default for all commands underneath.
func (d *Datastore) IsDefaultDatastore() bool {
	return d.Datastore == ""
}
