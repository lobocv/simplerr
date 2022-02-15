package simplerr

import (
	"fmt"
)

const (
	// maxStackFrames are the maximum number of stack frames returned with the error when using WithStackTrace
	maxStackFrames = 16
)

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
	// auxiliary are auxiliary informational fields that can be attached to the error
	auxiliary map[string]interface{}
	// stackTrace is the call stack trace for the error
	stackTrace []Call
}

// New creates a new SimpleError from a formatted string
func New(_fmt string, args ...interface{}) *SimpleError {
	return &SimpleError{msg: fmt.Sprintf(_fmt, args...), code: CodeUnknown, stackTrace: stackTrace(3)}
}

// Error satisfies the `error` interface
func (e *SimpleError) Error() string {
	return Formatter(e)
}

// Message sets the message text on the error
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

// Benign marks the error as "benign"
func (e *SimpleError) Benign() *SimpleError {
	e.benign = true
	return e
}

// BenignReason marks the error as "benign" and attaches a reason it was marked benign.
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

// Silence sets the error as silent
func (e *SimpleError) Silence() *SimpleError {
	e.silent = true
	return e
}

// GetAuxiliary gets the auxiliary informational data attached to this error
func (e *SimpleError) GetAuxiliary() map[string]interface{} {
	return e.auxiliary
}

// Aux attaches auxiliary informational data to the error as key value pairs.
// All keys must be of type `string` and have a value. Keys without values are ignored.
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
func (e *SimpleError) AuxMap(aux map[string]interface{}) *SimpleError {
	if e.auxiliary == nil {
		e.auxiliary = map[string]interface{}{}
	}
	for k, v := range aux {
		e.auxiliary[k] = v
	}
	return e
}

// GetDescription returns the description of the error code
func (e *SimpleError) GetDescription() string {
	return registry.CodeDescription(e.code)
}

// StackTrace returns the stack trace at the point at which the error was raised
func (e *SimpleError) StackTrace() []Call {
	return e.stackTrace
}

// Unwrap implement the interface required for error unwrapping. It returns the underlying (wrapped) error
func (e *SimpleError) Unwrap() error {
	return e.parent
}
