package simplehttp

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lobocv/simplerr"
)

type TestSuite struct {
	suite.Suite
}

func TestHTTP(t *testing.T) {
	s := new(TestSuite)

	// Change the default mappings to test that they apply
	m := DefaultMapping()
	m[simplerr.CodeCanceled] = http.StatusRequestTimeout
	SetMapping(m)
	SetDefaultErrorStatus(http.StatusInternalServerError)

	invM := DefaultInverseMapping()
	invM[http.StatusRequestTimeout] = simplerr.CodeCanceled
	SetInverseMapping(invM)
	suite.Run(t, s)
}

func (s *TestSuite) TestTranslateErrorCode() {

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
		{simplerr.New("something").Code(simplerr.CodeMalformedRequest), http.StatusBadRequest, true},
		{simplerr.New("something").Code(simplerr.CodeMissingParameter), http.StatusUnprocessableEntity, true},
		{simplerr.New("something").Code(simplerr.CodeInvalidArgument), http.StatusUnprocessableEntity, true},
		{fmt.Errorf("wrapped: %w", simplerr.New("something").Code(simplerr.CodeUnauthenticated)), http.StatusUnauthorized, true},
		{fmt.Errorf("opaque: %s", simplerr.New("something").Code(simplerr.CodeUnauthenticated)), http.StatusInternalServerError, false},
		{simplerr.Wrap(simplerr.New("something").Code(simplerr.CodePermissionDenied)), http.StatusForbidden, true},
		{nil, 200, false}, // default code for httptest.ResponseRecorder is 200
	}

	for ii, tc := range testCases {
		r := httptest.NewRecorder()
		SetStatus(r, tc.err)

		// Check that GetStatus returns a status when there is a mapping
		gotStatus, mappingFound := GetStatus(tc.err)
		s.Equal(tc.expectMappingFound, mappingFound, fmt.Sprintf("test case %d failed", ii))
		if mappingFound {
			s.Equal(tc.expected, gotStatus, fmt.Sprintf("test case %d failed", ii))
		}

		s.Equal(tc.expected, r.Code, fmt.Sprintf("test case %d failed", ii))
	}

}

func (s *TestSuite) TestTranslateStatusCode() {

	testCases := []struct {
		status             HTTPStatus
		expected           simplerr.Code
		expectMappingFound bool
	}{
		{http.StatusNotFound, simplerr.CodeNotFound, true},
		{http.StatusRequestTimeout, simplerr.CodeCanceled, true},
		{23587253923, simplerr.CodeUnknown, false},
	}

	for ii, tc := range testCases {
		gotCode, mappingFound := GetCode(tc.status)

		// Check that GetStatus returns a status when there is a mapping
		s.Equal(tc.expectMappingFound, mappingFound, fmt.Sprintf("test case %d failed", ii))
		if mappingFound {
			s.Equal(tc.expected, gotCode, fmt.Sprintf("test case %d failed", ii))
		}

		s.Equal(tc.expected, gotCode, fmt.Sprintf("test case %d failed", ii))
	}

}
