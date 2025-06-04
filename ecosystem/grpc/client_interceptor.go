package simplegrpc

import (
	"context"
	"github.com/lobocv/simplerr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type attr int

const (
	// AttrGRPCMethod is the simplerr attribute key for retrieving the grpc method
	AttrGRPCMethod = attr(1)
	// AttrGRPCStatus is the simplerr attribute key for retrieving the grpc Status object
	AttrGRPCStatus = attr(2)
)

// ReturnSimpleErrors returns a unary client interceptor that converts errors returned by the client to simplerr compatible
// errors. The underlying grpc status and code can still be extracted using the same status.FromError() and status.Code() methods
func ReturnSimpleErrors(registry *Registry) grpc.UnaryClientInterceptor {

	if registry == nil {
		registry = defaultRegistry
	}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		// Call the gRPC method
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			return nil
		}

		grpcCode := codes.Unknown

		serr := simplerr.Wrap(err).Attr(AttrGRPCMethod, method) // nolint: govet

		// Check if the error is a gRPC status error
		// The GRPC framework seems to always return grpc errors on the client side, even if the server does not
		// Therefore, this block should always run
		if st, ok := status.FromError(err); ok {
			_ = serr.Attr(AttrGRPCStatus, st)

			grpcCode = st.Code()
			simplerrCode, _ := registry.getGRPCCode(grpcCode)
			_ = serr.Code(simplerrCode)
		}

		return &grpcError{
			SimpleError: serr,
			code:        grpcCode,
		}

	}
}
