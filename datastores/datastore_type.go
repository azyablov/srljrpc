package datastores

import "fmt"

type EnumDatastores int

//	class EnumDatastores {
//		<<enumeration>>
//		note "Used to change the configuration of the system with the get, set, and validate methods; default datastore is used if the datastore parameter is not provided."
//		CANDIDATE
//		note "Used to retrieve the active configuration with the get method."
//		RUNNING
//		note "Used to retrieve the running (active) configuration along with the operational state."
//		STATE
//		note "Used to perform operational tasks on the system; only supported with the update action command and the set method."
//		TOOLS
//	}
//
// EnumDatastores "1" --o Datastore: OneOf
const (
	CANDIDATE EnumDatastores = iota
	RUNNING
	STATE
	TOOLS
)

// note for datastore "Optional; selects the datastore to perform the method against. CANDIDATE datastore is used if the datastore parameter is not provided."
//
//	class Datastore {
//		<<element>>
//		+GetDatastore(): EnumDatastores
//		+SetDatastore(d: EnumDatastores): error
//		+string Datastore
//	}
type Datastore struct {
	Datastore string `json:"datastore,omitempty"`
}

func (d *Datastore) GetDatastore() (EnumDatastores, error) {
	var rd EnumDatastores
	switch d.Datastore {
	case "candidate":
		rd = CANDIDATE
		break
	case "running":
		rd = RUNNING
		break
	case "state":
		rd = STATE
		break
	case "tools":
		rd = TOOLS
		break
	default:
		return rd, fmt.Errorf("datastore isn't set properly, while should be CANDIDATE / RUNNING / STATE / TOOLS")
	}
	return rd, nil
}

func (d *Datastore) SetDatastore(rd EnumDatastores) error {
	switch rd {
	case CANDIDATE:
		d.Datastore = "candidate"
		break
	case RUNNING:
		d.Datastore = "running"
		break
	case STATE:
		d.Datastore = "state"
		break
	case TOOLS:
		d.Datastore = "tools"
		break
	default:
		return fmt.Errorf("datastore provided isn't correct, while should be CANDIDATE / RUNNING / STATE / TOOLS")
	}
	return nil
}
