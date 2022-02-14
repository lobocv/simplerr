package simplerr

import "fmt"

// Formatter is the error string formatting function.
var Formatter = DefaultFormatter

// DefaultFormatter is the default error string formatting.
func DefaultFormatter(e *SimpleError) string {
	if parent := e.Unwrap(); parent != nil {
		return fmt.Sprintf("%s: %s", e.GetMessage(), parent.Error())
	}
	return e.msg
}
