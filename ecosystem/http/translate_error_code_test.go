package simplehttp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lobocv/simplerr"
)

func TestTranslateErrorCode(t *testing.T) {

	testCases := []struct {
		err                error
		expected           HTTPStatus
		expectMappingFound bool
	}{
		{fmt.Errorf("something"), http.StatusInternalServerError, false},
		{simplerr.New("something").Code(simplerr.CodeUnknown), http.StatusInternalServerError, false},
		{simplerr.New("something").Code(simplerr.CodePermissionDenied), http.StatusForbidden, true},
		{simplerr.New("something").Code(simplerr.CodeCanceled), http.StatusRequestTimeout, true},
		{simplerr.New("something").Code(simplerr.CodeConstraintViolated), http.StatusInternalServerError, false},
		{fmt.Errorf("wrapped: %w", simplerr.New("something").Code(simplerr.CodeUnauthenticated)), http.StatusUnauthorized, true},
		{fmt.Errorf("opaque: %s", simplerr.New("something").Code(simplerr.CodeUnauthenticated)), http.StatusInternalServerError, false},
		{simplerr.Wrap(simplerr.New("something").Code(simplerr.CodePermissionDenied)), http.StatusForbidden, true},
		{nil, 200, false}, // default code for httptest.ResponseRecorder is 200
	}

	// Alter the default mapping
	m := DefaultMapping()
	m[simplerr.CodeCanceled] = http.StatusRequestTimeout
	SetMapping(m)
	SetDefaultErrorStatus(http.StatusInternalServerError)

	for ii, tc := range testCases {
		r := httptest.NewRecorder()
		SetStatus(r, tc.err)

		// Check that GetStatus returns a status when there is a mapping
		gotStatus, mappingFound := GetStatus(tc.err)
		require.Equal(t, tc.expectMappingFound, mappingFound, fmt.Sprintf("test case %d failed", ii))
		if mappingFound {
			require.Equal(t, tc.expected, gotStatus, fmt.Sprintf("test case %d failed", ii))
		}

		require.Equal(t, tc.expected, r.Code, fmt.Sprintf("test case %d failed", ii))
	}

}
