package simplehttp

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/lobocv/simplerr"
)

func timestamp() string {
	return fmt.Sprintf("%d", time.Now().UnixMicro())
}

func parseTimestamp(s string) time.Time {
	v, err := strconv.Atoi(s)
	if err != nil {
		panic("somethings wrong with the time formatting")
	}

	return time.UnixMicro(int64(v))
}

// StandardHTTPMiddleware is a middleware written for standard library http.Handlers
func StandardHTTPMiddleware(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		r.Header.Add("pre-call-middleware", timestamp())
		h.ServeHTTP(w, r)
		w.Header().Add("post-call-middleware", timestamp())
	}

	return http.HandlerFunc(fn)
}

// PreCallMiddleware injects a timestamp in the request header before the handler is called
func PreCallMiddleware(h Handler) Handler {
	fn := func(writer http.ResponseWriter, request *http.Request) error {
		request.Header.Add("pre-call-middleware", timestamp())
		return h.ServeHTTP(writer, request)
	}
	return HandlerFunc(fn)
}

// PostCallMiddleware injects a timestamp in the response header after the handler is called
func PostCallMiddleware(h Handler) Handler {
	fn := func(writer http.ResponseWriter, request *http.Request) error {
		err := h.ServeHTTP(writer, request)
		writer.Header().Add("post-call-middleware", timestamp())
		return err
	}
	return HandlerFunc(fn)
}

// ErrorCausingMiddleware returns an error either before (pre) or after the handler is called
func ErrorCausingMiddleware(pre bool) Middleware {
	return func(h Handler) Handler {
		fn := func(writer http.ResponseWriter, request *http.Request) error {

			if pre {
				return simplerr.New("pre error in middleware").Code(simplerr.CodeInvalidArgument)
			}
			err := h.ServeHTTP(writer, request)
			if err != nil {
				return err
			}

			return simplerr.New("post error in middleware").Code(simplerr.CodeInvalidArgument)
		}

		return HandlerFunc(fn)
	}

}

func ensureHeaderOrder(t *testing.T, resp *httptest.ResponseRecorder, req *http.Request) {
	preCallTs := parseTimestamp(req.Header.Get("pre-call-middleware"))
	callTs := parseTimestamp(req.Header.Get("call"))
	postCallTs := parseTimestamp(resp.Header().Get("post-call-middleware"))

	// sort the headers by time, they should be in ascending order
	got := []time.Time{preCallTs, callTs, postCallTs}
	sort.Slice(got, func(i, j int) bool {
		return got[i].Before(got[j])
	})

	expect := []time.Time{preCallTs, callTs, postCallTs}
	require.ElementsMatch(t, expect, got, "middleware are called out of order")
}

func TestPreAndPostCallMiddleware(t *testing.T) {

	//nolint:unparam
	ep := func(writer http.ResponseWriter, request *http.Request) error {
		request.Header.Add("call", timestamp())

		n, err := writer.Write([]byte("done"))
		require.NoError(t, err)
		require.Greater(t, n, 0)

		return nil
	}

	req, err := http.NewRequest("GET", "url", nil)
	require.NoError(t, err)

	t.Run("with http adapter", func(t *testing.T) {
		ep := ApplyMiddleware(ep, PreCallMiddleware, PostCallMiddleware)
		rec := httptest.NewRecorder()
		NewHandlerFuncAdapter(ep)(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		ensureHeaderOrder(t, rec, req)
	})

	t.Run("without http adapter", func(t *testing.T) {
		ep := ApplyMiddleware(ep, PreCallMiddleware, PostCallMiddleware)
		rec := httptest.NewRecorder()
		err := ep(rec, req)
		require.NoError(t, err)
		ensureHeaderOrder(t, rec, req)
	})

	t.Run("error causing pre middleware", func(t *testing.T) {
		ep := ApplyMiddleware(ep, ErrorCausingMiddleware(true))
		rec := httptest.NewRecorder()
		err := ep(rec, req)
		require.EqualError(t, err, "pre error in middleware")
		require.True(t, simplerr.HasErrorCode(err, simplerr.CodeInvalidArgument))
		require.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
		// There should be no response written because the endpoint should not have been called
		_, err = rec.Body.ReadByte()
		require.True(t, err == io.EOF)
	})

	t.Run("error causing post middleware with written response", func(t *testing.T) {
		ep := ApplyMiddleware(ep, ErrorCausingMiddleware(false))
		rec := httptest.NewRecorder()
		err := ep(rec, req)
		require.EqualError(t, err, "post error in middleware")
		// While the error does have the correct simplerr code, the response status code can only be written
		// once. If the handler writes a body or sets the status, then the code will be set and the error
		// handler cannot overwrite it. This is a known limitation that may be resolved in the future if requested.
		require.True(t, simplerr.HasErrorCode(err, simplerr.CodeInvalidArgument))

		// The endpoint writes a response before the error is raised so the status code is already set to OK
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		// There should be a response written because the endpoint should have been called
		data, err := io.ReadAll(rec.Result().Body)
		require.NoError(t, err)
		require.Equal(t, []byte("done"), data)
	})

	t.Run("error causing post middleware without written response", func(t *testing.T) {

		ep := func(writer http.ResponseWriter, request *http.Request) error {
			// Note, no writes to the response so status code isn't set in the handler
			request.Header.Add("call", timestamp())
			return nil
		}

		ep = ApplyMiddleware(ep, ErrorCausingMiddleware(false))
		rec := httptest.NewRecorder()
		err := ep(rec, req)
		require.EqualError(t, err, "post error in middleware")
		require.True(t, simplerr.HasErrorCode(err, simplerr.CodeInvalidArgument))
		require.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
	})

}

func TestMiddlewareAdapter(t *testing.T) {

	//nolint:unparam
	ep := func(writer http.ResponseWriter, request *http.Request) error {
		request.Header.Add("call", timestamp())

		n, err := writer.Write([]byte("done"))
		require.NoError(t, err)
		require.Greater(t, n, 0)

		return nil
	}

	req, err := http.NewRequest("GET", "url", nil)
	require.NoError(t, err)

	t.Run("no errors", func(t *testing.T) {
		ep := ApplyMiddleware(ep, MiddlewareAdapter(StandardHTTPMiddleware))
		rec := httptest.NewRecorder()
		NewHandlerFuncAdapter(ep)(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		ensureHeaderOrder(t, rec, req)
	})

	t.Run("reverse adapter", func(t *testing.T) {
		// Convert the simplehttp middlewares to a standard lib middleware
		stdlibMiddlewarePre := MiddlewareReverseAdapter(PreCallMiddleware)
		stdlibMiddlewarePost := MiddlewareReverseAdapter(PostCallMiddleware)

		// Convert the simplehttp handler func to a standard lib handler
		stdlibHandler := http.HandlerFunc(NewHandlerAdapter(HandlerFunc(ep)).ServeHTTP)

		// Apply the middleware
		ep := stdlibMiddlewarePost(stdlibHandler)
		ep = stdlibMiddlewarePre(ep)

		rec := httptest.NewRecorder()
		ep.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		ensureHeaderOrder(t, rec, req)
	})

	t.Run("with pre-error", func(t *testing.T) {
		ep := ApplyMiddleware(ep, ErrorCausingMiddleware(true), MiddlewareAdapter(StandardHTTPMiddleware))
		rec := httptest.NewRecorder()
		NewHandlerFuncAdapter(ep)(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
		// There should be no response written because the endpoint should not have been called
		_, err = rec.Body.ReadByte()
		require.True(t, err == io.EOF)
	})

	t.Run("with post-error", func(t *testing.T) {
		ep := ApplyMiddleware(ep, ErrorCausingMiddleware(false), MiddlewareAdapter(StandardHTTPMiddleware))
		rec := httptest.NewRecorder()
		NewHandlerFuncAdapter(ep)(rec, req)

		// The endpoint writes a response before the error is raised so the status code is already set to OK
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		// There should be a response written because the endpoint should have been called
		data, err := io.ReadAll(rec.Result().Body)
		require.NoError(t, err)
		require.Equal(t, []byte("done"), data)
	})

}

func TestMiddlewareReverseAdapter(t *testing.T) {

	//nolint:unparam
	ep := func(writer http.ResponseWriter, request *http.Request) error {
		request.Header.Add("call", timestamp())

		n, err := writer.Write([]byte("done"))
		require.NoError(t, err)
		require.Greater(t, n, 0)

		return nil
	}

	req, err := http.NewRequest("GET", "url", nil)
	require.NoError(t, err)

	t.Run("pre error", func(t *testing.T) {
		// Convert the simplehttp middlewares to a standard lib middleware
		stdlibMiddlewarePre := MiddlewareReverseAdapter(ErrorCausingMiddleware(true))
		stdlibMiddlewarePost := MiddlewareReverseAdapter(PostCallMiddleware)

		// Convert the simplehttp handler func to a standard lib handler
		stdlibHandler := http.HandlerFunc(NewHandlerAdapter(HandlerFunc(ep)).ServeHTTP)

		// Apply the middleware
		ep := stdlibMiddlewarePost(stdlibHandler)
		ep = stdlibMiddlewarePre(ep)

		rec := httptest.NewRecorder()
		ep.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
		// There should be no response written because the endpoint should not have been called
		_, err = rec.Body.ReadByte()
		require.True(t, err == io.EOF)

		// The post call middleware should not be called
		h := rec.Header().Get("post-call-middleware")
		require.Empty(t, h, "post middleware should not have been called")
	})

	t.Run("post error", func(t *testing.T) {

		// Convert the simplehttp middlewares to a standard lib middleware
		stdlibMiddlewarePre := MiddlewareReverseAdapter(ErrorCausingMiddleware(false))

		// Convert the simplehttp handler func to a standard lib handler
		stdlibHandler := http.HandlerFunc(NewHandlerAdapter(HandlerFunc(ep)).ServeHTTP)

		// Apply the middleware
		ep := stdlibMiddlewarePre(stdlibHandler)

		rec := httptest.NewRecorder()
		ep.ServeHTTP(rec, req)

		// Because the handler func has been called, the status gets written and we can only
		// write the status once, so we cannot change the status from 200 even if we wanted.
		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

	})

}
