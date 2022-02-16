package simplerr

import (
	"errors"
	"fmt"
)

// Wrap wraps the error in a SimpleError
func Wrap(err error) *SimpleError {
	return &SimpleError{parent: err, stackTrace: stackTrace(3)}
}

// Wrapf returns a new SimpleError by wrapping an error with a formatted message string
func Wrapf(err error, msg string, a ...interface{}) *SimpleError {
	msg = fmt.Sprintf(msg, a...)
	return &SimpleError{parent: err, msg: msg, stackTrace: stackTrace(3)}
}

// As attempts to find a SimpleError in the chain of errors, similar to errors.As()
func As(err error) *SimpleError {
	var expecterErr *SimpleError
	if ok := errors.As(err, &expecterErr); !ok {
		return nil
	}
	return expecterErr
}

// HasErrorCode checks the error code of an error if it is a SimpleError{}
func HasErrorCode(err error, code Code) bool {
	if e := As(err); e != nil {
		if e.code == code {
			return true
		}
		return HasErrorCode(e.parent, code)
	}
	return false
}

// HasErrorCodes looks for the specified error codes in the chain of errors.
// It returns the first code in the list that is found in the chain and a boolean for whether
// anything was found.
func HasErrorCodes(err error, codes ...Code) (Code, bool) {
	if e := As(err); e != nil {
		for _, c := range codes {
			if c == e.code {
				return c, true
			}
		}

		return HasErrorCodes(e.parent, codes...)
	}
	return 0, false
}

// IsBenign checks the error or any error in the chain, is benign
func IsBenign(err error) (reason string, benign bool) {
	e := As(err)
	if e == nil {
		return "", false
	}
	if e.benign {
		return e.benignReason, e.benign
	}
	return IsBenign(e.Unwrap())
}

// IsSilent checks the error or any error in the chain, is silent
func IsSilent(err error) bool {
	e := As(err)
	if e == nil {
		return false
	}
	if e.silent {
		return true
	}
	return IsSilent(e.Unwrap())
}

// Convert converts a regular error to a SimpleError.
// If it is already a SimpleError then it simply returns it as a SimpleError
func Convert(err error) *SimpleError {
	if e := As(err); e != nil {
		return e
	}

	// Check if any registered code conversion functions create a SimpleError{} and return it if so
	for _, f := range registry.conversions {
		if e := f(err); e != nil {
			return e
		}
	}

	// Do a minimal conversion to SimpleError{}, assuming nothing about the error
	return &SimpleError{parent: err, stackTrace: stackTrace(3)}
}

// ExtractAuxiliary extracts a superset of auxiliary data from all errors in the chain.
// Wrapper error auxiliary data take precedent over later errors.
func ExtractAuxiliary(err error) map[string]interface{} {
	if err == nil {
		return nil
	}
	aux := map[string]interface{}{}
	e := As(err)
	for e != nil {
		for k, v := range e.GetAuxiliary() {
			aux[k] = v
		}
		e = As(e.Unwrap())
	}

	return aux
}

// GetAttribute gets the first instance of the key in the error chain
func GetAttribute(err error, key interface{}) interface{} {
	if err == nil {
		return nil
	}
	e := As(err)
	for e != nil {
		for _, attr := range e.attr {
			if attr.Key == key {
				return attr.Value
			}
		}
		e = As(e.Unwrap())
	}

	return nil
}
