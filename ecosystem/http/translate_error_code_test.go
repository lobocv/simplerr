package simplehttp

import (
	"fmt"
	"github.com/lobocv/simplerr"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestTranslateErrorCode(t *testing.T) {

	testCases := []struct {
		err      error
		expected HTTPStatus
	}{
		{fmt.Errorf("something"), 0},
		{simplerr.New("something").Code(simplerr.CodePermissionDenied), http.StatusForbidden},
		{simplerr.New("something").Code(simplerr.CodeCanceled), http.StatusRequestTimeout},
		{fmt.Errorf("wrapped: %w", simplerr.New("something").Code(simplerr.CodeUnauthenticated)), http.StatusUnauthorized},
		{fmt.Errorf("opaque: %s", simplerr.New("something").Code(simplerr.CodeUnauthenticated)), 0},
		{nil, 0},
	}

	// Alter the default mapping
	m := DefaultMapping()
	m[simplerr.CodeCanceled] = http.StatusRequestTimeout
	SetMapping(m)

	for _, tc := range testCases {
		r := http.Response{}
		SetStatus(&r, tc.err)

		require.Equal(t, tc.expected, r.StatusCode)
	}

}
