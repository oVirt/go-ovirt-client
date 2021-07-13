package ovirtclient

import (
	"errors"
	"fmt"
	"strings"

	ovirtsdk "github.com/ovirt/go-ovirt"
)

// ErrorCode is a code that can be used to identify error types. These errors are identified on a best effort basis
// from the underlying oVirt connection.
type ErrorCode string

// EAccessDenied signals that the provided credentials for the oVirt engine were incorrect.
const EAccessDenied ErrorCode = "access_denied"

// ENotAnOVirtEngine signals that the server did not respond with a proper oVirt response.
// the cre
const ENotAnOVirtEngine ErrorCode = "not_ovirt_engine"

// ETLSError signals that the provided CA certificate did not match the server that was attempted to connect.
const ETLSError ErrorCode = "tls_error"

// ENotFound signals that the resource requested was not found.
const ENotFound ErrorCode = "not_found"

// EBug signals an error that should never happen. Please report this.
const EBug ErrorCode = "bug"

// EConnection signals a problem with the connection.
const EConnection ErrorCode = "connection"

// EPending signals that the client library is still waiting for an action to be completed.
const EPending ErrorCode = "pending"

// ETimeout signals that the client library has timed out waiting for an action to be completed.
const ETimeout ErrorCode = "timeout"

// EFieldMissing indicates that the oVirt API did not return a specific field. This is most likely a bug, please report
// it.
const EFieldMissing ErrorCode = "field_missing"

// EUnidentified is an unidentified oVirt error. Please report this error so we can add an error code for it.
const EUnidentified ErrorCode = "generic_error"

// EUnsupported signals that an action is not supported. This can indicate a disk format or a combination of parameters.
const EUnsupported ErrorCode = "unsupported"

// IsPermanent returns true if the given error code is permanent and a retry should not be attempted.
func (e ErrorCode) IsPermanent() bool {
	switch e {
	case EAccessDenied:
		return true
	case ENotAnOVirtEngine:
		return true
	case ETLSError:
		return true
	case ENotFound:
		return true
	case EBug:
		return true
	case EConnection:
		return false
	case EPending:
		return false
	case ETimeout:
		return false
	case EUnsupported:
		return true
	case EFieldMissing:
		return true
	default:
		return false
	}
}

// EngineError is an error representation for errors received while interacting with the oVirt engine.
//
// Usage:
//
// if err != nil {
//     var realErr ovirtclient.EngineError
//     if errors.As(err, &realErr) {
//          // deal with EngineError
//     } else {
//          // deal with other errors
//     }
// }
type EngineError interface {
	error

	// Is returns true if the current error, or any preceding error has the specified error code.
	Is(ErrorCode) bool
	// Code returns an error code for the failure.
	Code() ErrorCode
	// Unwrap returns the underlying error
	Unwrap() error
	// IsPermanent returns if a retry should be attempted or not.
	IsPermanent() bool
}

type engineError struct {
	message string
	code    ErrorCode
	cause   error
}

func (e *engineError) Is(code ErrorCode) bool {
	if e.code == code {
		return true
	}
	if cause := e.Unwrap(); cause != nil {
		var causeE EngineError
		if errors.As(cause, &causeE) {
			return causeE.Is(code)
		}
	}
	return false
}

func (e *engineError) Error() string {
	return e.message
}

func (e *engineError) Code() ErrorCode {
	return e.code
}

func (e *engineError) Unwrap() error {
	return e.cause
}

func (e *engineError) IsPermanent() bool {
	return e.code.IsPermanent()
}

func newError(code ErrorCode, message string, args ...interface{}) EngineError {
	return &engineError{
		message: fmt.Sprintf(message, args...),
		code:    code,
	}
}

// wrap wraps an error, adding an error code and message in the process. The wrapped error is added
// to the message automatically in Go style. If the passed error code is EUnidentified or not an EngineError
// this function will attempt to identify the error deeper.
func wrap(err error, code ErrorCode, message string, args ...interface{}) EngineError {
	realArgs := make([]interface{}, len(args)+1)
	realArgs[len(args)] = err
	copy(realArgs[0:len(args)], args)
	if code == EUnidentified {
		var realErr EngineError
		if errors.As(err, &realErr) {
			code = realErr.Code()
		} else {
			if e := realIdentify(err); e != nil {
				err = e
				code = e.Code()
			}
		}
	}
	realMessage := fmt.Sprintf(fmt.Sprintf("%s %s", message, "(%v)"), realArgs...)
	return &engineError{
		message: realMessage,
		code:    code,
		cause:   err,
	}
}

// identify attempts to identify the reason for the error and create a structure accordingly. If it fails to identify
// the reason it will return nil.
//
// Usage:
//
// if err != nil {
//     if wrappedError := identify(err); wrappedError != nil {
//         return wrappedError
//     }
//     // Handle unknown error here
// }
func identify(err error) error {
	return realIdentify(err)
}

func realIdentify(err error) EngineError {
	var authErr *ovirtsdk.AuthError
	var notFoundErr *ovirtsdk.NotFoundError
	switch {
	case errors.As(err, &authErr):
		fallthrough
	case strings.Contains(err.Error(), "access_denied"):
		return wrap(err, EAccessDenied, "access denied, check your credentials")
	case strings.Contains(err.Error(), "parse non-array sso with response"):
		return wrap(err,
			ENotAnOVirtEngine, "invalid credentials, or the URL does not point to an oVirt Engine, check your settings")
	case strings.Contains(err.Error(), "server gave HTTP response to HTTPS client"):
		return wrap(err,
			ENotAnOVirtEngine, "the server gave a HTTP response to a HTTPS client, check if your URL is correct")
	case strings.Contains(err.Error(), "tls"):
		fallthrough
	case strings.Contains(err.Error(), "x509"):
		return wrap(err, ETLSError, "TLS error, check your CA certificate settings")
	case errors.As(err, &notFoundErr):
		return wrap(err, ENotFound, "the requested resource was not found")
	default:
		return nil
	}
}
