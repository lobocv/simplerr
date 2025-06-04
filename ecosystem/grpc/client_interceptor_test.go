package simplegrpc

import (
	"context"
	"fmt"
	"github.com/lobocv/simplerr"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

var mockInvoker = func(grpcError error) grpc.UnaryInvoker {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return grpcError
	}
}

func makeMockGrpcCall(returnedError error) func() error {

	defaultInverseMapping := DefaultInverseMapping()
	defaultInverseMapping[codes.DataLoss] = simplerr.CodeResourceExhausted
	GetDefaultRegistry().SetInverseMapping(defaultInverseMapping)

	interceptor := ReturnSimpleErrors(nil)

	return func() error {
		return interceptor(context.Background(), "/ping.PingService/Ping", nil, nil, nil, mockInvoker(returnedError))
	}
}

func TestClientInterceptor(t *testing.T) {

	err := makeMockGrpcCall(status.Error(codes.NotFound, "not found"))()

	require.True(t, simplerr.HasErrorCode(err, simplerr.CodeNotFound), "simplerror code can be detected")
	require.Equal(t, codes.NotFound, status.Code(err), "grpc code can be detected with grpc status package")

	st, ok := simplerr.GetAttribute(err, AttrGRPCStatus)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.(*status.Status).Code(), "can get the grpc Status") // nolint: errcheck

	method, ok := simplerr.GetAttribute(err, AttrGRPCMethod)
	require.True(t, ok)
	require.Equal(t, "/ping.PingService/Ping", method, "can get the grpc method which errored")

	// Test the custom added mapping
	err = makeMockGrpcCall(status.Error(codes.DataLoss, "data loss"))()
	require.True(t, simplerr.HasErrorCode(err, simplerr.CodeResourceExhausted), "simplerror code can be detected")

}

// When a non grpc error is returned, the client still returns a grpc error with code Unknown
// Our interceptor should still be able to detect attributes on the error
func TestClientInterceptorNotGPRCError(t *testing.T) {

	err := makeMockGrpcCall(fmt.Errorf("some error"))()

	require.True(t, simplerr.HasErrorCode(err, simplerr.CodeUnknown), "simplerror code can be detected")
	require.Equal(t, codes.Unknown, status.Code(err), "grpc code can be detected with grpc status package")

	method, ok := simplerr.GetAttribute(err, AttrGRPCMethod)
	require.True(t, ok)
	require.Equal(t, "/ping.PingService/Ping", method, "can get the grpc method which errored")

}

func TestClientInterceptorNoError(t *testing.T) {
	err := makeMockGrpcCall(nil)()
	require.Nil(t, err)
}
