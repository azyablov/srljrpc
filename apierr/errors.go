package apierr

import (
	"errors"
	"fmt"
	"strings"
)

type ClientError struct {
	CltFunction string
	Message     string
	Err         error
}

func (e ClientError) Error() string {
	f := strings.TrimSpace(e.CltFunction)
	f = strings.ToLower(string(f[0])) + f[1:]
	return fmt.Sprintf("%s(): %s", f, e.Message)
}

func (e ClientError) Is(target error) bool {
	if target == nil {
		return false
	}

	if x, ok := target.(ClientError); ok {
		if x.Err == nil && e.Err == nil || x.Err != nil && e.Err != nil {
			return x.CltFunction == e.CltFunction && x.Message == e.Message && x.Err.Error() == e.Err.Error()
		}
		return x.CltFunction == e.CltFunction && x.Message == e.Message
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
	Message     string
	Err         error
}

func (e MessageError) Error() string {
	f := strings.TrimSpace(e.MsgFunction)
	f = strings.ToLower(string(f[0])) + f[1:]
	return fmt.Sprintf("%s(): %s", f, e.Message)
}

func (e MessageError) Is(target error) bool {
	if target == nil {
		return false
	}

	if x, ok := target.(MessageError); ok {
		if x.Err == nil && e.Err == nil || x.Err != nil && e.Err != nil {
			return x.MsgFunction == e.MsgFunction && x.Message == e.Message && x.Err.Error() == e.Err.Error()
		}
		return x.MsgFunction == e.MsgFunction && x.Message == e.Message
	}
	return false
}

func (e MessageError) As(target interface{}) bool {
	return errors.As(e.Err, target)
}

func (e MessageError) Unwrap() error {
	return e.Err
}
