package simplegrpc

import (
	"github.com/lobocv/simplerr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// grpcError is a wrapper that exposes a SimpleError in a way that implements the gRPC status interface
// This is required because the grpc `status` library returns an error that does not implement unwrapping.
type grpcError struct {
	*simplerr.SimpleError
	code codes.Code
}

// Unwrap implement the interface required for error unwrapping
func (e *grpcError) Unwrap() error {
	return e.SimpleError
}

// GRPCStatus implements an interface that the gRPC framework uses to return the gRPC status code
func (e *grpcError) GRPCStatus() *status.Status {
	// If the status was attached as an attribute, return it
	v, _ := simplerr.GetAttribute(e.SimpleError, AttrGRPCStatus)
	st, ok := v.(*status.Status)
	if ok {
		return st
	}

	return status.New(e.code, e.Error())
}
