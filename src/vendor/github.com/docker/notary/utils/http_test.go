package utils

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/docker/distribution/registry/api/errcode"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"

	"github.com/docker/notary/tuf/signed"
)

func MockContextHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return nil
}

func MockBetterErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return errcode.ErrorCodeUnknown.WithDetail("Test Error")
}

func TestRootHandlerFactory(t *testing.T) {
	hand := RootHandlerFactory(context.Background(), nil, &signed.Ed25519{})
	handler := hand(MockContextHandler)
	if _, ok := interface{}(handler).(http.Handler); !ok {
		t.Fatalf("A rootHandler must implement the http.Handler interface")
	}

	ts := httptest.NewServer(handler)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestRootHandlerError(t *testing.T) {
	hand := RootHandlerFactory(context.Background(), nil, &signed.Ed25519{})
	handler := hand(MockBetterErrorHandler)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, res.StatusCode)

	content, err := ioutil.ReadAll(res.Body)

	require.NoError(t, err)
	contentStr := strings.Trim(string(content), "\r\n\t ")
	if strings.TrimSpace(contentStr) != `{"errors":[{"code":"UNKNOWN","message":"unknown error","detail":"Test Error"}]}` {
		t.Fatalf("Error Body Incorrect: `%s`", content)
	}
}

// If no CacheControlConfig is passed, wrapping the handler just returns the handler
func TestWrapWithCacheHeaderNilCacheControlConfig(t *testing.T) {
	mux := http.NewServeMux()
	wrapped := WrapWithCacheHandler(nil, mux)
	require.Equal(t, mux, wrapped)
}

// If the wrapped handler returns a non-200, no matter which CacheControlConfig is
// used, the Cache-Control header not set.
func TestWrapWithCacheHeaderNon200Response(t *testing.T) {
	mux := http.NewServeMux()
	configs := []CacheControlConfig{NewCacheControlConfig(10, true), NewCacheControlConfig(0, true)}

	for _, conf := range configs {
		req := &http.Request{URL: &url.URL{Path: "/"}, Body: ioutil.NopCloser(bytes.NewBuffer(nil))}

		wrapped := WrapWithCacheHandler(conf, mux)
		require.NotEqual(t, mux, wrapped)
		rw := httptest.NewRecorder()
		wrapped.ServeHTTP(rw, req)

		require.Equal(t, "", rw.HeaderMap.Get("Cache-Control"))
		require.Equal(t, "", rw.HeaderMap.Get("Last-Modified"))
		require.Equal(t, "", rw.HeaderMap.Get("Pragma"))
	}
}

// If the wrapped handler writes no cache headers whatsoever, and a PublicCacheControl
// is used, the Cache-Control header is set with the given maxAge and re-validate value.
// The Last-Modified header is also set to the beginning of (computer) time.  If a
// Pragma header is written is deleted
func TestWrapWithCacheHeaderPublicCacheControlNoCacheHeaders(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello!"))
	})
	mux.HandleFunc("/a", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Pragma", "no-cache")
		w.Write([]byte("hello!"))
	})

	for _, path := range []string{"/", "/a"} {
		req := &http.Request{URL: &url.URL{Path: path}, Body: ioutil.NopCloser(bytes.NewBuffer(nil))}

		// must-revalidate is set if revalidate is set to true, and not if revalidate is set to false
		for _, revalidate := range []bool{true, false} {
			wrapped := WrapWithCacheHandler(NewCacheControlConfig(10, revalidate), mux)
			require.NotEqual(t, mux, wrapped)
			rw := httptest.NewRecorder()
			wrapped.ServeHTTP(rw, req)

			cacheControl := "public, max-age=10, s-maxage=10"
			if revalidate {
				cacheControl = cacheControl + ", must-revalidate"
			}
			require.Equal(t, cacheControl, rw.HeaderMap.Get("Cache-Control"))

			lastModified, err := time.Parse(time.RFC1123, rw.HeaderMap.Get("Last-Modified"))
			require.NoError(t, err)
			require.True(t, lastModified.Equal(time.Time{}))
			require.Equal(t, "", rw.HeaderMap.Get("Pragma"))
		}
	}
}

// If the wrapped handler writes a last modified header, and a PublicCacheControl
// is used, the Cache-Control header is set with the given maxAge and re-validate value.
// The Last-Modified header is not replaced. The Pragma header is deleted though.
func TestWrapWithCacheHeaderPublicCacheControlLastModifiedHeader(t *testing.T) {
	now := time.Now()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		SetLastModifiedHeader(w.Header(), now)
		w.Header().Set("Pragma", "no-cache")
		w.Write([]byte("hello!"))
	})
	req := &http.Request{URL: &url.URL{Path: "/"}, Body: ioutil.NopCloser(bytes.NewBuffer(nil))}

	wrapped := WrapWithCacheHandler(NewCacheControlConfig(10, true), mux)
	require.NotEqual(t, mux, wrapped)
	rw := httptest.NewRecorder()
	wrapped.ServeHTTP(rw, req)

	require.Equal(t, "public, max-age=10, s-maxage=10, must-revalidate", rw.HeaderMap.Get("Cache-Control"))
	lastModified, err := time.Parse(time.RFC1123, rw.HeaderMap.Get("Last-Modified"))
	require.NoError(t, err)
	// RFC1123 does not include nanoseconds
	nowToNearestSecond := now.Add(time.Duration(-1 * now.Nanosecond()))
	require.True(t, lastModified.Equal(nowToNearestSecond))
	require.Equal(t, "", rw.HeaderMap.Get("Pragma"))
}

// If the wrapped handler writes a Cache-Control header, even if the last modified
// header is not written, then the Cache-Control header is not written, nor is a
// Last-Modified header written.  The Pragma header is not deleted.
func TestWrapWithCacheHeaderPublicCacheControlCacheControlHeader(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "some invalid cache control value")
		w.Header().Set("Pragma", "invalid value")
		w.Write([]byte("hello!"))
	})
	req := &http.Request{URL: &url.URL{Path: "/"}, Body: ioutil.NopCloser(bytes.NewBuffer(nil))}

	wrapped := WrapWithCacheHandler(NewCacheControlConfig(10, true), mux)
	require.NotEqual(t, mux, wrapped)
	rw := httptest.NewRecorder()
	wrapped.ServeHTTP(rw, req)

	require.Equal(t, "some invalid cache control value", rw.HeaderMap.Get("Cache-Control"))
	require.Equal(t, "", rw.HeaderMap.Get("Last-Modified"))
	require.Equal(t, "invalid value", rw.HeaderMap.Get("Pragma"))
}

// If the wrapped handler writes no cache headers whatsoever, and NoCacheControl
// is used, the Cache-Control and Pragma headers are set with no-cache.
func TestWrapWithCacheHeaderNoCacheControlNoCacheHeaders(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Pragma", "invalid value")
		w.Write([]byte("hello!"))
	})
	req := &http.Request{URL: &url.URL{Path: "/"}, Body: ioutil.NopCloser(bytes.NewBuffer(nil))}

	wrapped := WrapWithCacheHandler(NewCacheControlConfig(0, false), mux)
	require.NotEqual(t, mux, wrapped)
	rw := httptest.NewRecorder()
	wrapped.ServeHTTP(rw, req)

	require.Equal(t, "max-age=0, no-cache, no-store", rw.HeaderMap.Get("Cache-Control"))
	require.Equal(t, "", rw.HeaderMap.Get("Last-Modified"))
	require.Equal(t, "no-cache", rw.HeaderMap.Get("Pragma"))
}

// If the wrapped handler writes a last modified header, and NoCacheControl
// is used, the Cache-Control and Pragma headers are set with no-cache without
// messing with the Last-Modified header.
func TestWrapWithCacheHeaderNoCacheControlLastModifiedHeader(t *testing.T) {
	now := time.Now()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		SetLastModifiedHeader(w.Header(), now)
		w.Write([]byte("hello!"))
	})
	req := &http.Request{URL: &url.URL{Path: "/"}, Body: ioutil.NopCloser(bytes.NewBuffer(nil))}

	wrapped := WrapWithCacheHandler(NewCacheControlConfig(0, true), mux)
	require.NotEqual(t, mux, wrapped)
	rw := httptest.NewRecorder()
	wrapped.ServeHTTP(rw, req)

	require.Equal(t, "max-age=0, no-cache, no-store", rw.HeaderMap.Get("Cache-Control"))
	require.Equal(t, "no-cache", rw.HeaderMap.Get("Pragma"))

	lastModified, err := time.Parse(time.RFC1123, rw.HeaderMap.Get("Last-Modified"))
	require.NoError(t, err)
	// RFC1123 does not include nanoseconds
	nowToNearestSecond := now.Add(time.Duration(-1 * now.Nanosecond()))
	require.True(t, lastModified.Equal(nowToNearestSecond))
}

// If the wrapped handler writes a Cache-Control header, even if the last modified
// header is not written, then the Cache-Control header is not written, nor is a
// Pragma added.  The Last-Modified header is untouched.
func TestWrapWithCacheHeaderNoCacheControlCacheControlHeader(t *testing.T) {
	now := time.Now()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "some invalid cache control value")
		SetLastModifiedHeader(w.Header(), now)
		w.Write([]byte("hello!"))
	})
	req := &http.Request{URL: &url.URL{Path: "/"}, Body: ioutil.NopCloser(bytes.NewBuffer(nil))}

	wrapped := WrapWithCacheHandler(NewCacheControlConfig(0, true), mux)
	require.NotEqual(t, mux, wrapped)
	rw := httptest.NewRecorder()
	wrapped.ServeHTTP(rw, req)

	require.Equal(t, "some invalid cache control value", rw.HeaderMap.Get("Cache-Control"))
	require.Equal(t, "", rw.HeaderMap.Get("Pragma"))

	lastModified, err := time.Parse(time.RFC1123, rw.HeaderMap.Get("Last-Modified"))
	require.NoError(t, err)
	// RFC1123 does not include nanoseconds
	nowToNearestSecond := now.Add(time.Duration(-1 * now.Nanosecond()))
	require.True(t, lastModified.Equal(nowToNearestSecond))
}

func TestBuildCatalogRecord(t *testing.T) {
	r := buildCatalogRecord()
	require.Len(t, r, 1)
	r0 := r[0]
	require.Equal(t, "registry", r0.Resource.Type)
	require.Equal(t, "catalog", r0.Resource.Name)
	require.Equal(t, "*", r0.Action)
}

func TestDoAuthNonWildcardImage(t *testing.T) {
	// success
	ac := TestingAccessController{}
	r := rootHandler{
		auth: ac,
	}
	rec := httptest.NewRecorder()
	_, err := r.doAuth(
		context.Background(),
		"docker.io/library/alpine",
		rec,
	)
	require.NoError(t, err)
	require.Equal(t, 200, rec.Code)

	// challenge error
	e := TestingAuthChallenge{}
	ac = TestingAccessController{
		Err: &e,
	}
	r = rootHandler{
		auth: ac,
	}
	rec = httptest.NewRecorder()
	_, err = r.doAuth(
		context.Background(),
		"docker.io/library/alpine",
		rec,
	)
	require.Error(t, err)
	require.True(t, e.SetHeadersCalled)
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	// non-challenge error
	ac = TestingAccessController{
		Err: errors.New("Non challenge error"),
	}
	r = rootHandler{
		auth: ac,
	}
	rec = httptest.NewRecorder()
	_, err = r.doAuth(
		context.Background(),
		"docker.io/library/alpine",
		rec,
	)
	require.Error(t, err)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestDoAuthWildcardImage(t *testing.T) {
	// success
	ac := TestingAccessController{}
	r := rootHandler{
		auth: ac,
	}
	rec := httptest.NewRecorder()
	_, err := r.doAuth(
		context.Background(),
		"",
		rec,
	)
	require.NoError(t, err)
	require.Equal(t, 200, rec.Code)

	// challenge error
	e := TestingAuthChallenge{}
	ac = TestingAccessController{
		Err: &e,
	}
	r = rootHandler{
		auth: ac,
	}
	rec = httptest.NewRecorder()
	_, err = r.doAuth(
		context.Background(),
		"",
		rec,
	)
	require.Error(t, err)
	require.True(t, e.SetHeadersCalled)
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	// non-challenge error
	ac = TestingAccessController{
		Err: errors.New("Non challenge error"),
	}
	r = rootHandler{
		auth: ac,
	}
	rec = httptest.NewRecorder()
	_, err = r.doAuth(
		context.Background(),
		"",
		rec,
	)
	require.Error(t, err)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
