//go:generate stringer -type=EnumCltErr,EnumMsgErr -linecomment -output=errors_string.go
package apierr

import (
	"errors"
)

type EnumCltErr int
type EnumMsgErr int

// NewClientError returns ClientError with the specified code and wrapped error.
func NewClientError(ce EnumCltErr, e error) error {
	return ClientError{
		Code: ce,
		Err:  e,
	}
}

// NewMessageError returns MessageError with the specified code and wrapped error.
func NewMessageError(me EnumMsgErr, e error) error {
	return MessageError{
		Code: me,
		Err:  e,
	}
}

// Error codes for the Client class, which is the main class of the package.
const (
	CodeClntUndefined             EnumCltErr = iota // undefined error
	CodeClntNoHost                                  // host is not set, but mandatory
	CodeClntTargetVerification                      // target verification error
	CodeClntReqMarshalling                          // request marshalling error
	CodeClntHTTPReqCreation                         // HTTP request creation error
	CodeClntHTTPSend                                // HTTP send error
	CodeClntHTTPStatus                              // HTTP status error
	CodeClntRespJSONUnmarshalling                   // response JSON unmarshalling error
	CodeClntIDMismatch                              // request and response IDs do not match
	CodeClntJSONRPCResp                             // JSON-RPC response error
	CodeClntCmdCreation                             // command creation error
	CodeClntRPCReqCreation                          // RPC request creation error
	CodeClntActNONE                                 // action can't be NONE
	CodeClntActUnsupported                          // unsupported action specified
	CodeClntNoPort                                  // port could not be nil
	CodeClntNoUsername                              // username could not be nil
	CodeClntNoPassword                              // password could not be nil
	CodeClntTLSFilesUnspecified                     // one of more files for rootCA / certificate / key are not specified
	CodeClntTLSFOpenCA                              // failed to open rootCA file
	CodeClntTLSLoadCAPEM                            // can't load PEM file for rootCA
	CodeClntTLSLoadCertPair                         // can't load PEM file for certificate / key pair
	CodeClntTLSCertParsing                          // certificate parsing error
	CodeClntCBFuncLowerThanCT                       // callback timeout must be lower than confirm timeout
	CodeClntCBFuncIsNil                             // callback function is nil
	CodeClntCBFuncExec                              // callback function execution error
	CodeClntDatastoreUnsupported                    // datastore is not supported for this method
)

var (
	ErrClntUndefined            = NewClientError(CodeClntUndefined, nil)
	ErrClntTargetVerification   = NewClientError(CodeClntTargetVerification, nil)
	ErrClntMarshalling          = NewClientError(CodeClntReqMarshalling, nil)
	ErrClntHTTPReqCreation      = NewClientError(CodeClntHTTPReqCreation, nil)
	ErrClntHTTPSend             = NewClientError(CodeClntHTTPSend, nil)
	ErrClntHTTPStatus           = NewClientError(CodeClntHTTPStatus, nil)
	ErrClntJSONUnmarshalling    = NewClientError(CodeClntRespJSONUnmarshalling, nil)
	ErrClntIDMismatch           = NewClientError(CodeClntIDMismatch, nil)
	ErrClntJSONRPCResp          = NewClientError(CodeClntJSONRPCResp, nil)
	ErrClntCmdCreation          = NewClientError(CodeClntCmdCreation, nil)
	ErrClntRPCReqCreation       = NewClientError(CodeClntRPCReqCreation, nil)
	ErrClntActNONE              = NewClientError(CodeClntActNONE, nil)
	ErrClntActUnsupported       = NewClientError(CodeClntActUnsupported, nil)
	ErrClntNoPort               = NewClientError(CodeClntNoPort, nil)
	ErrClntNoUsername           = NewClientError(CodeClntNoUsername, nil)
	ErrClntNoPassword           = NewClientError(CodeClntNoPassword, nil)
	ErrClntTLSFilesUnspecified  = NewClientError(CodeClntTLSFilesUnspecified, nil)
	ErrClntTLSFOpenCA           = NewClientError(CodeClntTLSFOpenCA, nil)
	ErrClntTLSLoadCAPEM         = NewClientError(CodeClntTLSLoadCAPEM, nil)
	ErrClntTLSLoadCertPair      = NewClientError(CodeClntTLSLoadCertPair, nil)
	ErrClntTLSCertParsing       = NewClientError(CodeClntTLSCertParsing, nil)
	ErrClntCBFuncLowerThanCT    = NewClientError(CodeClntCBFuncLowerThanCT, nil)
	ErrClntCBFuncIsNil          = NewClientError(CodeClntCBFuncIsNil, nil)
	ErrClntCBFuncExec           = NewClientError(CodeClntCBFuncExec, nil)
	ErrClntDatastoreUnsupported = NewClientError(CodeClntDatastoreUnsupported, nil)
)

// Error codes for the Message class, which is the main class of the package.
const (
	CodeMsgUndefined                        EnumMsgErr = iota // undefined error
	CodeMsgCmdCreation                                        // command creation error
	CodeMsgSetNotAllowedActForTools                           // no delete or replace actions allowed for method set and datastore TOOLS
	CodeMsgSettingMethod                                      // error setting method in request
	CodeMsgReqAddingCmds                                      // error adding commands in request
	CodeMsgReqMarshalling                                     // marshalling error
	CodeMsgReqSettingOutFormat                                // error setting output format in request
	CodeMsgGettingMethod                                      // error getting method
	CodeMsgYANGSpecNotAllowed                                 // yang models specification on Request.Params level is not supported for method
	CodeMsgReqSettingYMParams                                 // error setting yang models specification on Request.Params level
	CodeMsgReqGetDSNotAllowed                                 // datastore is not allowed for method get
	CodeMsgReqSetSettingAction                                // setting action error for method set
	CodeMsgDSCandidateUpdateNoValue                           // value isn't specified or not found in the path for method set and datastore CANDIDATE
	CodeMsgDSToolsSetUpdateOnly                               // only update action is allowed with TOOLS datastore for method set
	CodeMsgDSToolsCandidateSetOnly                            // only CANDIDATE and TOOLS datastores allowed for method set
	CodeMsgDSCandidateValidateOnly                            // only CANDIDATE datastore allowed for method validate
	CodeMsgDSCandidateDiffOnly                                // only CANDIDATE datastore allowed for method diff
	CodeMsgDSSpecNotAllowedForUnknownMethod                   // datastore specification on Request.Params level is not supported for unknown method
	CodeMsgCLISettingMethod                                   // error setting cli method
	CodeMsgCLIAddingCmdsInReq                                 // error adding cli commands in request
	CodeMsgCLISettingOutFormat                                // error setting output format for cli method
	CodeMsgCLIMarshalling                                     // cli request marshalling error
	CodeMsgRespMarshalling                                    // JSON response marshalling error
	CodeMsgReqSettingConfirmTimeout                           // confirm timeout is allowed for SET method only
	CodeMsgReqSettingDSParams                                 // error setting datastore parameters in request (check underlying error)
)

var (
	ErrMsgUndefined                        = NewMessageError(CodeMsgUndefined, nil)
	ErrMsgCmdCreation                      = NewMessageError(CodeMsgCmdCreation, nil)
	ErrMsgSetNotAllowedActForTools         = NewMessageError(CodeMsgSetNotAllowedActForTools, nil)
	ErrMsgSettingMethod                    = NewMessageError(CodeMsgSettingMethod, nil)
	ErrMsgReqAddingCmds                    = NewMessageError(CodeMsgReqAddingCmds, nil)
	ErrMsgReqMarshalling                   = NewMessageError(CodeMsgReqMarshalling, nil)
	ErrMsgReqSettingOutFormat              = NewMessageError(CodeMsgReqSettingOutFormat, nil)
	ErrMsgGettingMethod                    = NewMessageError(CodeMsgGettingMethod, nil)
	ErrMsgYANGSpecNotAllowed               = NewMessageError(CodeMsgYANGSpecNotAllowed, nil)
	ErrMsgReqSettingYMParams               = NewMessageError(CodeMsgReqSettingYMParams, nil)
	ErrMsgReqGetDSNotAllowed               = NewMessageError(CodeMsgReqGetDSNotAllowed, nil)
	ErrMsgReqSetSettingAction              = NewMessageError(CodeMsgReqSetSettingAction, nil)
	ErrMsgDSCandidateUpdateNoValue         = NewMessageError(CodeMsgDSCandidateUpdateNoValue, nil)
	ErrMsgDSToolsSetUpdateOnly             = NewMessageError(CodeMsgDSToolsSetUpdateOnly, nil)
	ErrMsgDSToolsCandidateSetOnly          = NewMessageError(CodeMsgDSToolsCandidateSetOnly, nil)
	ErrMsgDSCandidateValidateOnly          = NewMessageError(CodeMsgDSCandidateValidateOnly, nil)
	ErrMsgDSCandidateDiffOnly              = NewMessageError(CodeMsgDSCandidateDiffOnly, nil)
	ErrMsgDSSpecNotAllowedForUnknownMethod = NewMessageError(CodeMsgDSSpecNotAllowedForUnknownMethod, nil)
	ErrMsgCLISettingMethod                 = NewMessageError(CodeMsgCLISettingMethod, nil)
	ErrMsgCLIAddingCmdsInReq               = NewMessageError(CodeMsgCLIAddingCmdsInReq, nil)
	ErrMsgCLISettingOutFormat              = NewMessageError(CodeMsgCLISettingOutFormat, nil)
	ErrMsgCLIMarshalling                   = NewMessageError(CodeMsgCLIMarshalling, nil)
	ErrMsgRespMarshalling                  = NewMessageError(CodeMsgRespMarshalling, nil)
	ErrMsgReqSettingConfirmTimeout         = NewMessageError(CodeMsgReqSettingConfirmTimeout, nil)
	ErrMsgReqSettingDSParams               = NewMessageError(CodeMsgReqSettingDSParams, nil)
)

type ClientError struct {
	CltFunction string
	Code        EnumCltErr
	Err         error
}

func (e ClientError) Error() string {
	var m string
	// f := strings.TrimSpace(e.CltFunction)
	// f = strings.ToLower(string(f[0])) + f[1:]

	switch e.Code {
	case CodeClntUndefined, CodeClntNoHost, CodeClntTargetVerification, CodeClntReqMarshalling,
		CodeClntHTTPReqCreation, CodeClntHTTPSend, CodeClntHTTPStatus, CodeClntRespJSONUnmarshalling,
		CodeClntIDMismatch, CodeClntJSONRPCResp, CodeClntCmdCreation, CodeClntRPCReqCreation, CodeClntActNONE,
		CodeClntActUnsupported, CodeClntNoPort, CodeClntNoUsername, CodeClntNoPassword, CodeClntTLSFilesUnspecified,
		CodeClntTLSFOpenCA, CodeClntTLSLoadCAPEM, CodeClntTLSLoadCertPair, CodeClntTLSCertParsing, CodeClntCBFuncLowerThanCT,
		CodeClntCBFuncIsNil, CodeClntCBFuncExec, CodeClntDatastoreUnsupported:
		m = e.Code.String()
	// case CodeClntUndefined:
	// 	m = "undefined error"
	// case CodeClntNoHost:
	// 	m = "host is not set, but mandatory"
	// case CodeClntTargetVerification:
	// 	m = "target verification error"
	// case CodeClntMarshalling:
	// 	m = "marshalling error"
	// case CodeClntHTTPReqCreation:
	// 	m = "HTTP request creation error"
	// case CodeClntHTTPSend:
	// 	m = "HTTP sending error"
	// case CodeClntHTTPStatus:
	// 	m = "HTTP status error"
	// case CodeClntJSONUnmarshalling:
	// 	m = "JSON unmarshalling error"
	// case CodeClntIDMismatch:
	// 	m = "request and response IDs do not match"
	// case CodeClntJSONRPC:
	// 	m = "JSON-RPC error"
	// case CodeClntCmdCreation:
	// 	m = "command creation error"
	// case CodeClntRPCReqCreation:
	// 	m = "RPC request creation error"
	// case CodeClntActNONE:
	// 	m = "action can't be NONE"
	// case CodeClntActUnsupported:
	// 	m = "unsupported action specified"
	// case CodeClntNoPort:
	// 	m = "port could not be nil"
	// case CodeClntNoUsername:
	// 	m = "username could not be nil"
	// case CodeClntNoPassword:
	// 	m = "password could not be nil"
	// case CodeClntTLSFilesUnspecified:
	// 	m = "one of more files for rootCA / certificate / key are not specified"
	// case CodeClntTLSFOpenCA:
	// 	m = "failed to open rootCA file"
	// case CodeClntTLSLoadCAPEM:
	// 	m = "can't load PEM file for rootCA"
	// case CodeClntTLSLoadCertPair:
	// 	m = "can't load PEM file for certificate / key pair"
	// case CodeClntTLSCertParsing:
	// 	m = "certificate parsing error"
	// case CodeClntCBFuncLowerThanCT:
	// 	m = "callback timeout must be lower than confirm timeout"
	// case CodeClntCBFuncIsNil:
	// 	m = "callback function is nil"
	// case CodeClntCBFuncExec:
	// 	m = "callback function execution error"
	// case CodeClntDatastoreUnsupported:
	// 	m = "datastore is not supported for this method"
	default:
		m = "incorrect code error"
	}
	return m
}

func (e ClientError) Is(target error) bool {
	if target == nil {
		return false
	}

	if x, ok := target.(ClientError); ok {
		return x.Code == e.Code
	} else {
		return errors.Is(e.Err, target)
	}
}

func (e ClientError) As(target interface{}) bool {
	return errors.As(e, target)
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
	// f := strings.TrimSpace(e.MsgFunction)
	// f = strings.ToLower(string(f[0])) + f[1:]
	switch e.Code {
	case CodeMsgUndefined, CodeMsgCmdCreation, CodeMsgSetNotAllowedActForTools, CodeMsgSettingMethod,
		CodeMsgReqAddingCmds, CodeMsgReqMarshalling, CodeMsgReqSettingOutFormat, CodeMsgGettingMethod,
		CodeMsgYANGSpecNotAllowed, CodeMsgReqSettingYMParams, CodeMsgReqGetDSNotAllowed, CodeMsgReqSetSettingAction,
		CodeMsgDSCandidateUpdateNoValue, CodeMsgDSToolsSetUpdateOnly, CodeMsgDSToolsCandidateSetOnly,
		CodeMsgDSCandidateValidateOnly, CodeMsgDSCandidateDiffOnly, CodeMsgDSSpecNotAllowedForUnknownMethod,
		CodeMsgCLISettingMethod, CodeMsgCLIAddingCmdsInReq, CodeMsgCLISettingOutFormat, CodeMsgCLIMarshalling,
		CodeMsgRespMarshalling, CodeMsgReqSettingConfirmTimeout, CodeMsgReqSettingDSParams:
		m = e.Code.String()
	// case CodeMsgCmdCreation:
	// 	m = "command creation error"
	// case CodeMsgSetNotAllowedActForTools:
	// 	m = "no delete or replace actions allowed for method set and datastore TOOLS"
	// case CodeMsgSettingMethod:
	// 	m = "error setting method"
	// case CodeMsgReqAddingCmds:
	// 	m = "error adding commands in request"
	// case CodeMsgReqMarshalling:
	// 	m = "marshalling error"
	// case CodeMsgReqSettingOutFormat:
	// 	m = "output format can't be set on request"
	// case CodeMsgGettingMethod:
	// 	m = "error getting method"
	// case CodeMsgYANGSpecNotAllowed:
	// 	m = "yang models specification on Request.Params level is not supported for method"
	// case CodeMsgReqSettingYMParams:
	// 	m = "error setting yang models specification on Request.Params level"
	// case CodeMsgReqGetDSNotAllowed:
	// 	m = "datastore is not allowed for method get"
	// case CodeMsgReqSetSettingAction:
	// 	m = "error setting action"
	// case CodeMsgDSCandidateUpdateNoValue:
	// 	m = "value isn't specified or not found in the path for method set and datastore CANDIDATE"
	// case CodeMsgDSToolsSetUpdateOnly:
	// 	m = "only update action is allowed with TOOLS datastore for method set"
	// case CodeMsgDSToolsCandidateSetOnly:
	// 	m = "only CANDIDATE and TOOLS datastores allowed for method set"
	// case CodeMsgDSCandidateValidateOnly:
	// 	m = "only CANDIDATE datastore allowed for method validate"
	// case CodeMsgDSCandidateDiffOnly:
	// 	m = "only CANDIDATE datastore allowed for method diff"
	// case CodeMsgDSSpecNotAllowedForUnknownMethod:
	// 	m = "datastore specification on Request.Params level is not supported for unknown method"
	// case CodeMsgCLISettingMethod:
	// 	m = "error setting method"
	// case CodeMsgCLIAddingCmdsInReq:
	// 	m = "error adding commands in request"
	// case CodeMsgCLISettingOutFormat:
	// 	m = "output format can't be set on request"
	// case CodeMsgCLIMarshalling:
	// 	m = "marshalling error"
	// case CodeMsgRespMarshalling:
	// 	m = "marshalling error"
	// case CodeMsgReqSettingConfirmTimeout:
	// 	m = "error setting confirm timeout"
	// case CodeMsgReqSettingDSParams:
	// 	m = "error setting datastore parameters, check underlying error"
	default:
		m = "incorrect code error"
	}
	return m
}

func (e MessageError) Is(target error) bool {
	if target == nil {
		return false
	}

	if x, ok := target.(MessageError); ok {
		return x.Code == e.Code
	} else {
		return errors.Is(e.Err, target)
	}
}

func (e MessageError) As(target interface{}) bool {
	return errors.As(e, target)
}

func (e MessageError) Unwrap() error {
	return e.Err
}
