package simplegrpc

import (
	"github.com/lobocv/simplerr"
	"google.golang.org/grpc/codes"
)

// Registry is a registry which contains the mapping between simplerr codes and grpc error codes
type Registry struct {
	toGRPC   map[simplerr.Code]codes.Code
	fromGRPC map[codes.Code]simplerr.Code
}

// NewRegistry creates a new registry which contains the mapping to and from simplerr codes and grpc error codes
func NewRegistry() *Registry {
	return &Registry{
		toGRPC:   DefaultMapping(),
		fromGRPC: DefaultInverseMapping(),
	}
}

// SetMapping sets the mapping from simplerr.Code to gRPC code
func (r *Registry) SetMapping(m map[simplerr.Code]codes.Code) {
	r.toGRPC = m
}

// SetInverseMapping sets the mapping from gRPC code to simplerr.Code
func (r *Registry) SetInverseMapping(m map[codes.Code]simplerr.Code) {
	r.fromGRPC = m
}

// getGRPCCode gets the simplerr Code that corresponds to the GRPC code. It returns CodeUnknown if it cannot map the status.
func (r *Registry) getGRPCCode(grpcCode codes.Code) (code simplerr.Code, found bool) {
	code, ok := r.fromGRPC[grpcCode]
	if !ok {
		return simplerr.CodeUnknown, false
	}
	return code, true
}
