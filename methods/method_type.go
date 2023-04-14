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
// At the time of object creation method is not set, so we use INVALID_METHOD as default value in order to force user to set it properly.
const (
	INVALID_METHOD EnumMethods = iota
	GET                        // note "Used to retrieve configuration and state details from the system. The get method can be used with candidate, running, and state datastores, but cannot be used with the tools datastore."
	SET                        // note "Used to set a configuration or run operational transaction. The set method can be used with the candidate and tools datastores."
	CLI                        // note "Used to run CLI commands. The get and set methods are restricted to accessing data structures via the YANG models, but the cli method can access any commands added to the system via python plug-ins or aliases."
	VALIDATE                   // note "Used to verify that the system accepts a configuration transaction before applying it to the system."
)

const (
	GetErrMsg = "method isn't set properly, while should be GET / SET / CLI / VALIDATE"
	SetErrMsg = "method provided isn't correct, while should be GET / SET / CLI / VALIDATE"
)

// note for method "Mandatory. Supported options are get, set, and validate. "
//
//	class method {
//		<<element>>
//		~GetMethod() EnumMethods
//		~SetMethod(EnumMethods) bool
//		+string Method
//	}
//
// Method class implementation.
type Method struct {
	Method string `json:"method"`
}

func (m *Method) GetMethod() (EnumMethods, error) {
	var rm EnumMethods
	switch m.Method {
	case "get":
		rm = GET
	case "set":
		rm = SET
	case "cli":
		rm = CLI
	case "validate":
		rm = VALIDATE
	default:
		return rm, fmt.Errorf(GetErrMsg)
	}
	return rm, nil
}

func (m *Method) SetMethod(rm EnumMethods) error {
	switch rm {
	case GET:
		m.Method = "get"
	case SET:
		m.Method = "set"
	case CLI:
		m.Method = "cli"
	case VALIDATE:
		m.Method = "validate"
	default:
		return fmt.Errorf(SetErrMsg)
	}
	return nil
}

// MethodName implementation helper returns method name in case it is set properly, otherwise returns INVALID_METHOD.
func (m *Method) MethodName() string {
	if m.Method == "" {
		return "INVALID_METHOD"
	}
	return m.Method
}
