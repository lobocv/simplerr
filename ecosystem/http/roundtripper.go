package simplehttp

import (
	"github.com/lobocv/simplerr"
	"net/http"
)

type attr int

const (
	attrHTTPResponse = attr(1)
)

// roundTripper is a wrapper around the given http.RoundTripper that converts 4XX and 5XX series errors to SimpleErrors
type roundTripper struct {
	rt http.RoundTripper
}

// RoundTrip calls the underlying RoundTripper and converts any 4XX or 5XX series errors to SimpleErrors
func (s roundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	resp, err := s.rt.RoundTrip(request)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 600 {
		code, _ := GetCode(resp.StatusCode)
		serr := simplerr.New("%s", resp.Status).
			Code(code).
			Attr(attrHTTPResponse, resp)
		return nil, serr
	}

	return resp, nil
}

// EnableHTTPStatusErrors wraps the http.RoundTripper in middleware that converts 4XX and 5XX series errors to SimpleErrors
// with the code defined in the inverse mapping.
func EnableHTTPStatusErrors(rt http.RoundTripper) http.RoundTripper {
	return roundTripper{rt: rt}
}

// GetHTTPResponseAttr gets the *http.Response attached to the error, if it exists.
func GetHTTPResponseAttr(err error) *http.Response {
	v, ok := simplerr.GetAttribute(err, attrHTTPResponse)
	if !ok {
		return nil
	}

	resp, ok := v.(*http.Response)
	if !ok {
		return nil
	}

	return resp
}
