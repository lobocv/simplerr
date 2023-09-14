package simplehttp

import (
	"fmt"
	"github.com/lobocv/simplerr"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

type dummyTransport struct {
	response *http.Response
	err      error
}

func (d dummyTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return d.response, d.err
}

func TestRoundTripperConvertCode(t *testing.T) {

	for status, expectedCode := range inverseMapping {
		originalResponse := &http.Response{}
		originalResponse.StatusCode = status

		rt := EnableHTTPStatusErrors(dummyTransport{
			response: originalResponse,
		})
		resp, err := rt.RoundTrip(nil)

		require.NotNil(t, simplerr.As(err), fmt.Sprintf("'%s' failed: simplerr not returned", http.StatusText(status)))
		require.True(t, simplerr.HasErrorCode(err, expectedCode), fmt.Sprintf("'%s' failed: unexpected code", http.StatusText(status)))
		require.Nil(t, resp)
	}
}

func TestRoundTripperNoConversionFound(t *testing.T) {
	originalResponse := &http.Response{}
	originalResponse.StatusCode = http.StatusOK

	rt := EnableHTTPStatusErrors(dummyTransport{
		response: originalResponse,
	})
	resp, err := rt.RoundTrip(nil)
	require.NoError(t, err)
	require.Equal(t, originalResponse, resp)
}

func TestRoundTripperAttrResponse(t *testing.T) {

	originalResponse := &http.Response{StatusCode: http.StatusTeapot}

	rt := EnableHTTPStatusErrors(dummyTransport{
		response: originalResponse,
	})

	resp, err := rt.RoundTrip(nil)
	expectedCode := simplerr.CodeUnknown
	require.True(t, simplerr.HasErrorCode(err, expectedCode), "failed to convert error code")
	require.Nil(t, resp)

	gotOriginalResponse := GetHTTPResponseAttr(err)
	require.Equal(t, originalResponse, gotOriginalResponse)

	require.Nil(t, GetHTTPResponseAttr(nil))
	require.Nil(t, GetHTTPResponseAttr(simplerr.New("something").Attr(attrHTTPResponse, "not an *http.Response")))
}

func TestRoundTripperErrorOnUnderlyingRoundTripper(t *testing.T) {

	rt := EnableHTTPStatusErrors(dummyTransport{
		err: fmt.Errorf("some error"),
	})
	resp, err := rt.RoundTrip(nil)

	require.Nil(t, resp)
	require.Errorf(t, err, "some error")

}
