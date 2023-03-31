package formats

import "fmt"

//	class EnumOutputFormats {
//		<<enumeration>>
//		JSON
//		TEXT
//		TABLE
//	}
//
// EnumOutputFormats "1" --o OutputFormat: OneOf
type EnumOutputFormats int

const (
	JSON EnumOutputFormats = iota
	XML
	TABLE
)

const (
	GetErrMsg = "output format isn't set properly, while should be JSON / XML / TABLE"
	SetErrMsg = "output format provided isn't correct, while should be JSON / XML / TABLE"
)

// note for outputFormat "Optional. Defines the output format. Output defaults to JSON if not specified."
//
//	class OutputFormat {
//		<<element>>
//		+GetFormat() EnumOutputFormats
//		+SetFormat(EnumOutputFormats of) error
//		#string OutputFormat
//	}
type OutputFormat struct {
	OutputFormat string `json:"output-format,omitempty"`
}

func (of *OutputFormat) GetFormat() (EnumOutputFormats, error) {
	var rf EnumOutputFormats
	switch of.OutputFormat {
	case "json":
		rf = JSON
	case "xml":
		rf = XML
	case "table":
		rf = TABLE
	case "":
		rf = JSON
	default:
		return rf, fmt.Errorf(GetErrMsg)
	}
	return rf, nil
}

func (of *OutputFormat) SetFormat(ofs EnumOutputFormats) error {
	switch ofs {
	case JSON:
		of.OutputFormat = "json"
	case XML:
		of.OutputFormat = "xml"
	case TABLE:
		of.OutputFormat = "table"
	default:
		return fmt.Errorf(SetErrMsg)
	}
	return nil
}
