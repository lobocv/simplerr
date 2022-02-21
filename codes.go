package simplerr

// Code is an error code that indicates the category of error
type Code int

// These are common impact error codes that are found throughout our services
const (
	// CodeUnknown is the default code for errors that are not classified
	CodeUnknown Code = 0
	// CodeAlreadyExists means an attempt to create an entity failed because one
	// already exists.
	CodeAlreadyExists Code = 1
	// CodeNotFound means some requested entity (e.g., file or directory) was not found.
	CodeNotFound Code = 2
	// CodeInvalidArgument indicates that the caller specified an invalid argument.
	CodeInvalidArgument Code = 3
	// CodeMalformedRequest indicates the syntax of the request cannot be interpreted (eg JSON decoding error)
	CodeMalformedRequest Code = 4
	// CodeUnauthenticated indicates the request does not have valid authentication credentials for the operation.
	CodeUnauthenticated Code = 5
	// CodePermissionDenied indicates that the identity of the user is confirmed but they do not have permissions
	// to perform the request
	CodePermissionDenied Code = 6
	// CodeConstraintViolated indicates that a constraint in the system has been violated.
	// Eg. a duplicate key error from a unique index
	CodeConstraintViolated Code = 7
	// CodeNotSupported indicates that the request is not supported
	CodeNotSupported Code = 8
	// CodeNotImplemented indicates that the request is not implemented
	CodeNotImplemented Code = 9
	// CodeMissingParameter indicates that a required parameter is missing or empty
	CodeMissingParameter Code = 10
	// CodeDeadlineExceeded indicates that a request exceeded it's deadline before completion
	CodeDeadlineExceeded Code = 11
	// CodeCanceled indicates that the request was canceled before completion
	CodeCanceled Code = 12
	// CodeResourceExhausted indicates that some limited resource (eg rate limit or disk space) has been reached
	CodeResourceExhausted Code = 13
	// CodeUnavailable indicates that the server itself is unavailable for processing requests.
	CodeUnavailable Code = 14
)

// NumberOfReservedCodes is the code number, under which, are reserved for use by this library.
const NumberOfReservedCodes = 100

var defaultErrorCodes = map[Code]string{
	CodeUnknown:            "unknown",
	CodeAlreadyExists:      "already exists",
	CodeNotFound:           "not found",
	CodeInvalidArgument:    "invalid argument",
	CodeMalformedRequest:   "malformed request",
	CodeUnauthenticated:    "unauthenticated",
	CodePermissionDenied:   "permission denied",
	CodeConstraintViolated: "constraint violated",
	CodeNotSupported:       "not supported",
	CodeMissingParameter:   "parameter is missing",
	CodeNotImplemented:     "not implemented",
	CodeDeadlineExceeded:   "deadline exceeded",
	CodeCanceled:           "canceled",
	CodeResourceExhausted:  "resource exhausted",
	CodeUnavailable:        "unavailable",
}
