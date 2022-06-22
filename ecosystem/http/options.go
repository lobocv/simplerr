package simplehttp

// HandlerOption are options to change the behaviour of the HandlerAdapter
type HandlerOption func(adapter *HandlerAdapter)

// WithErrorHandler is an HandlerOption to change the error handling function
func WithErrorHandler(h ErrorHandler) HandlerOption {
	return func(a *HandlerAdapter) {
		a.errHandler = h
	}
}
