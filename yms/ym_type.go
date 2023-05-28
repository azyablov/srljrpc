package yms

import "fmt"

type EnumYmType int

// YmType is enumeration type for the yang models in request.
const (
	SRL EnumYmType = iota
	OC
)

// YANG models to use for the get/set/validate/diff actions. Default is SRL
type YmType struct {
	YangModels string `json:"yang-models,omitempty"`
}

// Error messages for the Method class.
const (
	GetErrMsg = "yang models isn't set properly, while should be SRL or OC"
	SetErrMsg = "yang models provided isn't correct, while should be SRL or OC"
)

// SetYmType sets the yang models and non nil error if provided yang models format is not correct.
func (y *YmType) SetYmType(ym EnumYmType) error {
	switch ym {
	case SRL:
		y.YangModels = "srl"
	case OC:
		y.YangModels = "oc"
	default:
		return fmt.Errorf(SetErrMsg)
	}
	return nil
}

// GetYmType returns the yang models and non nil error if provided yang models was not set properly.
func (y *YmType) GetYmType() (EnumYmType, error) {
	var ym EnumYmType
	switch y.YangModels {
	case "srl":
		ym = SRL
	case "oc":
		ym = OC
	case "":
		ym = SRL
	default:
		return ym, fmt.Errorf(GetErrMsg)
	}
	return ym, nil
}
