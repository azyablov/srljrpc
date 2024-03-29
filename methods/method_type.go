package methods

import "fmt"

// EnumMethods is an enumeration type of the methods.
type EnumMethods string

//	Valid enumeration EnumMethods:
//		GET - Used to retrieve configuration and state details from the system. The get method can be used with candidate, running, and state datastores, but cannot be used with the tools datastore.
//		SET - Used to set a configuration or run operational transaction. The set method can be used with the candidate and tools datastores.
//		CLI - Used to run CLI commands. The get and set methods are restricted to accessing data structures via the YANG models, but the cli method can access any commands added to the system via python plug-ins or aliases.
//		VALIDATE - Used to verify that the system accepts a configuration transaction before applying it to the system.
//		DIFF - Used to retrieve the difference between the set and candidate / tools configurations.
//
// At the time of object creation method is not set, so we use INVALID_METHOD as default value in order to force user to set it properly.
const (
	INVALID_METHOD EnumMethods = ""         // Default value, used to force user to set method properly.
	GET            EnumMethods = "get"      // Used to retrieve configuration and state details from the system. The get method can be used with candidate, running, and state datastores, but cannot be used with the tools datastore.
	SET            EnumMethods = "set"      // Used to set a configuration or run operational transaction. The set method can be used with the candidate and tools datastores.
	CLI            EnumMethods = "cli"      // Used to run CLI commands. The get and set methods are restricted to accessing data structures via the YANG models, but the cli method can access any commands added to the system via python plug-ins or aliases.
	VALIDATE       EnumMethods = "validate" // Used to verify that the system accepts a configuration transaction before applying it to the system.
	DIFF           EnumMethods = "diff"     // Used to retrieve the difference between the set and candidate / tools configurations.
)

// Error messages for the Method class.
const (
	GetErrMsg = "method isn't set properly, while should be GET / SET / CLI / VALIDATE"
	SetErrMsg = "method provided isn't correct, while should be GET / SET / CLI / VALIDATE"
)

// Method is Mandatory. Supported options are get, set, and validate.
type Method struct {
	Method string `json:"method"`
}

// GetMethod returns the method type and non nil error if the method is not set properly.
func (m *Method) GetMethod() (EnumMethods, error) {
	var rm EnumMethods
	switch m.Method {
	case "get", "set", "cli", "validate", "diff":
		rm = EnumMethods(m.Method)
	// case "get":
	// 	rm = GET
	// case "set":
	// 	rm = SET
	// case "cli":
	// 	rm = CLI
	// case "validate":
	// 	rm = VALIDATE
	// case "diff":
	// 	rm = DIFF
	default:
		return rm, fmt.Errorf(GetErrMsg)
	}
	return rm, nil
}

// SetMethod sets the method type and non nil error if provided method is not correct.
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
	case DIFF:
		m.Method = "diff"
	default:
		return fmt.Errorf(SetErrMsg)
	}
	return nil
}

// MethodName implementation helper returns method name in case it is set properly, otherwise returns INVALID_METHOD.
// func (m *Method) MethodName() string {
// 	if m.Method == "" {
// 		return "INVALID_METHOD"
// 	}
// 	return m.Method
// }
