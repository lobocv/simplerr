package simplerr

import (
	"context"
	"errors"
)

var defaultErrorConversions = []ErrorConversion{
	ContextCanceled,
	ContextDeadlineExceeded,
}

func ContextCanceled(err error) *SimpleError {
	if errors.Is(err, context.Canceled) {
		return New(err).Code(CodeCanceled)
	}
	return nil
}

func ContextDeadlineExceeded(err error) *SimpleError {
	if errors.Is(err, context.DeadlineExceeded) {
		return New(err).Code(CodeDeadlineExceeded)
	}
	return nil
}
