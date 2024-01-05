package simplegrpc

import (
	"github.com/lobocv/simplerr"
	"google.golang.org/grpc/codes"
)

var (
	// defaultRegistry is a global registry used by default.
	defaultRegistry = NewRegistry()
)

func GetDefaultRegistry() *Registry {
	return defaultRegistry
}

// DefaultMapping returns the default mapping of SimpleError codes to gRPC error codes
func DefaultMapping() map[simplerr.Code]codes.Code {
	m := map[simplerr.Code]codes.Code{
		simplerr.CodeUnknown:           codes.Unknown,
		simplerr.CodeAlreadyExists:     codes.AlreadyExists,
		simplerr.CodeNotFound:          codes.NotFound,
		simplerr.CodeDeadlineExceeded:  codes.DeadlineExceeded,
		simplerr.CodeCanceled:          codes.Canceled,
		simplerr.CodeUnauthenticated:   codes.Unauthenticated,
		simplerr.CodePermissionDenied:  codes.PermissionDenied,
		simplerr.CodeNotImplemented:    codes.Unimplemented,
		simplerr.CodeInvalidArgument:   codes.InvalidArgument,
		simplerr.CodeResourceExhausted: codes.ResourceExhausted,
		simplerr.CodeUnavailable:       codes.Unavailable,
	}

	return m
}

// DefaultInverseMapping returns the default inverse mapping of gRPC error codes to SimpleError codes
func DefaultInverseMapping() map[codes.Code]simplerr.Code {
	m := map[codes.Code]simplerr.Code{
		codes.Unknown:           simplerr.CodeUnknown,
		codes.AlreadyExists:     simplerr.CodeAlreadyExists,
		codes.NotFound:          simplerr.CodeNotFound,
		codes.DeadlineExceeded:  simplerr.CodeDeadlineExceeded,
		codes.Canceled:          simplerr.CodeCanceled,
		codes.Unauthenticated:   simplerr.CodeUnauthenticated,
		codes.PermissionDenied:  simplerr.CodePermissionDenied,
		codes.Unimplemented:     simplerr.CodeNotImplemented,
		codes.InvalidArgument:   simplerr.CodeInvalidArgument,
		codes.ResourceExhausted: simplerr.CodeResourceExhausted,
		codes.Unavailable:       simplerr.CodeUnavailable,
	}

	return m
}
