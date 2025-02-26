package errors

import (
	nethttp "net/http"

	"github.com/go-kratos/kratos/v2/errors"
)

const (
	// UnknownCode is unknown code for error info.
	UnknownCode = nethttp.StatusInternalServerError
	// UnknownReason is unknown reason for error info.
	UnknownReason = "UNKNOWN"

	// SupportPackageIsVersion1 this constant should not be referenced by any other code.
	SupportPackageIsVersion1 = true
)

var (
	BadRequest = errors.BadRequest
	NotFound   = errors.NotFound
)

// MethodNotAllowed new MethodNotAllowed error that is mapped to a 405 response.
func MethodNotAllowed(reason, message string) *Error {
	return New(nethttp.StatusMethodNotAllowed, reason, message)
}

// NotAcceptable new NotAcceptable error that is mapped to a 406
func NotAcceptable(reason, message string) *Error {
	return New(nethttp.StatusNotAcceptable, reason, message)
}
