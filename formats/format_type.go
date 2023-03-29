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
		break
	case "xml":
		rf = XML
		break
	case "table":
		rf = TABLE
		break
	case "":
		rf = JSON
		break
	default:
		return rf, fmt.Errorf("output format isn't set properly, while should be JSON / XML / TABLE")
	}
	return rf, nil
}

func (of *OutputFormat) SetFormat(ofs EnumOutputFormats) error {
	switch ofs {
	case JSON:
		of.OutputFormat = "json"
		break
	case XML:
		of.OutputFormat = "xml"
		break
	case TABLE:
		of.OutputFormat = "table"
		break
	default:
		return fmt.Errorf("output format provided isn't correct, while should be JSON / XML / TABLE")
	}
	return nil
}
