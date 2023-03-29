package methods

import "fmt"

type EnumMethods int

//	class EnumMethods {
//		<<enumeration>>
//		note "Used to retrieve configuration and state details from the system. The get method can be used with candidate, running, and state datastores, but cannot be used with the tools datastore."
//		GET
//		note "Used to set a configuration or run operational transaction. The set method can be used with the candidate and tools datastores."
//		SET
//		note "Used to run CLI commands. The get and set methods are restricted to accessing data structures via the YANG models, but the cli method can access any commands added to the system via python plug-ins or aliases."
//		CLI
//		note "Used to verify that the system accepts a configuration transaction before applying it to the system."
//		VALIDATE
//	}
//
// EnumMethods "1" --o Method: is one of
const (
	INVALID_METHOD EnumMethods = iota
	GET                        // note "Used to retrieve configuration and state details from the system. The get method can be used with candidate, running, and state datastores, but cannot be used with the tools datastore."
	SET                        // note "Used to set a configuration or run operational transaction. The set method can be used with the candidate and tools datastores."
	CLI                        // note "Used to run CLI commands. The get and set methods are restricted to accessing data structures via the YANG models, but the cli method can access any commands added to the system via python plug-ins or aliases."
	VALIDATE                   // note "Used to verify that the system accepts a configuration transaction before applying it to the system."
)

// note for method "Mandatory. Supported options are get, set, and validate. "
//
//	class method {
//		<<element>>
//		~GetMethod() EnumMethods
//		~SetMethod(EnumMethods) bool
//		+string Method
//	}
type Method struct {
	Method string `json:"method"`
}

func (m *Method) GetMethod() (EnumMethods, error) {
	var rm EnumMethods
	switch m.Method {
	case "get":
		rm = GET
		break
	case "set":
		rm = SET
		break
	case "cli":
		rm = CLI
		break
	case "validate":
		rm = VALIDATE
		break
	default:
		return rm, fmt.Errorf("method isn't set properly, while should be GET / SET / CLI / VALIDATE")
	}
	return rm, nil
}

func (m *Method) SetMethod(rm EnumMethods) error {
	switch rm {
	case GET:
		m.Method = "get"
		break
	case SET:
		m.Method = "set"
		break
	case CLI:
		m.Method = "cli"
		break
	case VALIDATE:
		m.Method = "validate"
		break
	default:
		return fmt.Errorf("method provided isn't correct, while should be GET / SET / CLI / VALIDATE")
	}
	return nil
}

func (m *Method) MethodName() string {
	if m.Method == "" {
		return "INVALID_METHOD"
	}
	return m.Method
}
