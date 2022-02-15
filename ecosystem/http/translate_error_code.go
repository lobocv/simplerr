package simplehttp

import (
	"github.com/lobocv/simplerr"
	"net/http"
)

// HTTPStatus is the HTTP status code
type HTTPStatus = int

var mapping map[simplerr.Code]HTTPStatus

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
}

func init() {
	SetMapping(DefaultMapping())
}

// SetStatus sets the http.Response status from the error code in the provided error, if it is a SimpleError
// If error is nil or not a SimpleError, the status will not be set.
func SetStatus(r *http.Response, err error) {
	if err == nil {
		return
	}
	e := simplerr.As(err)
	if e == nil {
		r.StatusCode = http.StatusInternalServerError
		return
	}
	code, ok := mapping[e.GetCode()]
	if !ok {
		return
	}
	r.StatusCode = code
}
