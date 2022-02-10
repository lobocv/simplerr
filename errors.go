package simplerr

import (
	"errors"
	"fmt"
)

// SimpleError is an implementation of the `error` interface which provides functionality
// to ease in the operating and handling of errors in applications.
type SimpleError struct {
	// err is the underlying error
	err error
	// code is the error code of the error defined in the registry
	code Code
	// silent is a flag that signals that this error should be recorded or logged silently
	// eg. This error should not be logged at all
	silent bool
	// skip is a flag that signals that, from the application's perspective, this error is a benign error.
	// eg. This error can be logged at INFO level and then discarded.
	skip bool
	// skipReason is the reason this error was marked as "skip"
	skipReason string
}

func New(err error) *SimpleError {
	return &SimpleError{err: err, code: CodeUnknown}
}

// Newf creates a new SimpleError from a formatted string
func Newf(_fmt string, args ...interface{}) *SimpleError {
	return New(fmt.Errorf(_fmt, args...))
}

// Error satisfies the `error` interface
func (e *SimpleError) Error() string {
	return e.err.Error()
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

// Skip marks the error as "skip"
func (e *SimpleError) Skip() *SimpleError {
	e.skip = true
	return e
}

// SkipReason marks the error as "skip" and attaches a reason it was marked skip.
func (e *SimpleError) SkipReason(reason string) *SimpleError {
	e.skip = true
	e.skipReason = reason
	return e
}

// GetSkipReason returns the skip reason and whether the error was marked as skip
func (e *SimpleError) GetSkipReason() (string, bool) {
	return e.skipReason, e.skip
}

// GetSilent returns if the error is silent
func (e *SimpleError) GetSilent() bool {
	return e.silent
}

// Silence sets the error as silent
func (e *SimpleError) Silence() *SimpleError {
	e.silent = true
	return e
}

// Unwrap implement the interface required for error unwrapping. It returns the underlying (wrapped) error
func (e *SimpleError) Unwrap() error {
	return errors.Unwrap(e.err)
}
