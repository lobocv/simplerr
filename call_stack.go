package simplerr

import (
	"runtime"
	"strings"
)

const (
	// maxStackFrames are the maximum number of stack frames returned with the error when using WithStackTrace
	maxStackFrames = 16
)

// Call contains information for a specific call in the call stack
type Call struct {
	Line     int
	File     string
	Func     string
	FuncName string
	Package  string
}

// stackTrace returns a more human readable form of the the stack trace from the slice of program counters
func stackTrace(pcs []uintptr) []Call {
	calls := make([]Call, 0, maxStackFrames)

	frames := runtime.CallersFrames(pcs)
	for {
		f, ok := frames.Next()
		if !ok {
			break
		}

		var pkg string
		function := f.Function

		if function != "" {
			pkg, function = splitQualifiedFunctionName(function)
		}

		calls = append(calls, Call{
			Line:     f.Line,
			File:     f.File,
			FuncName: function,
			Package:  pkg,
			Func:     runtime.FuncForPC(f.PC).Name(),
		})

	}
	return calls
}

// rawStackFrames extracts the slice of program counters associated with the stack trace and skips the first `skip`
// number of calls.
func rawStackFrames(skip int) []uintptr {
	pcs := make([]uintptr, maxStackFrames)
	runtime.Callers(skip, pcs)
	return pcs
}

// packageName returns the name of the package from the fully qualified package name
func packageName(fullName string) string {
	var pkg string

	// Look at the last segment only
	pathend := strings.LastIndex(fullName, "/")
	if pathend < 0 {
		pathend = 0
	}

	// Look for the "." that separates the package name from the function name
	// and take only the first part
	i := strings.Index(fullName[pathend:], ".")
	if i != -1 {
		pkg = fullName[:pathend+i]
	}
	return pkg
}

// splitQualifiedFunctionName splits a package path-qualified function name into
// package name and function name. Such qualified names are found in
// runtime.Frame.Function values.
func splitQualifiedFunctionName(name string) (pkg string, fun string) {
	pkg = packageName(name)
	fun = strings.TrimPrefix(name, pkg+".")
	return
}
