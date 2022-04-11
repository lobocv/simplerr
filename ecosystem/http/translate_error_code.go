package simplehttp

import (
	"github.com/lobocv/simplerr"
	"net/http"
)

// HTTPStatus is the HTTP status code
type HTTPStatus = int

var mapping map[simplerr.Code]HTTPStatus
var simplerrCodes []simplerr.Code
var defaultErrorCode = http.StatusInternalServerError

// DefaultMapping returns the default mapping of SimpleError codes to HTTP status codes
func DefaultMapping() map[simplerr.Code]HTTPStatus {
	var m = map[simplerr.Code]HTTPStatus{
		simplerr.CodeUnknown:           http.StatusInternalServerError,
		simplerr.CodeNotFound:          http.StatusNotFound,
		simplerr.CodeDeadlineExceeded:  http.StatusRequestTimeout,
		simplerr.CodePermissionDenied:  http.StatusForbidden,
		simplerr.CodeUnauthenticated:   http.StatusUnauthorized,
		simplerr.CodeNotImplemented:    http.StatusNotImplemented,
		simplerr.CodeInvalidArgument:   http.StatusBadRequest,
		simplerr.CodeResourceExhausted: http.StatusTooManyRequests,
	}
	return m
}

// SetMapping sets the mapping from simplerr.Code to HTTP status code
func SetMapping(m map[simplerr.Code]HTTPStatus) {
	mapping = m
	simplerrCodes = []simplerr.Code{}
	// Get a list of simplerr codes to search for in the error chain
	for c := range m {
		// Ignore CodeUnknown because it is the default code
		if c == simplerr.CodeUnknown {
			continue
		}
		simplerrCodes = append(simplerrCodes, c)
	}
}

// SetDefaultErrorCode changes the default HTTP status code for when a translation could not be found.
// The default status code is 500.
func SetDefaultErrorCode(code int) {
	defaultErrorCode = code
}

func init() {
	SetMapping(DefaultMapping())
}

// SetStatus sets the http.Response status from the error code in the provided error, if it is a SimpleError
// If error is nil or not a SimpleError, the status will not be set.
func SetStatus(r http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	e := simplerr.As(err)
	if e == nil {
		r.WriteHeader(defaultErrorCode)
		return
	}

	// Check if the error has any of the codes in it's chain
	// If we cant find any matches, set it to the default error code
	code, ok := simplerr.HasErrorCodes(e, simplerrCodes...)
	if !ok {
		r.WriteHeader(defaultErrorCode)
		return
	}

	// Get the HTTP code, this lookup should never fail
	r.WriteHeader(mapping[code])
}
