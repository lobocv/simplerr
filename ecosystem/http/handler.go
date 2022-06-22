package simplehttp

import (
	"net/http"
)

// ErrorHandler is an error handling function for HTTP handlers
type ErrorHandler func(http.ResponseWriter, *http.Request, error)

// DefaultErrorHandler is the default error handling function for HTTP handlers, it
// Sets the response status based on the handler's returned error using the SimpleError to HTTP mapping.
// The DefaultErrorHandler can be changed to also provide other error handling such as logging.
var DefaultErrorHandler ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
	SetStatus(w, err)
}

// Handler is analogous to http.Handler but returns an error
type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request) error
}

// HandlerAdapter adapts the Handler interface to the http.Handler interface
type HandlerAdapter struct {
	h          Handler
	errHandler ErrorHandler
}

// NewHandlerAdapter returns a HandlerAdapter that can be used with the standard library http package.
func NewHandlerAdapter(h Handler, opts ...HandlerOption) *HandlerAdapter {
	ha := &HandlerAdapter{h: h, errHandler: DefaultErrorHandler}
	for _, opt := range opts {
		opt(ha)
	}
	return ha
}

// ServeHTTP calls the underlying handler's ServeHTTP method and calls SetStatus on the returned error
func (h HandlerAdapter) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	err := h.h.ServeHTTP(writer, request)
	h.errHandler(writer, request, err)
}

// HandlerFunc is analogous to http.HandlerFunc but returns an error
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// ServerHTTP implements the Handler interface
func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return h(w, r)
}

// Adapter returns a http.HandlerFunc which calls SetStatus on the returned error
func (h HandlerFunc) Adapter(opts ...HandlerOption) http.HandlerFunc {
	return NewHandlerFuncAdapter(h, opts...)
}

// NewHandlerFuncAdapter returns a http.HandlerFunc which calls SetStatus on the returned error
func NewHandlerFuncAdapter(h HandlerFunc, opts ...HandlerOption) http.HandlerFunc {
	ha := NewHandlerAdapter(h, opts...)
	return ha.ServeHTTP
}
