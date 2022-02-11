package interceptors

import (
	"context"
	"fmt"
	"github.com/lobocv/simplerr"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestTranslateErrorCode(t *testing.T) {

	testCases := []struct {
		err      error
		expected codes.Code
	}{
		{fmt.Errorf("something"), codes.Unknown},
		{simplerr.Newf("something").Code(simplerr.CodeInvalidAuth), codes.PermissionDenied},
		{fmt.Errorf("wrapped: %w", simplerr.Newf("something").Code(simplerr.CodeUnauthenticated)), codes.Unauthenticated},
		{fmt.Errorf("opaque: %s", simplerr.Newf("something").Code(simplerr.CodeUnauthenticated)), codes.Unknown},
		{nil, codes.OK},
	}

	interceptor := TranslateErrorCode(DefaultMapping())
	for _, tc := range testCases {
		_, gotErr := interceptor(context.Background(), nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			return 1, tc.err
		})

		grpcStatusCode := status.Code(gotErr)
		require.Equal(t, tc.expected, grpcStatusCode)
	}

}
