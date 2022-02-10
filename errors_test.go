package simplerr

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/suite"
	"testing"
)

type TestSuite struct {
	suite.Suite
}

func TestErrors(t *testing.T) {
	s := new(TestSuite)
	suite.Run(t, s)
}

func (s *TestSuite) TestErrorWrapping() {
	original := fmt.Errorf("test")
	notfound := NewNotFoundErrorW(original, "wrapped in not found")
	serviceErr := Wrap(notfound, "manually wrapped")
	wrappedErr := fmt.Errorf("stdlib wrapper: %w", serviceErr)
	opaqueErr := fmt.Errorf("opaque: %v", serviceErr)

	// Print the error chain for debugging purpose
	e := wrappedErr
	for ii := 1; ; ii++ {
		fmt.Printf("layer %d: %v\n", ii, e)
		e = errors.Unwrap(e)
		if e == nil {
			break
		}
	}

	// Test that we can detect if the error chain contains ServiceErrors
	s.Equal(serviceErr, As(serviceErr))
	s.Equal(serviceErr, As(wrappedErr))
	s.Equal(notfound, As(notfound))
	// These errors do not contain a service error, results should be nil
	s.Nil(As(opaqueErr))
	s.Nil(As(original))

	// Test that we can detect if the error chain contains NotFoundErrors
	s.False(IsNotFoundError(original))
	s.True(IsNotFoundError(notfound))
	s.True(IsNotFoundError(serviceErr))
	s.True(IsNotFoundError(wrappedErr))

}

func (s *TestSuite) TestSkippable() {
	original := fmt.Errorf("test")

	s.Run("stdlib are not set as skip", func() {
		reason, skip := IsSkippable(original)
		s.Empty(reason)
		s.False(skip)
	})

	s.Run("wrapped errors are not set as skip by default", func() {
		serr := Wrap(original, "wrapped")
		gotReason, isSkip := serr.GetSkipReason()
		s.Empty(gotReason)
		s.False(isSkip)
	})

	s.Run("using SkipReason to set skip", func() {
		serr := Wrap(original, "wrapped")
		_ = serr.SkipReason("good reason")
		gotReason, isSkip := serr.GetSkipReason()
		s.Equal(gotReason, "good reason")
		s.True(isSkip)
		gotReason, isSkip = IsSkippable(serr)
		s.Equal(gotReason, "good reason")
		s.True(isSkip)
	})

	s.Run("using Skip to set skip", func() {
		serr := Wrap(original, "wrapped")
		_ = serr.Skip()
		gotReason, isSkip := serr.GetSkipReason()
		s.Empty(gotReason)
		s.True(isSkip)
		gotReason, isSkip = IsSkippable(serr)
		s.Empty(gotReason)
		s.True(isSkip)
	})

	s.Run("check wrapping DOES NOT hide the skip", func() {
		serr := Wrap(original, "wrapped").Skip()
		wrapped := Wrap(serr, "wrapped")
		gotReason, isSkip := IsSkippable(wrapped)
		s.Empty(gotReason, "")
		s.True(isSkip)
		wrappedstdlib := fmt.Errorf("stdlib wrap %w", wrapped)
		gotReason, isSkip = IsSkippable(wrappedstdlib)
		s.Empty(gotReason, "")
		s.True(isSkip)
	})

	s.Run("check opaquing DOES hide the skip", func() {
		serr := Wrap(original, "wrapped").Skip()
		opaque := fmt.Errorf("stdlib wrap %s", serr)
		gotReason, isSkip := IsSkippable(opaque)
		s.Empty(gotReason, "")
		s.False(isSkip)
	})
}

func (s *TestSuite) TestSilent() {
	original := fmt.Errorf("test")

	s.Run("stdlib are not set as silent", func() {
		silent := IsSilent(original)
		s.False(silent)
	})

	s.Run("wrapped errors are not set as silent by default", func() {
		serr := Wrap(original, "wrapped")
		s.False(serr.GetSilent())
		s.False(IsSilent(serr))
	})

	s.Run("using Silence to set silent", func() {
		serr := Wrap(original, "wrapped").Silence()
		s.True(serr.GetSilent())
		s.True(IsSilent(serr))
	})

	s.Run("check wrapping DOES NOT hide the silence", func() {
		serr := Wrap(original, "wrapped").Silence()
		wrapped := Wrap(serr, "wrapped")
		s.True(IsSilent(wrapped))
		wrappedstdlib := fmt.Errorf("stdlib wrap %w", wrapped)
		s.True(IsSilent(wrappedstdlib))
	})

	s.Run("check opaquing DOES hide the skip", func() {
		serr := Wrap(original, "wrapped").Skip()
		opaque := fmt.Errorf("stdlib wrap %s", serr)
		s.False(IsSilent(opaque))
	})
}

// Test that any error can be convert to an HSError, and if it is already an HSError, it gets casted instead
func (s *TestSuite) TestConvert() {
	serr := Newf("original").Silence()

	s.Run("SimpleErrors are not converted, just returned", func() {
		got := Convert(serr)
		s.Equal(got, serr)
	})

	s.Run("convert non-simple error to simple error (no conversions)", func() {
		stdErr := fmt.Errorf("something")
		got := Convert(stdErr)
		s.Equal(got, &SimpleError{err: stdErr})
	})

	s.Run("convert non-simple error to simple error (with conversions)", func() {
		err := context.DeadlineExceeded
		got := Convert(err)
		s.Equal(got, New(err).Code(CodeTimedOut))
	})

}
