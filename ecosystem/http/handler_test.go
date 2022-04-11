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
	ep := HandlerFunc(func(writer http.ResponseWriter, request *http.Request) error {
		return simplerr.New("something").Code(simplerr.CodeNotFound)
	})
	rec := httptest.NewRecorder()
	ep.Handler()(rec, nil)
	require.Equal(t, http.StatusNotFound, rec.Code)
}
