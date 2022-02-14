package simplerr

import (
	"context"
	"errors"
)

var defaultErrorConversions = []ErrorConversion{
	ContextCanceled,
	ContextDeadlineExceeded,
}

// ContextCanceled wraps a context.Canceled as a SimpleError with the code CodeCanceled
func ContextCanceled(err error) *SimpleError {
	if errors.Is(err, context.Canceled) {
		return Wrap(err).Code(CodeCanceled)
	}
	return nil
}

// ContextDeadlineExceeded wraps a context.DeadlineExceeded as a SimpleError with the code CodeDeadlineExceeded
func ContextDeadlineExceeded(err error) *SimpleError {
	if errors.Is(err, context.DeadlineExceeded) {
		return Wrap(err).Code(CodeDeadlineExceeded)
	}
	return nil
}
