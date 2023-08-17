package formats

import "fmt"

// EnumOutputFormats is an enumeration type of output formats.
type EnumOutputFormats string

// Valid enumeration EnumOutputFormats:
// JSON - JSON format
// TEXT - Text format
// TABLE - Table format
// By default we use JSON output format, which is empty string OR "json" string in case specified explicitly.
const (
	JSON  EnumOutputFormats = "json"
	TEXT  EnumOutputFormats = "text"
	TABLE EnumOutputFormats = "table"
)

// Error messages for the OutputFormat class.
const (
	GetErrMsg = "output format isn't set properly, while should be JSON / XML / TABLE"
	SetErrMsg = "output format provided isn't correct, while should be JSON / XML / TABLE"
)

// OutputFormat is optional. Defines the output format. Output defaults to JSON if not specified.
// OutputFormat class implementation inherited directly by Params.
type OutputFormat struct {
	OutputFormat string `json:"output-format,omitempty"`
}

// GetFormat returns the output format type and non nil error if the output format is not set properly.
func (of *OutputFormat) GetFormat() (EnumOutputFormats, error) {
	var rf EnumOutputFormats
	switch of.OutputFormat {
	case "json", "text", "table":
		rf = EnumOutputFormats(of.OutputFormat)
	// case "json":
	// 	rf = JSON
	// case "text":
	// 	rf = TEXT
	// case "table":
	// 	rf = TABLE
	case "":
		rf = JSON
	default:
		return rf, fmt.Errorf(GetErrMsg)
	}
	return rf, nil
}

// SetFormat sets the output format type and non nil error if provided output format is not correct.
func (of *OutputFormat) SetFormat(ofs EnumOutputFormats) error {
	switch ofs {
	case JSON, TEXT, TABLE:
		of.OutputFormat = string(ofs)
	// case JSON:
	// 	of.OutputFormat = "json"
	// case TEXT:
	// 	of.OutputFormat = "text"
	// case TABLE:
	// 	of.OutputFormat = "table"
	default:
		return fmt.Errorf(SetErrMsg)
	}
	return nil
}
