package simplehttp

import (
	"fmt"
	"github.com/lobocv/simplerr"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTranslateErrorCode(t *testing.T) {

	testCases := []struct {
		err      error
		expected HTTPStatus
	}{
		{fmt.Errorf("something"), http.StatusInternalServerError},
		{simplerr.New("something").Code(simplerr.CodeUnknown), http.StatusInternalServerError},
		{simplerr.New("something").Code(simplerr.CodePermissionDenied), http.StatusForbidden},
		{simplerr.New("something").Code(simplerr.CodeCanceled), http.StatusRequestTimeout},
		{simplerr.New("something").Code(simplerr.CodeConstraintViolated), http.StatusInternalServerError},
		{fmt.Errorf("wrapped: %w", simplerr.New("something").Code(simplerr.CodeUnauthenticated)), http.StatusUnauthorized},
		{fmt.Errorf("opaque: %s", simplerr.New("something").Code(simplerr.CodeUnauthenticated)), http.StatusInternalServerError},
		{simplerr.Wrap(simplerr.New("something").Code(simplerr.CodePermissionDenied)), http.StatusForbidden},
		{nil, 200}, // default code for httptest.ResponseRecorder is 200
	}

	// Alter the default mapping
	m := DefaultMapping()
	m[simplerr.CodeCanceled] = http.StatusRequestTimeout
	SetMapping(m)
	SetDefaultErrorCode(http.StatusInternalServerError)

	for ii, tc := range testCases {
		r := httptest.NewRecorder()
		SetStatus(r, tc.err)

		require.Equal(t, tc.expected, r.Code, fmt.Sprintf("test case %d failed", ii))
	}

}
