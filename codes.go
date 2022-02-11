package simplerr

// These are common impact error codes that are found throughout our services
const (
	// CodeUnknown is the default code for errors that are not classified
	CodeUnknown Code = iota
	// CodeAlreadyExists means an attempt to create an entity failed because one
	// already exists.
	CodeAlreadyExists
	// CodeNotFound means some requested entity (e.g., file or directory) was not found.
	CodeNotFound
	// CodeInvalidArgument indicates that the caller specified an invalid argument.
	CodeInvalidArgument
	// CodeUnauthenticated indicates the request does not have valid authentication credentials for the operation.
	CodeUnauthenticated
	// CodePermissionDenied indicates that the identity of the user is confirmed but they do not have permissions
	// to perform the request
	CodePermissionDenied
	// CodeConstraintViolated indicates that a constraint in the system has been violated.
	// Eg. a duplicate key error from a unique index
	CodeConstraintViolated
	// CodeNotSupported indicates that the request is not supported
	CodeNotSupported
	// CodeNotImplemented indicates that the request is not implemented
	CodeNotImplemented
	// CodeMissingParameter indicates that a required parameter is missing or empty
	CodeMissingParameter
	// CodeDeadlineExceeded indicates that a request exceeded it's deadline before completion
	CodeDeadlineExceeded
	// CodeCanceled indicates that the request was canceled before completion
	CodeCanceled
)

// NumberOfReservedCodes is the code number, under which, are reserved for use by this library.
const NumberOfReservedCodes = 100

var defaultErrorCodes = map[Code]string{
	CodeUnknown:            "unknown",
	CodeAlreadyExists:      "already exists",
	CodeNotFound:           "not found",
	CodeInvalidArgument:    "invalid argument",
	CodeUnauthenticated:    "unauthenticated",
	CodePermissionDenied:   "permission denied",
	CodeConstraintViolated: "constraint violated",
	CodeNotSupported:       "not supported",
	CodeMissingParameter:   "parameter is missing",
	CodeNotImplemented:     "not implemented",
	CodeDeadlineExceeded:   "deadline exceeded",
	CodeCanceled:           "canceled",
}
