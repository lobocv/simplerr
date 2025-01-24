package simplegrpc

import (
	"context"
	"fmt"
	"github.com/lobocv/simplerr"
	"github.com/lobocv/simplerr/ecosystem/grpc/internal/ping"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"net"
	"testing"
)

type PingService struct {
	err error
}

func (s *PingService) Ping(_ context.Context, _ *ping.PingRequest) (*ping.PingResponse, error) {
	// Your implementation of the Ping method goes here
	fmt.Println("Received Ping request")
	if s.err != nil {
		return nil, s.err
	}
	return &ping.PingResponse{}, nil
}

func setupServerAndClient(port int) (*PingService, ping.PingServiceClient) {

	server := grpc.NewServer()
	service := &PingService{err: status.Error(codes.NotFound, "test error")}
	ping.RegisterPingServiceServer(server, service)

	// Create a listener on TCP port 50051
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(fmt.Sprintf("Error creating listener: %v", err))
	}

	go func() {
		if err = server.Serve(listener); err != nil {
			panic(fmt.Sprintf("Error serving: %v", err))
		}
	}()

	defaultInverseMapping := DefaultInverseMapping()
	defaultInverseMapping[codes.DataLoss] = simplerr.CodeResourceExhausted
	GetDefaultRegistry().SetInverseMapping(defaultInverseMapping)

	interceptor := ReturnSimpleErrors(nil)

	conn, err := grpc.NewClient(fmt.Sprintf(":%d", port),
		grpc.WithUnaryInterceptor(interceptor),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}
	client := ping.NewPingServiceClient(conn)

	return service, client
}

func TestClientInterceptor(t *testing.T) {

	server, client := setupServerAndClient(50051)
	_, err := client.Ping(context.Background(), &ping.PingRequest{})

	require.True(t, simplerr.HasErrorCode(err, simplerr.CodeNotFound), "simplerror code can be detected")
	require.Equal(t, codes.NotFound, status.Code(err), "grpc code can be detected with grpc status package")

	st, ok := simplerr.GetAttribute(err, AttrGRPCStatus)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.(*status.Status).Code(), "can get the grpc Status") // nolint: errcheck

	method, ok := simplerr.GetAttribute(err, AttrGRPCMethod)
	require.True(t, ok)
	require.Equal(t, "/ping.PingService/Ping", method, "can get the grpc method which errored")

	// Test the custom added mapping
	server.err = status.Error(codes.DataLoss, "test error")
	_, err = client.Ping(context.Background(), &ping.PingRequest{})
	require.True(t, simplerr.HasErrorCode(err, simplerr.CodeResourceExhausted), "simplerror code can be detected")

}

// When a non grpc error is returned, the client still returns a grpc error with code Unknown
// Our interceptor should still be able to detect attributes on the error
func TestClientInterceptorNotGPRCError(t *testing.T) {

	server, client := setupServerAndClient(50052)
	server.err = fmt.Errorf("not a grpc error")

	_, err := client.Ping(context.Background(), &ping.PingRequest{})

	require.True(t, simplerr.HasErrorCode(err, simplerr.CodeUnknown), "simplerror code can be detected")
	require.Equal(t, codes.Unknown, status.Code(err), "grpc code can be detected with grpc status package")

	st, ok := simplerr.GetAttribute(err, AttrGRPCStatus)
	require.True(t, ok)
	require.Equal(t, codes.Unknown, st.(*status.Status).Code(), "can get the grpc Status") // nolint: errcheck

	method, ok := simplerr.GetAttribute(err, AttrGRPCMethod)
	require.True(t, ok)
	require.Equal(t, "/ping.PingService/Ping", method, "can get the grpc method which errored")

}

func TestClientInterceptorNoError(t *testing.T) {
	server, client := setupServerAndClient(50053)
	server.err = nil

	_, err := client.Ping(context.Background(), &ping.PingRequest{})
	require.Nil(t, err)
}
