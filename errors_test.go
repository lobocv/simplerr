package simplerr

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

type CustomError struct {
	error
}

type TestSuite struct {
	suite.Suite
}

func TestErrors(t *testing.T) {
	s := new(TestSuite)
	suite.Run(t, s)
}

func (s *TestSuite) TestErrorWrapping() {
	original := &CustomError{error: fmt.Errorf("test")}
	notfound := Wrapf(original, "wrapped in not found").Code(CodeNotFound)
	wrappedErr := Wrapf(notfound, "manually wrapped")
	stdlibwrappedErr := fmt.Errorf("stdlib wrapper: %w", wrappedErr)
	opaqueErr := New("opaque: %v", wrappedErr)
	stdlibOpaqueErr := fmt.Errorf("opaque: %v", wrappedErr)

	// Print the error chain for debugging purpose
	e := stdlibwrappedErr
	for ii := 1; ; ii++ {
		fmt.Printf("layer %d: %v\n", ii, e)
		e = errors.Unwrap(e)
		if e == nil {
			break
		}
	}

	// Test that we can detect if the error chain contains SimpleErrors
	s.Equal(wrappedErr, As(wrappedErr))
	s.Equal(wrappedErr, As(stdlibwrappedErr))
	s.Equal(notfound, As(notfound))
	// These errors do not contain a service error, results should be nil
	s.Nil(As(stdlibOpaqueErr))
	s.Nil(As(original))

	// Test using errors.Is works as expected
	s.True(errors.Is(notfound, original))
	s.True(errors.Is(wrappedErr, original))
	s.True(errors.Is(wrappedErr, notfound))
	s.True(errors.Is(stdlibwrappedErr, notfound))
	s.True(errors.Is(stdlibwrappedErr, original))
	s.False(errors.Is(wrappedErr, fmt.Errorf("something not matching")))

	// Test that errors.As works as expected
	ce := &CustomError{}
	s.True(errors.As(wrappedErr, &ce))
	s.False(errors.As(opaqueErr, &ce))

	// Test that we can detect if the error chain contains CodeNotFound
	s.False(HasErrorCode(original, CodeNotFound))
	s.True(HasErrorCode(notfound, CodeNotFound))
	s.True(HasErrorCode(wrappedErr, CodeNotFound))
	s.True(HasErrorCode(stdlibwrappedErr, CodeNotFound))

}

func (s *TestSuite) TestBenign() {
	original := fmt.Errorf("test")

	s.Run("stdlib are not set as benign", func() {
		reason, benign := IsBenign(original)
		s.Empty(reason)
		s.False(benign)
	})

	s.Run("wrapped errors are not set as benign by default", func() {
		serr := Wrapf(original, "wrapped")
		gotReason, isBenign := serr.GetBenignReason()
		s.Empty(gotReason)
		s.False(isBenign)
	})

	s.Run("using BenignReason to set benign", func() {
		serr := Wrapf(original, "wrapped")
		_ = serr.BenignReason("good reason")
		gotReason, isBenign := serr.GetBenignReason()
		s.Equal(gotReason, "good reason")
		s.True(isBenign)
		gotReason, isBenign = IsBenign(serr)
		s.Equal(gotReason, "good reason")
		s.True(isBenign)
	})

	s.Run("using Benign to set benign", func() {
		serr := Wrapf(original, "wrapped")
		_ = serr.Benign()
		gotReason, isBenign := serr.GetBenignReason()
		s.Empty(gotReason)
		s.True(isBenign)
		gotReason, isBenign = IsBenign(serr)
		s.Empty(gotReason)
		s.True(isBenign)
	})

	s.Run("check wrapping DOES NOT hide the benign", func() {
		serr := Wrapf(original, "wrapped").Benign()
		wrapped := Wrapf(serr, "wrapped")
		gotReason, isBenign := IsBenign(wrapped)
		s.Empty(gotReason, "")
		s.True(isBenign)
		wrappedstdlib := fmt.Errorf("stdlib wrap %w", wrapped)
		gotReason, isBenign = IsBenign(wrappedstdlib)
		s.Empty(gotReason, "")
		s.True(isBenign)
	})

	s.Run("check opaquing DOES hide the benign", func() {
		serr := Wrapf(original, "wrapped").Benign()
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
		serr := Wrapf(original, "wrapped")
		s.False(serr.GetSilent())
		s.False(IsSilent(serr))
	})

	s.Run("using Silence to set silent", func() {
		serr := Wrapf(original, "wrapped").Silence()
		s.True(serr.GetSilent())
		s.True(IsSilent(serr))
	})

	s.Run("check wrapping DOES NOT hide the silence", func() {
		serr := Wrapf(original, "wrapped").Silence()
		wrapped := Wrapf(serr, "wrapped")
		s.True(IsSilent(wrapped))
		wrappedstdlib := fmt.Errorf("stdlib wrap %w", wrapped)
		s.True(IsSilent(wrappedstdlib))
	})

	s.Run("check opaquing DOES hide the benign", func() {
		serr := Wrapf(original, "wrapped").Benign()
		opaque := fmt.Errorf("stdlib wrap %s", serr)
		s.False(IsSilent(opaque))
	})
}

// Test that any error can be convert to an HSError, and if it is already an HSError, it gets casted instead
func (s *TestSuite) TestConvert() {
	serr := New("original").Silence()

	s.Run("SimpleErrors are not converted, just returned", func() {
		got := Convert(serr)
		s.Equal(got, serr)
	})

	s.Run("convert non-simple error to simple error (no conversions)", func() {
		stdErr := fmt.Errorf("something")
		got := Convert(stdErr)
		s.Equal(got, &SimpleError{parent: stdErr})
	})

	s.Run("convert non-simple error to simple error (with conversions)", func() {
		err := context.DeadlineExceeded
		got := Convert(err)
		s.Equal(got, Wrap(err).Code(CodeDeadlineExceeded))
	})

}

func (s *TestSuite) TestAuxiliaryFields() {
	serr := New("something").Aux("one", 1, "two", 2.0, "three", "THREE")
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

func (s *TestSuite) TestErrorCodeDescriptions() {
	serr := New("something")
	s.Equal("unknown", serr.Description())
	_ = serr.Code(CodeNotSupported)
	s.Equal("not supported", serr.Description())
}

func (s *TestSuite) TestCustomRegistry() {
	r := NewRegistry()
	const CodeCustom = 100
	r.RegisterErrorCode(CodeCustom, "custom")
	// Because the registry is a global, to prevent mucking with other tests, set it back afterwards
	defaultRegistry := registry
	SetRegistry(r)
	defer SetRegistry(defaultRegistry)

	serr := New("something").Code(CodeNotFound)
	s.Equal("", serr.Description(), "custom registry doesnt have NotFound code defined")

	serr = serr.Code(CodeCustom)
	s.Equal("custom", serr.Description())
}

func (s *TestSuite) TestErrorFormatting() {
	original := fmt.Errorf("original")
	serr1 := Wrapf(original, "wrapper %d", 1)
	serr2 := Wrapf(serr1, "wrapper %d", 2)
	s.Equal("wrapper 1: original", serr1.Error())
	s.Equal("wrapper 2: wrapper 1: original", serr2.Error())

	// Change the error formatting style
	Formatter = func(e *SimpleError) string {
		return strings.Join([]string{e.msg, e.parent.Error()}, "\n")
	}
	s.Equal("wrapper 1\noriginal", serr1.Error())
	s.Equal("wrapper 2\nwrapper 1\noriginal", serr2.Error())

}

func (s *TestSuite) TestStackTrace() {

	checkCall := func(c Call, funcName string) {
		s.True(strings.HasSuffix(c.Func, funcName))
	}

	e := First()
	stack := e.StackTrace()
	checkCall(stack[0], "First")

	e = e.Unwrap().(*SimpleError)
	stack = e.StackTrace()
	checkCall(stack[0], "Second")
	checkCall(stack[1], "First")

	e = e.Unwrap().(*SimpleError)
	stack = e.StackTrace()
	checkCall(stack[0], "Third")
	checkCall(stack[1], "Second")
	checkCall(stack[2], "First")

	e = e.Unwrap().(*SimpleError)
	stack = e.StackTrace()
	checkCall(stack[0], "Fourth")
	checkCall(stack[1], "Third")
	checkCall(stack[2], "Second")
	checkCall(stack[3], "First")
}

func Fourth() *SimpleError {
	return New("something").WithStackTrace()
}

func Third() *SimpleError {
	e := Fourth()
	return Wrapf(e, "third wrapper").WithStackTrace()
}
func Second() *SimpleError {
	e := Third()
	return Wrapf(e, "second wrapper").WithStackTrace()
}
func First() *SimpleError {
	e := Second()
	return Wrapf(e, "first wrapper").WithStackTrace()
}
