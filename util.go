package simplerr

import (
	"errors"
	"fmt"
)

// Wrap wraps the error in a SimpleError. It defaults the error code to CodeUnknown.
func Wrap(err error) *SimpleError {
	return &SimpleError{parent: err, rawStackFrames: rawStackFrames(3)}
}

// Wrapf returns a new SimpleError by wrapping an error with a formatted message string.
// It defaults the error code to CodeUnknown
func Wrapf(err error, msg string, a ...interface{}) *SimpleError {
	msg = fmt.Sprintf(msg, a...)
	return &SimpleError{parent: err, msg: msg, rawStackFrames: rawStackFrames(3)}
}

// As attempts to find a SimpleError in the chain of errors, similar to errors.As().
// Note that this will NOT match structs which embed a *SimpleError.
func As(err error) *SimpleError {
	var expecterErr *SimpleError
	if ok := errors.As(err, &expecterErr); !ok {
		return nil
	}
	return expecterErr
}

// HasErrorCode checks the error code of an error if it is a SimpleError{}.
// nil errors or errors that are not SimplErrors return false.
func HasErrorCode(err error, code Code) bool {
	type CodedError interface {
		GetCode() Code
	}
	if err == nil {
		return false
	}
	// The err may wrap another CodedError who's value may be set. Therefore, we only exit if we find
	// a matching code, otherwise we traverse the remaining error chain
	codedErr, ok := err.(CodedError)
	if ok && codedErr.GetCode() == code {
		return true
	}

	return HasErrorCode(errors.Unwrap(err), code)
}

// HasErrorCodes looks for the specified error codes in the chain of errors.
// It returns the first code in the list that is found in the chain and a boolean for whether
// anything was found.
func HasErrorCodes(err error, codes ...Code) (Code, bool) {
	type CodedError interface {
		GetCode() Code
	}
	if err == nil {
		return 0, false
	}
	for _, code := range codes {
		// The err may wrap another CodedError who's value may be set. Therefore, we only exit if we find
		// a matching code, otherwise we traverse the remaining error chain
		codedErr, ok := err.(CodedError)
		if ok && codedErr.GetCode() == code {
			return code, true
		}
	}

	return HasErrorCodes(errors.Unwrap(err), codes...)
}

// IsBenign checks the error or any error in the chain, is marked as benign.
// It also returns the reason the error was marked benign. Benign errors should be logged or handled
// less severely than non-benign errors. For example, you may choose to log benign errors at INFO level,
// rather than ERROR.
func IsBenign(err error) (string, bool) {
	type BenignError interface {
		GetBenignReason() (string, bool)
	}
	if err == nil {
		return "", false
	}
	// The err may wrap another BenignError who's value may be set to true. Therefore, we only exit if the
	// benign flag is true, otherwise we keep traversing the error chain
	benignErr, ok := err.(BenignError)
	if ok {
		if reason, benign := benignErr.GetBenignReason(); benign {
			return reason, true
		}
	}

	return IsBenign(errors.Unwrap(err))
}

// IsSilent checks the error or any error in the chain, is marked silent.
// Silent errors should not need to be logged at all.
func IsSilent(err error) bool {
	type SilencedError interface {
		GetSilent() bool
	}
	if err == nil {
		return false
	}
	// The err may wrap another SilencedError who's value may be set to true. Therefore, we only exit if the
	// silent flag is true, otherwise we keep traversing the error chain
	silenterr, ok := err.(SilencedError)
	if ok && silenterr.GetSilent() {
		return true
	}

	return IsSilent(errors.Unwrap(err))
}

// IsRetriable checks the error or any error in the chain is retriable, meaning the caller should retry the operation
// which caused this error in hopes of it succeeding.
// Errors are assumed not retriable by default unless an error in the chain says otherwise. A single error in the chain
// that is retriable will make the entire error retriable.
func IsRetriable(err error) bool {
	type RetriableError interface {
		GetRetriable() bool
	}

	// If there is no error, then it is not retriable
	if err == nil {
		return false
	}

	// The error may wrap another RetriableError which may not be retriable. Therefore, we only exit if the
	// error is not retriable, otherwise we keep traversing the error chain
	retriableErr, ok := err.(RetriableError)
	if ok && retriableErr.GetRetriable() {
		return true
	}

	// Check errors further in the chain
	return IsRetriable(errors.Unwrap(err))
}

// ExtractAuxiliary extracts a superset of auxiliary data from all errors in the chain.
// Wrapper error auxiliary data take precedent over later errors.
func ExtractAuxiliary(err error) map[string]interface{} {
	type AuxHolder interface {
		GetAuxiliary() map[string]interface{}
	}
	if err == nil {
		return nil
	}
	aux := map[string]interface{}{}

	e := err
	for e != nil {
		if auxHolder, ok := e.(AuxHolder); ok {
			for k, v := range auxHolder.GetAuxiliary() {
				aux[k] = v
			}
		}

		e = errors.Unwrap(e)
	}

	return aux
}

// GetAttribute gets the first instance of the key in the error chain.
// This can be used to define attributes on the error that do not have first-class support
// with simplerr. Much like keys in the `context` package, the `key` should be a custom type so it does
// not have naming collisions with other values.
func GetAttribute(err error, key interface{}) (interface{}, bool) {
	type AttrHolder interface {
		GetAttribute(key interface{}) (interface{}, bool)
	}
	if err == nil {
		return nil, false
	}
	e := err
	for e != nil {
		var attr interface{}
		if attrHolder, ok := e.(AttrHolder); ok {
			if attr, ok = attrHolder.GetAttribute(key); ok {
				return attr, true
			}
		}

		e = errors.Unwrap(e)
	}

	return nil, false
}
