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

func (s *TestSuite) TestBenign() {
	original := fmt.Errorf("test")

	s.Run("stdlib are not set as benign", func() {
		reason, benign := IsBenign(original)
		s.Empty(reason)
		s.False(benign)
	})

	s.Run("wrapped errors are not set as benign by default", func() {
		serr := Wrap(original, "wrapped")
		gotReason, isBenign := serr.GetBenignReason()
		s.Empty(gotReason)
		s.False(isBenign)
	})

	s.Run("using BenignReason to set benign", func() {
		serr := Wrap(original, "wrapped")
		_ = serr.BenignReason("good reason")
		gotReason, isBenign := serr.GetBenignReason()
		s.Equal(gotReason, "good reason")
		s.True(isBenign)
		gotReason, isBenign = IsBenign(serr)
		s.Equal(gotReason, "good reason")
		s.True(isBenign)
	})

	s.Run("using Benign to set benign", func() {
		serr := Wrap(original, "wrapped")
		_ = serr.Benign()
		gotReason, isBenign := serr.GetBenignReason()
		s.Empty(gotReason)
		s.True(isBenign)
		gotReason, isBenign = IsBenign(serr)
		s.Empty(gotReason)
		s.True(isBenign)
	})

	s.Run("check wrapping DOES NOT hide the benign", func() {
		serr := Wrap(original, "wrapped").Benign()
		wrapped := Wrap(serr, "wrapped")
		gotReason, isBenign := IsBenign(wrapped)
		s.Empty(gotReason, "")
		s.True(isBenign)
		wrappedstdlib := fmt.Errorf("stdlib wrap %w", wrapped)
		gotReason, isBenign = IsBenign(wrappedstdlib)
		s.Empty(gotReason, "")
		s.True(isBenign)
	})

	s.Run("check opaquing DOES hide the benign", func() {
		serr := Wrap(original, "wrapped").Benign()
		opaque := fmt.Errorf("stdlib wrap %s", serr)
		gotReason, isBenign := IsBenign(opaque)
		s.Empty(gotReason, "")
		s.False(isBenign)
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

	s.Run("check opaquing DOES hide the benign", func() {
		serr := Wrap(original, "wrapped").Benign()
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

func (s *TestSuite) TestAuxiliaryFields() {
	serr := Newf("something").Aux("one", 1, "two", 2.0, "three", "THREE")
	expected := map[string]interface{}{
		"one":   1,
		"two":   2.0,
		"three": "THREE",
	}
	s.Equal(expected, serr.GetAuxiliary())

	// Add more fields, but this time with one incomplete KV pair
	serr = serr.Aux("four", 4, "no_value")
	expected["four"] = 4
	s.Equal(expected, serr.GetAuxiliary())

	// Add more fields but this time with a map
	serr = serr.AuxMap(map[string]interface{}{"five": 5})
	expected["five"] = 5
	s.Equal(expected, serr.GetAuxiliary())

}
