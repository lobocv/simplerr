package simplerr

import (
	"fmt"
	"log/slog"
	"slices"
)

// attribute is a key value pair for attributes on errors
type attribute struct {
	Key, Value interface{}
}

// SimpleError is an implementation of the `error` interface which provides functionality
// to ease in the operating and handling of errors in applications.
type SimpleError struct {
	// parent is the error being wrapped
	parent error
	// msg is the error message
	msg string
	// code is the error code of the error defined in the registry
	code Code
	// silent is a flag that signals that this error should be recorded or logged silently on the server side
	// eg. This error should not be logged at all
	silent bool
	// benign is a flag that signals that, from the server's perspective, this error is a benign error.
	// eg. This error can be logged at INFO level and then discarded.
	benign bool
	// benignReason is the reason this error was marked as "benign"
	benignReason string
	// retriable is a flag indicating that this error is transient and that the user should retry the operation
	retriable bool
	// auxiliary are auxiliary informational fields that can be attached to the error
	auxiliary map[string]interface{}
	// logger is a scoped logger that can be attached to the error
	logger *slog.Logger
	// attr is a list of custom attributes attached the error
	attr []attribute
	// stackTrace is the call stack trace for the error
	rawStackFrames []uintptr
}

// New creates a new SimpleError from a formatted string
func New(_fmt string, args ...interface{}) *SimpleError {
	rawFrames := rawStackFrames(3)
	return &SimpleError{msg: fmt.Sprintf(_fmt, args...), code: CodeUnknown, rawStackFrames: rawFrames}
}

// Error satisfies the `error` interface. It uses the `simplerr.Formatter` to generate an error string.
func (e *SimpleError) Error() string {
	if parent := e.Unwrap(); parent != nil {
		return fmt.Sprintf("%s: %s", e.GetMessage(), parent.Error())
	}
	return e.msg
}

// Message sets the message text on the error. This message it used to wrap the underlying error, if it exists.
func (e *SimpleError) Message(msg string, args ...interface{}) *SimpleError {
	e.msg = fmt.Sprintf(msg, args...)
	return e
}

// GetMessage gets the error string for this error, exclusive of any wrapped errors.
func (e *SimpleError) GetMessage() string {
	return e.msg
}

// GetCode returns the error code as defined in the registry
func (e *SimpleError) GetCode() Code {
	return e.code
}

// Code sets the error code. The assigned code should be defined in the registry.
func (e *SimpleError) Code(code Code) *SimpleError {
	e.code = code
	return e
}

// Benign marks the error as "benign". A benign error is an error that depends on the context of the caller.
// eg a NotFoundError is only an error if the caller is expecting the entity to exist.
// These errors can usually be logged less severely (ie at INFO rather than ERROR level)
func (e *SimpleError) Benign() *SimpleError {
	e.benign = true
	return e
}

// BenignReason marks the error as "benign" and attaches a reason it was marked benign.
// A benign error is an error depending on the context of the caller.
// eg a NotFoundError is only an error if the caller is expecting the entity to exist
// These errors can usually be logged less severely (ie at INFO rather than ERROR level)
func (e *SimpleError) BenignReason(reason string) *SimpleError {
	e.benign = true
	e.benignReason = reason
	return e
}

// GetBenignReason returns the benign reason and whether the error was marked as benign
// ie. This error can be logged at INFO level and then discarded.
func (e *SimpleError) GetBenignReason() (string, bool) {
	return e.benignReason, e.benign
}

// GetSilent returns a flag that signals that this error should be recorded or logged silently on the server side
// ie. This error should not be logged at all
func (e *SimpleError) GetSilent() bool {
	return e.silent
}

// Silence sets the error as silent. Silent errors can be ignored by loggers.
func (e *SimpleError) Silence() *SimpleError {
	e.silent = true
	return e
}

// GetRetriable returns a flag that signals that the operation which created this error is transient and that the
// user should retry the operation in hopes of it succeeding.
func (e *SimpleError) GetRetriable() bool {
	return e.retriable
}

// Retriable sets the error as retriable.
func (e *SimpleError) Retriable() *SimpleError {
	e.retriable = true
	return e
}

// GetAuxiliary gets the auxiliary informational data attached to this error.
// This key-value data can be attached to structured loggers.
func (e *SimpleError) GetAuxiliary() map[string]interface{} {
	return e.auxiliary
}

// GetAttribute gets an attribute attached to this specific SimpleError. It does NOT traverse the error chain.
// This can be used to define attributes on the error that do not have first-class support
// with simplerr. Much like keys in the `context` package, the `key` should be a custom type so it does
// not have naming collisions with other values.
func (e *SimpleError) GetAttribute(key interface{}) (interface{}, bool) {
	for _, attr := range e.attr {
		if attr.Key == key {
			return attr.Value, true
		}
	}
	return nil, false
}

// Aux attaches auxiliary informational data to the error as key value pairs.
// All keys must be of type `string` and have a value. Keys without values are ignored.
// This auxiliary data can be retrieved by using `ExtractAuxiliary()` and attached to structured loggers.
// Do not use this to detect any attributes on the error, instead use Attr().`
func (e *SimpleError) Aux(kv ...interface{}) *SimpleError {
	if e.auxiliary == nil {
		e.auxiliary = map[string]interface{}{}
	}
	var key interface{}
	for _, item := range kv {
		if key == nil {
			key = item
			continue
		}
		keyStr, ok := key.(string)
		if ok {
			e.auxiliary[keyStr] = item
		}
		key = nil
	}
	return e
}

// AuxMap attaches auxiliary informational data to the error from a map[string]interface{}.
// This auxiliary data can be retrieved by using `ExtractAuxiliary()` and attached to structured loggers.
// Do not use this to detect any attributes on the error, instead use Attr().`
func (e *SimpleError) AuxMap(aux map[string]interface{}) *SimpleError {
	if e.auxiliary == nil {
		e.auxiliary = map[string]interface{}{}
	}
	for k, v := range aux {
		e.auxiliary[k] = v
	}
	return e
}

// Attr attaches an attribute to the error that can be detected when handling the error.
// Attr() behaves similarly to `context.WithValue()`. Keys should be custom types in order to avoid naming collisions.
// Use `GetAttribute()` to get the value of the attribute.
func (e *SimpleError) Attr(key, value interface{}) *SimpleError {
	e.attr = append(e.attr, attribute{Key: key, Value: value})
	return e
}

// Logger attaches a structured logger to the error
func (e *SimpleError) Logger(l *slog.Logger) *SimpleError {
	e.logger = l
	return e
}

// GetLogger returns a structured logger with auxiliary information preset
// It uses the first found attached structured logger and otherwise uses the default logger
func (e *SimpleError) GetLogger() *slog.Logger {
	l := e.logger

	fields := make([]any, 0, 2*len(e.auxiliary))
	var errInChain = e
	for errInChain != nil {
		if l == nil && errInChain.logger != nil {
			l = errInChain.logger
		}

		for k, v := range errInChain.GetAuxiliary() {
			fields = append(fields, v, k) // append backwards because of the slice reversal
		}

		errInChain = As(errInChain.Unwrap())
	}
	// Reverse the slice so that the aux values at the top of the stack take precedent over the lower ones
	// For cases where there are conflicting keys
	slices.Reverse(fields)

	if l == nil {
		l = slog.Default()
	}

	return l.With(fields...)
}

// GetDescription returns the description of the error code on the error.
func (e *SimpleError) GetDescription() string {
	return registry.CodeDescription(e.code)
}

// StackTrace returns the stack trace at the point at which the error was raised.
func (e *SimpleError) StackTrace() []Call {
	return stackTrace(e.rawStackFrames)
}

// StackFrames returns a slice of pointers to program counters
// This method is primarily used to better integrate with sentry stack trace extraction
func (e *SimpleError) StackFrames() []uintptr {
	return e.rawStackFrames
}

// Unwrap implement the interface required for error unwrapping. It returns the underlying (wrapped) error.
func (e *SimpleError) Unwrap() error {
	return e.parent
}
