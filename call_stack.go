package simplerr

import "runtime"

const (
	// maxStackFrames are the maximum number of stack frames returned with the error when using WithStackTrace
	maxStackFrames = 16
)

// Call contains information for a specific call in the call stack
type Call struct {
	Line int
	File string
	Func string
}

// stackTrace returns the call stack trace with `skip` frames skipped
func stackTrace(skip int) []Call {
	calls := make([]Call, 0, maxStackFrames)
	pcs := make([]uintptr, maxStackFrames)
	runtime.Callers(skip, pcs)
	frames := runtime.CallersFrames(pcs)
	for {
		f, ok := frames.Next()
		if !ok {
			break
		}

		calls = append(calls, Call{
			Line: f.Line,
			File: f.File,
			Func: runtime.FuncForPC(f.PC).Name(),
		})

	}
	return calls
}
