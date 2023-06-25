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
	ErrClntUndefined EnumCltErr = iota
	ErrClntNoHost
	ErrClntTargetVerification
	ErrClntMarshalling
	ErrClntHTTPReqCreation
	ErrClntHTTPSend
	ErrClntHTTPStatus
	ErrClntJSONUnmarshalling
	ErrClntIDMismatch
	ErrClntJSONRPC
	ErrClntCmdCreation
	ErrClntRPCReqCreation
	ErrClntActNONE
	ErrClntActUnsupported
	ErrClntNoPort
	ErrClntNoUsername
	ErrClntNoPassword
	ErrClntTLSFilesUnspecified
	ErrClntTLSFOpenCA
	ErrClntTLSLoadCAPEM
	ErrClntTLSLoadCertPair
	ErrClntTLSCertParsing
	ErrClntCBFuncLowerThanCT
	ErrClntCBFuncIsNil
	ErrClntCBFuncExec
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
	ErrMsgCLISettingMethod         // Contains underlying error.
	ErrMsgCLIAddingCmdsInReq       // Contains underlying error.
	ErrMsgCLISettingOutFormat      // Contains underlying error.
	ErrMsgCLIMarshalling           // Contains underlying error.
	ErrMsgRespMarshalling          // Contains underlying error.
	ErrMsgReqSettingConfirmTimeout // Contains underlying error.
	ErrMsgReqSettingDSParams       // Contains underlying error.
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
	case ErrClntUndefined:
		m = "undefined error"
	case ErrClntNoHost:
		m = "host is not set, but mandatory"
	case ErrClntTargetVerification:
		m = "target verification error"
	case ErrClntMarshalling:
		m = "marshalling error"
	case ErrClntHTTPReqCreation:
		m = "HTTP request creation error"
	case ErrClntHTTPSend:
		m = "HTTP sending error"
	case ErrClntHTTPStatus:
		m = "HTTP status error"
	case ErrClntJSONUnmarshalling:
		m = "JSON unmarshalling error"
	case ErrClntIDMismatch:
		m = "request and response IDs do not match"
	case ErrClntJSONRPC:
		m = "JSON-RPC error"
	case ErrClntCmdCreation:
		m = "command creation error"
	case ErrClntRPCReqCreation:
		m = "RPC request creation error"
	case ErrClntActNONE:
		m = "action can't be NONE"
	case ErrClntActUnsupported:
		m = "unsupported action specified"
	case ErrClntNoPort:
		m = "port could not be nil"
	case ErrClntNoUsername:
		m = "username could not be nil"
	case ErrClntNoPassword:
		m = "password could not be nil"
	case ErrClntTLSFilesUnspecified:
		m = "one of more files for rootCA / certificate / key are not specified"
	case ErrClntTLSFOpenCA:
		m = "failed to open rootCA file"
	case ErrClntTLSLoadCAPEM:
		m = "can't load PEM file for rootCA"
	case ErrClntTLSLoadCertPair:
		m = "can't load PEM file for certificate / key pair"
	case ErrClntTLSCertParsing:
		m = "certificate parsing error"
	case ErrClntCBFuncLowerThanCT:
		m = "callback timeout must be lower than confirm timeout"
	case ErrClntCBFuncIsNil:
		m = "callback function is nil"
	case ErrClntCBFuncExec:
		m = "callback function execution error"
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
	case ErrMsgReqSettingConfirmTimeout:
		m = "error setting confirm timeout"
	case ErrMsgReqSettingDSParams:
		m = "error setting datastore parameters, check underlying error"
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
