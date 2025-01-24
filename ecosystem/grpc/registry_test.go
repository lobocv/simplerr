package simplegrpc

import (
	"github.com/lobocv/simplerr"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"testing"
)

func TestGetGPRCCode(t *testing.T) {

	reg := NewRegistry()
	simplerrCode, found := reg.getGRPCCode(codes.NotFound)
	require.True(t, found)
	require.Equal(t, simplerr.CodeNotFound, simplerrCode)

	simplerrCode, found = reg.getGRPCCode(codes.Code(100000))
	require.False(t, found)
	require.Equal(t, simplerr.CodeUnknown, simplerrCode)
}
