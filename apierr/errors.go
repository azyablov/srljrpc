package apierr

import (
	"errors"
	"fmt"
	"strings"
)

type EnumCltErr int
type EnumMsgErr int

// Error messages for the Client class, which is the main class of the package.
// For the error definition see the ClientError.Error() method.
const (
	ErrCltUndefined EnumCltErr = iota
	ErrCtlNoHost
	ErrCtlTargetVerification
	ErrCltMarshalling
	ErrCltHTTPReqCreation
	ErrCltHTTPSend
	ErrCltHTTPStatus
	ErrCltJSONUnmarshalling
	ErrCltIDMismatch
	ErrCtlJSONRPC
	ErrCtlCmdCreation
	ErrCtlRPCReqCreation
	ErrCtlActNONE
	ErrCtlActUnsupported
	ErrCtlNoPort
	ErrCtlNoUsername
	ErrCtlNoPassword
	ErrCtlTLSFilesUnspecified
	ErrCtlTLSFOpenCA
	ErrCtlTLSLoadCAPEM
	ErrCtlTLSLoadCertPair
	ErrCtlTLSCertParsing
)

const (
	ErrMsgUndefined   EnumMsgErr = iota
	ErrMsgCmdCreation            // Contains underlying error.
	ErrMsgSetNotAllowedActForTools
	ErrMsgSettingMethod       // Contains underlying error.
	ErrMsgReqAddingCmds       // Contains underlying error.
	ErrMsgReqMarshalling      // Contains underlying error.
	ErrMsgReqSettingOutFormat // Contains underlying error.
	ErrMsgGettingMethod       // Contains underlying error.
	ErrMsgYANGSpecNotAllowed
	ErrMsgReqSettingYMParams // Contains underlying error.
	ErrMsgReqGetDSNotAllowed
	ErrMsgReqSetSettingAction // Contains underlying error.
	ErrMsgDSCandidateUpdateNoValue
	ErrMsgDSToolsSetUpdateOnly
	ErrMsgDSToolsCandidateSetOnly
	ErrMsgDSCandidateValidateOnly
	ErrMsgDSCandidateDiffOnly
	ErrMsgDSSpecNotAllowedForUnknownMethod
	ErrMsgCLISettingMethod    // Contains underlying error.
	ErrMsgCLIAddingCmdsInReq  // Contains underlying error.
	ErrMsgCLISettingOutFormat // Contains underlying error.
	ErrMsgCLIMarshalling      // Contains underlying error.
	ErrMsgRespMarshalling     // Contains underlying error.
)

type ClientError struct {
	CltFunction string
	Code        EnumCltErr
	Err         error
}

func (e ClientError) Error() string {
	var m string
	f := strings.TrimSpace(e.CltFunction)
	f = strings.ToLower(string(f[0])) + f[1:]

	switch e.Code {
	case ErrCltUndefined:
		m = "undefined error"
	case ErrCtlNoHost:
		m = "host is not set, but mandatory"
	case ErrCtlTargetVerification:
		m = "target verification error"
	case ErrCltMarshalling:
		m = "marshalling error"
	case ErrCltHTTPReqCreation:
		m = "HTTP request creation error"
	case ErrCltHTTPSend:
		m = "HTTP sending error"
	case ErrCltHTTPStatus:
		m = "HTTP status error"
	case ErrCltJSONUnmarshalling:
		m = "JSON unmarshalling error"
	case ErrCltIDMismatch:
		m = "request and response IDs do not match"
	case ErrCtlJSONRPC:
		m = "JSON-RPC error"
	case ErrCtlCmdCreation:
		m = "command creation error"
	case ErrCtlRPCReqCreation:
		m = "RPC request creation error"
	case ErrCtlActNONE:
		m = "action can't be NONE"
	case ErrCtlActUnsupported:
		m = "unsupported action specified"
	case ErrCtlNoPort:
		m = "port could not be nil"
	case ErrCtlNoUsername:
		m = "username could not be nil"
	case ErrCtlNoPassword:
		m = "password could not be nil"
	case ErrCtlTLSFilesUnspecified:
		m = "one of more files for rootCA / certificate / key are not specified"
	case ErrCtlTLSFOpenCA:
		m = "failed to open rootCA file"
	case ErrCtlTLSLoadCAPEM:
		m = "can't load PEM file for rootCA"
	case ErrCtlTLSLoadCertPair:
		m = "can't load PEM file for certificate / key pair"
	case ErrCtlTLSCertParsing:
		m = "certificate parsing error"
	default:
		m = "incorrect code error"
	}
	return fmt.Sprintf("%s: %s", f, m)
}

func (e ClientError) Is(target error) bool {
	if target == nil {
		return false
	}

	if x, ok := target.(ClientError); ok {
		if x.Err == nil && e.Err == nil || x.Err != nil && e.Err != nil {
			return x.CltFunction == e.CltFunction && x.Code == e.Code && x.Err.Error() == e.Err.Error()
		}
		return x.CltFunction == e.CltFunction && x.Code == e.Code
	}
	return false
}

func (e ClientError) As(target interface{}) bool {
	return errors.As(e.Err, target)
}

func (e ClientError) Unwrap() error {
	return e.Err
}

type MessageError struct {
	MsgFunction string
	Code        EnumMsgErr
	Err         error
}

func (e MessageError) Error() string {
	var m string
	f := strings.TrimSpace(e.MsgFunction)
	f = strings.ToLower(string(f[0])) + f[1:]
	switch e.Code {
	case ErrMsgCmdCreation:
		m = "command creation error"
	case ErrMsgSetNotAllowedActForTools:
		m = "no delete or replace actions allowed for method set and datastore TOOLS"
	case ErrMsgSettingMethod:
		m = "error setting method"
	case ErrMsgReqAddingCmds:
		m = "error adding commands in request"
	case ErrMsgReqMarshalling:
		m = "marshalling error"
	case ErrMsgReqSettingOutFormat:
		m = "output format can't be set on request"
	case ErrMsgGettingMethod:
		m = "error getting method"
	case ErrMsgYANGSpecNotAllowed:
		m = "yang models specification on Request.Params level is not supported for method"
	case ErrMsgReqSettingYMParams:
		m = "error setting yang models specification on Request.Params level"
	case ErrMsgReqGetDSNotAllowed:
		m = "datastore is not allowed for method get"
	case ErrMsgReqSetSettingAction:
		m = "error setting action"
	case ErrMsgDSCandidateUpdateNoValue:
		m = "value isn't specified or not found in the path for method set and datastore CANDIDATE"
	case ErrMsgDSToolsSetUpdateOnly:
		m = "only update action is allowed with TOOLS datastore for method set"
	case ErrMsgDSToolsCandidateSetOnly:
		m = "only CANDIDATE and TOOLS datastores allowed for method set"
	case ErrMsgDSCandidateValidateOnly:
		m = "only CANDIDATE datastore allowed for method validate"
	case ErrMsgDSCandidateDiffOnly:
		m = "only CANDIDATE datastore allowed for method diff"
	case ErrMsgDSSpecNotAllowedForUnknownMethod:
		m = "datastore specification on Request.Params level is not supported for unknown method"
	case ErrMsgCLISettingMethod:
		m = "error setting method"
	case ErrMsgCLIAddingCmdsInReq:
		m = "error adding commands in request"
	case ErrMsgCLISettingOutFormat:
		m = "output format can't be set on request"
	case ErrMsgCLIMarshalling:
		m = "marshalling error"
	case ErrMsgRespMarshalling:
		m = "marshalling error"
	default:
		m = "incorrect code error"
	}
	return fmt.Sprintf("%s(): %s", f, m)
}

func (e MessageError) Is(target error) bool {
	if target == nil {
		return false
	}

	if x, ok := target.(MessageError); ok {
		if x.Err == nil && e.Err == nil || x.Err != nil && e.Err != nil {
			return x.MsgFunction == e.MsgFunction && x.Code == e.Code && x.Err.Error() == e.Err.Error()
		}
		return x.MsgFunction == e.MsgFunction && x.Code == e.Code
	}
	return false
}

func (e MessageError) As(target interface{}) bool {
	return errors.As(e.Err, target)
}

func (e MessageError) Unwrap() error {
	return e.Err
}
