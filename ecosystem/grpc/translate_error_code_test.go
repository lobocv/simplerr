package simplegrpc

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
		{simplerr.New("something").Code(simplerr.CodePermissionDenied), codes.PermissionDenied},
		{simplerr.New("something").Code(simplerr.CodeMalformedRequest), codes.InvalidArgument},
		{simplerr.New("something").Code(simplerr.CodeMissingParameter), codes.Unknown},
		{fmt.Errorf("wrapped: %w", simplerr.New("something").Code(simplerr.CodeUnauthenticated)), codes.Unauthenticated},
		{fmt.Errorf("opaque: %s", simplerr.New("something").Code(simplerr.CodeUnauthenticated)), codes.Unknown},
		{simplerr.Wrap(simplerr.New("something").Code(simplerr.CodePermissionDenied)), codes.PermissionDenied},
		{nil, codes.OK},
	}

	// Alter the default mapping
	m := DefaultMapping()
	m[simplerr.CodeMalformedRequest] = codes.InvalidArgument

	interceptor := TranslateErrorCode(m)
	for _, tc := range testCases {
		_, gotErr := interceptor(context.Background(), nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			return 1, tc.err
		})

		grpcStatusCode := status.Code(gotErr)
		require.Equal(t, tc.expected, grpcStatusCode)

		// Check that the translated error can still be detected as a SimpleError
		expectSimplerr := simplerr.As(tc.err) != nil
		if expectSimplerr {
			gotSimplerr := simplerr.As(gotErr) != nil
			require.True(t, gotSimplerr)
		}
	}

}
