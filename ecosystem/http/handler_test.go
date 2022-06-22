package simplehttp

import (
	"github.com/lobocv/simplerr"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Endpoint struct {
	err error
}

func (e *Endpoint) ServeHTTP(_ http.ResponseWriter, _ *http.Request) error {
	return e.err
}

func TestHandlerAdapter(t *testing.T) {
	ep := &Endpoint{}
	h := NewHandlerAdapter(ep)

	t.Run("endpoint returns an error", func(t *testing.T) {
		ep.err = simplerr.New("something").Code(simplerr.CodeNotFound)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, nil)
		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("endpoint does not return an error", func(t *testing.T) {
		ep.err = nil
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, nil)
		require.Equal(t, rec.Code, http.StatusOK)
	})
}

func TestHandlerFunc(t *testing.T) {
	ep := func(writer http.ResponseWriter, request *http.Request) error {
		return simplerr.New("something").Code(simplerr.CodeNotFound)
	}

	rec := httptest.NewRecorder()
	NewHandlerFuncAdapter(ep)(rec, nil)
	require.Equal(t, http.StatusNotFound, rec.Code)

	rec = httptest.NewRecorder()
	HandlerFunc(ep).Adapter()(rec, nil)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

// TestErrorHandler tests changing the error handler
func TestErrorHandlerChangeDefault(t *testing.T) {

	// Change the default error handler
	var defaultErrHandlerCalled bool
	DefaultErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		SetStatus(w, err)
		defaultErrHandlerCalled = true
	}

	ep := func(writer http.ResponseWriter, request *http.Request) error {
		return simplerr.New("something").Code(simplerr.CodeNotFound)
	}

	rec := httptest.NewRecorder()
	NewHandlerFuncAdapter(ep)(rec, nil)
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.True(t, defaultErrHandlerCalled)
}

// TestErrorHandlerChangeSpecific changes the error handler for a specific handler
func TestErrorHandlerChangeSpecific(t *testing.T) {

	// Change the default error handler
	var defaultErrHandlerCalled bool
	DefaultErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		SetStatus(w, err)
		defaultErrHandlerCalled = true
	}

	// Create a custom error handler that we set on only a specific Handler
	var customErrHandlerCalled bool
	customErrHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		SetStatus(w, err)
		customErrHandlerCalled = true

	}

	ep := func(writer http.ResponseWriter, request *http.Request) error {
		return simplerr.New("something").Code(simplerr.CodeNotFound)
	}

	rec := httptest.NewRecorder()
	HandlerFunc(ep).Adapter(WithErrorHandler(customErrHandler))(rec, nil)
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.True(t, customErrHandlerCalled, "custom err handler should be called")
	require.False(t, defaultErrHandlerCalled, "default err handler should not be called")
}
