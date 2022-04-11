package simplehttp

import (
	"net/http"
)

// Handler is analogous to http.Handler but returns an error
type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request) error
}

// HandlerAdapter adapts the Handler interface to the http.Handler interface
type HandlerAdapter struct {
	h Handler
}

func NewHandlerAdapter(h Handler) *HandlerAdapter {
	return &HandlerAdapter{h: h}
}

// ServeHTTP calls the underlying handler's ServeHTTP method and calls SetStatus on the returned error
func (h HandlerAdapter) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	err := h.h.ServeHTTP(writer, request)
	SetStatus(writer, err)
}

// HandlerFunc is analogous to http.HandlerFunc but returns an error
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// Adapter returns a http.HandlerFunc which calls SetStatus on the returned error
func (h HandlerFunc) Adapter() http.HandlerFunc {
	return NewHandlerFuncAdapter(h)
}

// NewHandlerFuncAdapter returns a http.HandlerFunc which calls SetStatus on the returned error
func NewHandlerFuncAdapter(h HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		err := h(writer, request)
		SetStatus(writer, err)
	}
}
