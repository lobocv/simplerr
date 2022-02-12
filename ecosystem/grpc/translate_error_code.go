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
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		r, err := handler(ctx, req)
		// If no error, return early
		if err == nil {
			return r, nil
		}

		// Check the error to see if it's a SimpleError, then translate to the gRPC code
		if e := simplerr.As(err); e != nil {

			grpcCode, ok := toGRPC[e.GetCode()]
			if !ok {
				grpcCode = codes.Unknown
			}

			return r, status.Error(grpcCode, e.Error())
		}

		return r, err
	}
}
