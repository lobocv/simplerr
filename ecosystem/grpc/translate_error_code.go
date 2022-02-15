package simplegrpc

import (
	"context"
	"github.com/lobocv/simplerr"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
	}

	return m
}

// TranslateErrorCode inspects the error to see if it is a SimpleError. If it is, it attempts to translate the
// SimpleError code to the corresponding grpc error code.
// If no translation exists it returns a grpc error with Unknown error code.
func TranslateErrorCode(toGRPC map[simplerr.Code]codes.Code) grpc.UnaryServerInterceptor {

	// Get a list of simplerr codes to search for in the error chain
	var simplerrCodes []simplerr.Code
	for c := range toGRPC {
		// Ignore CodeUnknown because it is the default code
		if c == simplerr.CodeUnknown {
			continue
		}
		simplerrCodes = append(simplerrCodes, c)
	}

	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		r, err := handler(ctx, req)
		// If no error, return early
		if err == nil {
			return r, nil
		}

		// Check the error to see if it's a SimpleError, then translate to the gRPC code
		if e := simplerr.As(err); e != nil {

			// Check if the error has any of the codes in it's chain
			code, ok := simplerr.HasErrorCodes(e, simplerrCodes...)
			if !ok {
				return r, err
			}

			// Get the gRPC code, this lookup should never fail
			grpcCode := toGRPC[code]
			return r, status.Error(grpcCode, e.Error())
		}

		return r, err
	}
}
