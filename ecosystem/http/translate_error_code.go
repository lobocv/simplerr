package simplehttp

import (
	"net/http"
	"sync"

	"github.com/lobocv/simplerr"
)

// HTTPStatus is the HTTP status code
type HTTPStatus = int

var (
	mapping map[simplerr.Code]HTTPStatus

	inverseMapping map[HTTPStatus]simplerr.Code

	simplerrCodes      []simplerr.Code
	defaultErrorStatus = http.StatusInternalServerError

	lock = sync.Mutex{}
)

// DefaultMapping returns the default mapping of SimpleError codes to HTTP status codes
func DefaultMapping() map[simplerr.Code]HTTPStatus {
	var m = map[simplerr.Code]HTTPStatus{
		simplerr.CodeUnknown:           http.StatusInternalServerError,
		simplerr.CodeNotFound:          http.StatusNotFound,
		simplerr.CodeDeadlineExceeded:  http.StatusRequestTimeout,
		simplerr.CodePermissionDenied:  http.StatusForbidden,
		simplerr.CodeUnauthenticated:   http.StatusUnauthorized,
		simplerr.CodeNotImplemented:    http.StatusNotImplemented,
		simplerr.CodeMalformedRequest:  http.StatusBadRequest,
		simplerr.CodeInvalidArgument:   http.StatusUnprocessableEntity,
		simplerr.CodeUnavailable:       http.StatusServiceUnavailable,
		simplerr.CodeMissingParameter:  http.StatusUnprocessableEntity,
		simplerr.CodeResourceExhausted: http.StatusTooManyRequests,
	}
	return m
}

// DefaultInverseMapping returns the default mapping of HTTP status codes to SimpleError code
func DefaultInverseMapping() map[HTTPStatus]simplerr.Code {
	var m = map[HTTPStatus]simplerr.Code{
		http.StatusInternalServerError: simplerr.CodeUnknown,
		http.StatusNotFound:            simplerr.CodeNotFound,
		http.StatusRequestTimeout:      simplerr.CodeDeadlineExceeded,
		http.StatusForbidden:           simplerr.CodePermissionDenied,
		http.StatusUnauthorized:        simplerr.CodeUnauthenticated,
		http.StatusNotImplemented:      simplerr.CodeNotImplemented,
		http.StatusBadRequest:          simplerr.CodeMalformedRequest,
		http.StatusUnprocessableEntity: simplerr.CodeInvalidArgument,
		http.StatusServiceUnavailable:  simplerr.CodeUnavailable,
		http.StatusMethodNotAllowed:    simplerr.CodeMalformedRequest,
		http.StatusTooManyRequests:     simplerr.CodeResourceExhausted,
	}
	return m
}

// SetMapping sets the mapping from simplerr.Code to HTTP status code
func SetMapping(m map[simplerr.Code]HTTPStatus) {
	lock.Lock()
	defer lock.Unlock()
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

// SetInverseMapping sets the mapping from HTTP status code to simplerr.Code
func SetInverseMapping(m map[HTTPStatus]simplerr.Code) {
	lock.Lock()
	defer lock.Unlock()
	inverseMapping = m
}

// SetDefaultErrorStatus changes the default HTTP status code for when a translation could not be found.
// The default status code is 500.
func SetDefaultErrorStatus(code int) {
	defaultErrorStatus = code
}

func init() {
	SetMapping(DefaultMapping())
	SetInverseMapping(DefaultInverseMapping())
}

// SetStatus sets the http.Response status from the error code in the provided error.
// It returns the HTTPStatus that was written
// If the error contains a SimpleError, then the status is determined by the mapping.
// If the error is not a SimpleError then the default error status code will be set.
// If the error is nil, then no status will be set.
func SetStatus(r http.ResponseWriter, err error) HTTPStatus {
	if err == nil {
		return 0
	}
	httpStatus, ok := GetStatus(err)
	if !ok {
		httpStatus = defaultErrorStatus
	}
	r.WriteHeader(httpStatus)
	return httpStatus
}

// GetStatus returns the HTTP status that the error maps to if the provided error is a SimpleError.
// If a mapping could not be found or the error is nil, then the boolean second argument is returned as false
func GetStatus(err error) (status HTTPStatus, found bool) {
	if err == nil {
		return 0, false
	}

	// Check if the error is a SimpleError
	serr := simplerr.As(err)
	if serr == nil {
		return 0, false
	}

	// Check if the error has any of the codes in its chain of errors
	code, ok := simplerr.HasErrorCodes(serr, simplerrCodes...)
	if !ok {
		return 0, false
	}

	// Get the HTTP code, this lookup should never fail because the every item in simplerrCodes will have an
	// entry in the mapping
	httpCode := mapping[code]

	return httpCode, true
}

// GetCode gets the simplerror Code that corresponds to the HTTPStatus. It returns CodeUnknown if it cannot map the status.
func GetCode(status HTTPStatus) (code simplerr.Code, found bool) {
	code, ok := inverseMapping[status]
	if !ok {
		return simplerr.CodeUnknown, false
	}
	return code, true
}
