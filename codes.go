package simplerr

// These are common impact error codes that are found throughout our services
const (
	CodeUnknown Code = iota
	CodeAlreadyExists
	CodeNotFound
	CodeNothingUpdated
	CodeNothingDeleted
	CodeInvalidArgument
	CodeUnauthenticated
	CodeInvalidAuth
	CodeConstraintViolated
	CodeNotSupported
	CodeNotImplemented
	CodeMissingParameter
	CodeTimedOut
	CodeCanceled
)

// NumberOfReservedCodes is the code number, under which, are reserved for use by this library.
const NumberOfReservedCodes = 100

var defaultErrorCodes = map[Code]string{
	CodeUnknown:            "unknown",
	CodeAlreadyExists:      "already exists",
	CodeNotFound:           "not found",
	CodeNothingUpdated:     "nothing was updated",
	CodeNothingDeleted:     "nothing was deleted",
	CodeInvalidArgument:    "invalid argument",
	CodeUnauthenticated:    "unauthenticated",
	CodeInvalidAuth:        "invalid authentication details",
	CodeConstraintViolated: "constraint violated",
	CodeNotSupported:       "not supported",
	CodeMissingParameter:   "parameter is missing",
	CodeNotImplemented:     "not implemented",
	CodeTimedOut:           "timed out",
	CodeCanceled:           "canceled",
}
