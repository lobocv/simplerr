package simplegrpc

import (
	"context"
	"github.com/lobocv/simplerr"
	"google.golang.org/grpc"
)

// TranslateErrorCode inspects the error to see if it is a SimpleError. If it is, it attempts to translate the
// SimpleError code to the corresponding grpc error code.
// If no translation exists it returns a grpc error with Unknown error code.
func TranslateErrorCode(registry *Registry) grpc.UnaryServerInterceptor {

	if registry == nil {
		registry = defaultRegistry
	}

	// Get a list of simplerr codes to search for in the error chain
	var simplerrCodes []simplerr.Code
	for c := range registry.toGRPC {
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
			grpcCode := registry.toGRPC[code]
			return r, &grpcError{
				SimpleError: e,
				code:        grpcCode,
			}
		}

		return r, err
	}
}
