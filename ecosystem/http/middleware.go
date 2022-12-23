package simplehttp

import (
	"net/http"
)

// Middleware is an HTTP middleware
type Middleware = func(Handler) Handler

// middleware is a wrapper around a user-provided Middleware that also calls the error handler on any results
func middleware(m Middleware) Middleware {
	return func(h Handler) Handler {
		fn := func(w http.ResponseWriter, r *http.Request) error {
			err := m(h).ServeHTTP(w, r)
			if err != nil {
				DefaultErrorHandler(w, r, err)
			}
			return err
		}

		return HandlerFunc(fn)
	}
}

// ApplyMiddleware applies the given middlewares to the handler
func ApplyMiddleware(h HandlerFunc, mw ...Middleware) HandlerFunc {
	for _, m := range mw {
		h = middleware(m)(h).ServeHTTP
	}
	return h
}

// middlewareAdapter adapts the http.Handler into a Handler
type middlewareAdapter struct {
	h http.HandlerFunc
}

// ServeHTTP satisfies the Handler interface but disregards returning any errors because it uses the http.Handler
func (a middlewareAdapter) ServeHTTP(writer http.ResponseWriter, request *http.Request) error {
	a.h.ServeHTTP(writer, request)
	return nil
}

// MiddlewareAdapter is an adapter for turning standard library middleware into simplehttp compatible middleware
func MiddlewareAdapter(mw func(handler http.Handler) http.Handler) Middleware {

	return func(handler Handler) Handler {

		// Create a http.HandlerFunc from the Handler by using a thin wrapper that ignores the error
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = handler.ServeHTTP(w, r)
		})

		// Implement the middleware
		hh := mw(h)

		// Use an adapter over the http.Handler to make it a Handler
		return middlewareAdapter{h: hh.ServeHTTP}

	}

}

// middlewareReverseAdapter adapts the Handler into a http.Handler
type middlewareReverseAdapter struct {
	h HandlerFunc
}

// ServeHTTP satisfies the Handler interface, calls the error handler but disregards returning any errors because
// it needs to satisfy the http.Handler interface
func (a middlewareReverseAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := a.h.ServeHTTP(w, r)
	if err != nil {
		DefaultErrorHandler(w, r, err)
	}
}

// MiddlewareReverseAdapter is an adapter for turning simplehttp compatible middleware to standard library middleware
// This is useful for interfacing with other libraries like gorilla.mux which expect standard library middleware
func MiddlewareReverseAdapter(mw Middleware) func(handler http.Handler) http.Handler {

	return func(handler http.Handler) http.Handler {

		// Create a HandlerFunc from the http.Handler by using a thin wrapper that changes the signature
		h := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			handler.ServeHTTP(w, r)
			return nil
		})

		// Implement the middleware
		hh := mw(h)

		// Use an adapter over the Handler to make it a http.Handler
		return middlewareReverseAdapter{h: hh.ServeHTTP}
	}

}
