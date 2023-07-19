package simplerr

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CustomError struct {
	error
}

type EmbeddedSimpleErr struct {
	*SimpleError
}

type TestSuite struct {
	suite.Suite
}

func (s *TestSuite) checkCall(c Call, funcName string) {
	s.True(strings.HasSuffix(c.Func, funcName))
	s.Equal(c.FuncName, funcName)
	s.Equal(c.Package, "github.com/lobocv/simplerr")
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

func (s *TestSuite) TestMutations() {
	err := New("123").Code(CodeNotFound).Message("something")

	s.Run("get code", func() {
		s.Equal(CodeNotFound, err.GetCode())
	})

	s.Run("get message", func() {
		s.Equal("something", err.Error())
	})
}

func (s *TestSuite) TestHasErrorCodes() {

	original := New("something").Code(CodeMissingParameter)
	wrapped := Wrap(original).Code(CodeNotFound)
	wrapped2 := Wrap(wrapped).Code(CodeUnknown)

	s.Run("look for first code", func() {
		s.True(HasErrorCode(wrapped, CodeNotFound))

		c, ok := HasErrorCodes(wrapped, CodeNotFound, CodeMissingParameter)
		s.True(ok)
		s.Equal(c, CodeNotFound)
	})

	s.Run("look for second code", func() {
		s.True(HasErrorCode(wrapped, CodeNotFound))

		c, ok := HasErrorCodes(wrapped, CodeMissingParameter, CodeNotFound)
		s.True(ok)
		s.Equal(c, CodeNotFound)
	})

	s.Run("look for wrapped code", func() {
		s.True(HasErrorCode(wrapped2, CodeMissingParameter))

		c, ok := HasErrorCodes(wrapped2, CodeMissingParameter)
		s.True(ok)
		s.Equal(c, CodeMissingParameter)
	})

	s.Run("look for other wrapped code", func() {
		s.True(HasErrorCode(wrapped2, CodeNotFound))

		c, ok := HasErrorCodes(wrapped2, CodeNotFound)
		s.True(ok)
		s.Equal(c, CodeNotFound)
	})

	s.Run("look for other embedded error code", func() {
		embeddedErr := EmbeddedSimpleErr{SimpleError: wrapped}

		s.True(HasErrorCode(embeddedErr, CodeNotFound))

		c, ok := HasErrorCodes(embeddedErr, CodeNotFound)
		s.True(ok)
		s.Equal(c, CodeNotFound)
	})

	s.Run("look for non existing code", func() {
		s.False(HasErrorCode(wrapped, CodePermissionDenied))

		c, ok := HasErrorCodes(wrapped2, CodePermissionDenied)
		s.False(ok)
		s.Zero(c)
	})

	s.Run("look for code in non-simple error", func() {
		err := fmt.Errorf("something")

		s.False(HasErrorCode(err, CodeNotFound))

		c, ok := HasErrorCodes(err, CodeNotFound)
		s.False(ok)
		s.Zero(c)
	})

	s.Run("look at nil error", func() {
		s.False(HasErrorCode(nil, CodeNotFound))

		c, ok := HasErrorCodes(nil, CodeNotFound)
		s.False(ok)
		s.Zero(c)
	})

	s.Run("look for no codes", func() {
		c, ok := HasErrorCodes(wrapped2)
		s.False(ok)
		s.Zero(c)
	})
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

	s.Run("check embedding DOES NOT hide the benign", func() {
		serr := EmbeddedSimpleErr{SimpleError: Wrapf(original, "wrapped").Benign()}
		_, isBenign := IsBenign(serr)
		s.True(isBenign)
		wrappedstdlib := fmt.Errorf("stdlib wrap %w", serr)
		_, isBenign = IsBenign(wrappedstdlib)
		s.True(isBenign)
		wrappedSerr := Wrapf(wrappedstdlib, "again wrapped")
		_, isBenign = IsBenign(wrappedSerr)
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

	s.Run("check embedding DOES NOT hide the silence", func() {
		serr := EmbeddedSimpleErr{SimpleError: Wrapf(original, "wrapped").Silence()}
		s.True(IsSilent(serr))
		wrappedstdlib := fmt.Errorf("stdlib wrap %w", serr)
		s.True(IsSilent(wrappedstdlib))
		wrappedSerr := Wrapf(wrappedstdlib, "again wrapped")
		s.True(IsSilent(wrappedSerr))
	})

	s.Run("check wrapping DOES NOT hide the silence", func() {
		serr := Wrapf(original, "wrapped").Silence()
		wrapped := Wrapf(serr, "wrapped")
		s.True(IsSilent(wrapped))
		wrappedstdlib := fmt.Errorf("stdlib wrap %w", wrapped)
		s.True(IsSilent(wrappedstdlib))
	})

	s.Run("check opaquing DOES hide the silence", func() {
		serr := Wrapf(original, "wrapped").Silence()
		opaque := fmt.Errorf("stdlib wrap %s", serr)
		s.False(IsSilent(opaque))
	})
}

func (s *TestSuite) TestRetriable() {
	original := fmt.Errorf("test")

	s.Run("stdlib are seen as not retriable by default", func() {
		retriable := IsRetriable(original)
		s.False(retriable)
	})

	s.Run("nil errors are not retriable", func() {
		var err error
		s.False(IsRetriable(err))
	})

	s.Run("wrapped errors are set as not retriable by default", func() {
		serr := Wrapf(original, "wrapped")
		s.False(serr.GetRetriable())
		s.False(IsRetriable(serr))
	})

	s.Run("using Retriable to set retriable", func() {
		serr := Wrapf(original, "wrapped").Retriable()
		s.True(serr.GetRetriable())
		s.True(IsRetriable(serr))
	})

	s.Run("check embedding DOES NOT hide the retriable status", func() {
		serr := EmbeddedSimpleErr{SimpleError: Wrapf(original, "wrapped")}
		s.False(IsRetriable(serr))
		_ = serr.Retriable()
		s.True(IsRetriable(serr))
		wrappedstdlib := fmt.Errorf("stdlib wrap %w", serr)
		s.True(IsRetriable(wrappedstdlib))
		wrappedSerr := Wrapf(wrappedstdlib, "again wrapped")
		s.True(IsRetriable(wrappedSerr))
	})

	s.Run("check wrapping DOES NOT hide the retriable status", func() {
		serr := Wrapf(original, "wrapped").Retriable()
		wrapped := Wrapf(serr, "wrapped")
		s.True(IsRetriable(wrapped))
		wrappedstdlib := fmt.Errorf("stdlib wrap %w", wrapped)
		s.True(IsRetriable(wrappedstdlib))
	})

	s.Run("check opaquing DOES hide the retriable status", func() {
		serr := Wrapf(original, "wrapped").Retriable()
		opaque := fmt.Errorf("stdlib wrap %s", serr)
		s.False(IsRetriable(opaque))
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
	serr = New("something")
	expected = map[string]interface{}{"one": 1, "two": 2}
	serr = serr.AuxMap(expected)
	serr = serr.AuxMap(expected)

	s.Equal(expected, serr.GetAuxiliary())

	s.Run("extract all aux from wrapped and embedded errors", func() {
		wrapped := Wrap(serr).Aux("name", "Calvin")
		expected["name"] = "Calvin"

		embeddedErr := EmbeddedSimpleErr{SimpleError: wrapped}
		_ = embeddedErr.Aux("country", "Canada")
		expected["country"] = "Canada"

		wrapped2 := Wrap(embeddedErr).Aux("province", "Ontario")
		expected["province"] = "Ontario"

		s.Equal(expected, ExtractAuxiliary(wrapped2))
	})

	s.Run("extract all aux from nil error", func() {
		s.Nil(ExtractAuxiliary(nil))
	})

}

func (s *TestSuite) TestAttributes() {

	s.Run("single attribute", func() {
		serr := New("something").Attr(1, "one")
		v, exists := GetAttribute(serr, 1)
		s.True(exists)
		s.Equal("one", v)
	})

	s.Run("non-existing attribute", func() {
		serr := New("something")
		v, exists := GetAttribute(serr, "does-not-exist")
		s.False(exists)
		s.Nil(v)
	})

	s.Run("nil error", func() {
		v, exists := GetAttribute(nil, "does-not-exist")
		s.False(exists)
		s.Nil(v)
	})

	s.Run("single attribute on a wrapped error", func() {
		serr := Wrap(New("something").Attr(1, "one"))
		v, exists := GetAttribute(serr, 1)
		s.True(exists)
		s.Equal("one", v)
	})

	s.Run("single attribute on an embedded error", func() {
		embeddedErr := EmbeddedSimpleErr{SimpleError: Wrap(New("something").Attr(1, "one"))}
		_ = embeddedErr.Attr(2, "two")

		v, exists := GetAttribute(embeddedErr, 1)
		s.True(exists)
		s.Equal("one", v)

		v, exists = GetAttribute(embeddedErr, 2)
		s.True(exists)
		s.Equal("two", v)

		v, exists = GetAttribute(embeddedErr, "does not exist")
		s.False(exists)
		s.Nil(v)
	})

	s.Run("duplicate attribute with same key type and value", func() {
		serr := New("something").
			Attr(1, "one").
			Attr(1, "two")
		v, exists := GetAttribute(serr, 1)
		s.True(exists)
		s.Equal("one", v, "first attribute does not get overwritten")
	})

	s.Run("custom key-type attribute does not overlap with same underlying type value", func() {
		// Much like the context package, using a custom type prevents naming collisions
		type customKey int
		const attrKey = customKey(1)

		serr := New("something").
			Attr(attrKey, "one").
			Attr(1, "two")

		v, exists := GetAttribute(serr, attrKey)
		s.True(exists)
		s.Equal("one", v)

		v, exists = GetAttribute(serr, 1)
		s.True(exists)
		s.Equal("two", v)
	})

}

func (s *TestSuite) TestErrorCodeDescriptions() {
	serr := New("something")
	s.Equal("unknown", serr.GetDescription())
	_ = serr.Code(CodeNotSupported)
	s.Equal("not supported", serr.GetDescription())
}

func (s *TestSuite) TestCustomRegistry() {
	r := NewRegistry()
	const CodeCustom = 100
	r.RegisterErrorCode(CodeCustom, "custom")
	// Because the registry is a global, to prevent mucking with other tests, set it back afterwards
	defaultRegistry := GetRegistry()
	SetRegistry(r)
	defer SetRegistry(defaultRegistry)

	s.Run("use custom convert ", func() {
		serr := New("convert this")
		s.Equal("", serr.GetDescription(), "custom registry doesnt have NotFound code defined")
		serr = serr.Code(CodeCustom)
		s.Equal("custom", serr.GetDescription())
	})

	s.Run("get error codes", func() {
		codes := r.ErrorCodes()
		s.Equal(map[Code]string{CodeCustom: "custom"}, codes)
	})

	s.Run("cannot register reserved code range ", func() {
		s.Panics(func() {
			r.RegisterErrorCode(NumberOfReservedCodes-1, "something else")
		})
	})

	s.Run("cannot register reserved codes already in use ", func() {
		s.Panics(func() {
			r.RegisterErrorCode(CodeNotFound, "something else")
		})
	})

	s.Run("cannot register custom code already in use", func() {
		s.Panics(func() {
			r.RegisterErrorCode(CodeCustom, "custom")
		})
	})

}

func (s *TestSuite) TestErrorFormatting() {
	original := fmt.Errorf("original")
	serr1 := Wrapf(original, "wrapper %d", 1)
	serr2 := Wrapf(serr1, "wrapper %d", 2)

	s.Equal(serr1.GetMessage(), "wrapper 1")
	s.Equal(serr2.GetMessage(), "wrapper 2")

	s.Equal("wrapper 1: original", serr1.Error())
	s.Equal("wrapper 2: wrapper 1: original", serr2.Error())

	serr3 := New("something")
	s.Equal("something", serr3.Error())

	// Change the error formatting style
	Formatter = func(e *SimpleError) string {
		parent := e.Unwrap()
		if parent == nil {
			return e.GetMessage()
		}
		return strings.Join([]string{e.GetMessage(), parent.Error()}, "\n")
	}
	s.Equal("wrapper 1\noriginal", serr1.Error())
	s.Equal("wrapper 2\nwrapper 1\noriginal", serr2.Error())
	Formatter = DefaultFormatter
}

func (s *TestSuite) TestStackTrace() {

	e := First()
	stack := e.StackTrace()
	s.checkCall(stack[0], "First")

	e = e.Unwrap().(*SimpleError) // nolint: errcheck
	stack = e.StackTrace()
	s.checkCall(stack[0], "Second")
	s.checkCall(stack[1], "First")

	e = e.Unwrap().(*SimpleError) // nolint: errcheck
	stack = e.StackTrace()
	s.checkCall(stack[0], "Third")
	s.checkCall(stack[1], "Second")
	s.checkCall(stack[2], "First")

	e = e.Unwrap().(*SimpleError) // nolint: errcheck
	stack = e.StackTrace()
	s.checkCall(stack[0], "Fourth")
	s.checkCall(stack[1], "Third")
	s.checkCall(stack[2], "Second")
	s.checkCall(stack[3], "First")

	// There's not really a great way to test the raw pointers. As long as this isn't an empty
	// list, we should be fine because StackTrace() is already tested, and it uses StackFrames()
	s.NotEmpty(e.StackFrames())
}

func Fourth() *SimpleError {
	return New("something")
}

func Third() *SimpleError {
	e := Fourth()
	return Wrapf(e, "third wrapper")
}
func Second() *SimpleError {
	e := Third()
	return Wrapf(e, "second wrapper")
}
func First() *SimpleError {
	e := Second()
	return Wrapf(e, "first wrapper")
}
