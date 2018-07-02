package utils

import (
	"fmt"
	"net/http"
	"time"

	ctxu "github.com/docker/distribution/context"
	"github.com/docker/distribution/registry/api/errcode"
	"github.com/docker/distribution/registry/auth"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"

	"github.com/docker/notary"
	"github.com/docker/notary/tuf/signed"
)

// ContextHandler defines an alternate HTTP handler interface which takes in
// a context for authorization and returns an HTTP application error.
type ContextHandler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// rootHandler is an implementation of an HTTP request handler which handles
// authorization and calling out to the defined alternate http handler.
type rootHandler struct {
	handler ContextHandler
	auth    auth.AccessController
	actions []string
	context context.Context
	trust   signed.CryptoService
}

// AuthWrapper wraps a Handler with and Auth requirement
type AuthWrapper func(ContextHandler, ...string) *rootHandler

// RootHandlerFactory creates a new rootHandler factory  using the given
// Context creator and authorizer.  The returned factory allows creating
// new rootHandlers from the alternate http handler contextHandler and
// a scope.
func RootHandlerFactory(ctx context.Context, auth auth.AccessController, trust signed.CryptoService) func(ContextHandler, ...string) *rootHandler {
	return func(handler ContextHandler, actions ...string) *rootHandler {
		return &rootHandler{
			handler: handler,
			auth:    auth,
			actions: actions,
			context: ctx,
			trust:   trust,
		}
	}
}

// ServeHTTP serves an HTTP request and implements the http.Handler interface.
func (root *rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		ctx  = ctxu.WithRequest(root.context, r)
		log  = ctxu.GetRequestLogger(ctx)
		vars = mux.Vars(r)
	)
	ctx, w = ctxu.WithResponseWriter(ctx, w)
	ctx = ctxu.WithLogger(ctx, log)
	ctx = context.WithValue(ctx, notary.CtxKeyCryptoSvc, root.trust)

	defer func(ctx context.Context) {
		ctxu.GetResponseLogger(ctx).Info("response completed")
	}(ctx)

	if root.auth != nil {
		ctx = context.WithValue(ctx, notary.CtxKeyRepo, vars["gun"])
		if ctx, err = root.doAuth(ctx, vars["gun"], w); err != nil {
			// errors have already been logged/output to w inside doAuth
			// just return
			return
		}
	}
	if err := root.handler(ctx, w, r); err != nil {
		serveError(log, w, err)
	}
}

func serveError(log ctxu.Logger, w http.ResponseWriter, err error) {
	if httpErr, ok := err.(errcode.Error); ok {
		// info level logging for non-5XX http errors
		httpErrCode := httpErr.ErrorCode().Descriptor().HTTPStatusCode
		if httpErrCode >= http.StatusInternalServerError {
			// error level logging for 5XX http errors
			log.Errorf("%s: %s: %v", httpErr.ErrorCode().Error(), httpErr.Message, httpErr.Detail)
		} else {
			log.Infof("%s: %s: %v", httpErr.ErrorCode().Error(), httpErr.Message, httpErr.Detail)
		}
	}
	e := errcode.ServeJSON(w, err)
	if e != nil {
		log.Error(e)
	}
	return
}

func (root *rootHandler) doAuth(ctx context.Context, gun string, w http.ResponseWriter) (context.Context, error) {
	var access []auth.Access
	if gun == "" {
		access = buildCatalogRecord(root.actions...)
	} else {
		access = buildAccessRecords(gun, root.actions...)
	}

	log := ctxu.GetRequestLogger(ctx)
	var authCtx context.Context
	var err error
	if authCtx, err = root.auth.Authorized(ctx, access...); err != nil {
		if challenge, ok := err.(auth.Challenge); ok {
			// Let the challenge write the response.
			challenge.SetHeaders(w)

			if err := errcode.ServeJSON(w, errcode.ErrorCodeUnauthorized.WithDetail(access)); err != nil {
				log.Errorf("failed to serve challenge response: %s", err.Error())
				return nil, err
			}
			return nil, err
		}
		errcode.ServeJSON(w, errcode.ErrorCodeUnauthorized)
		return nil, err
	}
	return authCtx, nil
}

func buildAccessRecords(repo string, actions ...string) []auth.Access {
	requiredAccess := make([]auth.Access, 0, len(actions))
	for _, action := range actions {
		requiredAccess = append(requiredAccess, auth.Access{
			Resource: auth.Resource{
				Type: "repository",
				Name: repo,
			},
			Action: action,
		})
	}
	return requiredAccess
}

// buildCatalogRecord returns the only valid format for the catalog
// resource. Only admins can get this access level from the token
// server.
func buildCatalogRecord(actions ...string) []auth.Access {
	requiredAccess := []auth.Access{{
		Resource: auth.Resource{
			Type: "registry",
			Name: "catalog",
		},
		Action: "*",
	}}

	return requiredAccess
}

// CacheControlConfig is an interface for something that knows how to set cache
// control headers
type CacheControlConfig interface {
	// SetHeaders will actually set the cache control headers on a Headers object
	SetHeaders(headers http.Header)
}

// NewCacheControlConfig returns CacheControlConfig interface for either setting
// cache control or disabling cache control entirely
func NewCacheControlConfig(maxAgeInSeconds int, mustRevalidate bool) CacheControlConfig {
	if maxAgeInSeconds > 0 {
		return PublicCacheControl{MustReValidate: mustRevalidate, MaxAgeInSeconds: maxAgeInSeconds}
	}
	return NoCacheControl{}
}

// PublicCacheControl is a set of options that we will set to enable cache control
type PublicCacheControl struct {
	MustReValidate  bool
	MaxAgeInSeconds int
}

// SetHeaders sets the public headers with an optional must-revalidate header
func (p PublicCacheControl) SetHeaders(headers http.Header) {
	cacheControlValue := fmt.Sprintf("public, max-age=%v, s-maxage=%v",
		p.MaxAgeInSeconds, p.MaxAgeInSeconds)

	if p.MustReValidate {
		cacheControlValue = fmt.Sprintf("%s, must-revalidate", cacheControlValue)
	}
	headers.Set("Cache-Control", cacheControlValue)
	// delete the Pragma directive, because the only valid value in HTTP is
	// "no-cache"
	headers.Del("Pragma")
	if headers.Get("Last-Modified") == "" {
		SetLastModifiedHeader(headers, time.Time{})
	}
}

// NoCacheControl is an object which represents a directive to cache nothing
type NoCacheControl struct{}

// SetHeaders sets the public headers cache-control headers and pragma to no-cache
func (n NoCacheControl) SetHeaders(headers http.Header) {
	headers.Set("Cache-Control", "max-age=0, no-cache, no-store")
	headers.Set("Pragma", "no-cache")
}

// cacheControlResponseWriter wraps an existing response writer, and if Write is
// called, will try to set the cache control headers if it can
type cacheControlResponseWriter struct {
	http.ResponseWriter
	config     CacheControlConfig
	statusCode int
}

// WriteHeader stores the header before writing it, so we can tell if it's been set
// to a non-200 status code
func (c *cacheControlResponseWriter) WriteHeader(statusCode int) {
	c.statusCode = statusCode
	c.ResponseWriter.WriteHeader(statusCode)
}

// Write will set the cache headers if they haven't already been set and if the status
// code has either not been set or set to 200
func (c *cacheControlResponseWriter) Write(data []byte) (int, error) {
	if c.statusCode == http.StatusOK || c.statusCode == 0 {
		headers := c.ResponseWriter.Header()
		if headers.Get("Cache-Control") == "" {
			c.config.SetHeaders(headers)
		}
	}
	return c.ResponseWriter.Write(data)
}

type cacheControlHandler struct {
	http.Handler
	config CacheControlConfig
}

func (c cacheControlHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.Handler.ServeHTTP(&cacheControlResponseWriter{ResponseWriter: w, config: c.config}, r)
}

// WrapWithCacheHandler wraps another handler in one that can add cache control headers
// given a 200 response
func WrapWithCacheHandler(ccc CacheControlConfig, handler http.Handler) http.Handler {
	if ccc != nil {
		return cacheControlHandler{Handler: handler, config: ccc}
	}
	return handler
}

// SetLastModifiedHeader takes a time and uses it to set the LastModified header using
// the right date format
func SetLastModifiedHeader(headers http.Header, lmt time.Time) {
	headers.Set("Last-Modified", lmt.Format(time.RFC1123))
}
