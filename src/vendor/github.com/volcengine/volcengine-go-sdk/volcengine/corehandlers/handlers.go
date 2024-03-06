package corehandlers

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/credentials"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
	"github.com/volcengine/volcengine-go-sdk/volcengine/volcengineerr"
)

// Interface for matching types which also have a Len method.
type lener interface {
	Len() int
}

// BuildContentLengthHandler builds the content length of a request based on the volcenginebody,
// or will use the HTTPRequest.Header's "Content-Length" if defined. If unable
// to determine request volcenginebody length and no "Content-Length" was specified it will panic.
//
// The Content-Length will only be added to the request if the length of the volcenginebody
// is greater than 0. If the volcenginebody is empty or the current `Content-Length`
// header is <= 0, the header will also be stripped.
var BuildContentLengthHandler = request.NamedHandler{Name: "core.BuildContentLengthHandler", Fn: func(r *request.Request) {
	var length int64

	if slength := r.HTTPRequest.Header.Get("Content-Length"); slength != "" {
		length, _ = strconv.ParseInt(slength, 10, 64)
	} else {
		if r.Body != nil {
			var err error
			length, err = volcengine.SeekerLen(r.Body)
			if err != nil {
				r.Error = volcengineerr.New(request.ErrCodeSerialization, "failed to get request volcenginebody's length", err)
				return
			}
		}
	}

	if length > 0 {
		r.HTTPRequest.ContentLength = length
		r.HTTPRequest.Header.Set("Content-Length", fmt.Sprintf("%d", length))
	} else {
		r.HTTPRequest.ContentLength = 0
		r.HTTPRequest.Header.Del("Content-Length")
	}
}}

var reStatusCode = regexp.MustCompile(`^(\d{3})`)

// ValidateReqSigHandler is a request handler to ensure that the request's
// signature doesn't expire before it is sent. This can happen when a request
// is built and signed significantly before it is sent. Or significant delays
// occur when retrying requests that would cause the signature to expire.
var ValidateReqSigHandler = request.NamedHandler{
	Name: "core.ValidateReqSigHandler",
	Fn: func(r *request.Request) {
		// Unsigned requests are not signed
		if r.Config.Credentials == credentials.AnonymousCredentials {
			return
		}

		signedTime := r.Time
		if !r.LastSignedAt.IsZero() {
			signedTime = r.LastSignedAt
		}

		// 5 minutes to allow for some clock skew/delays in transmission.
		// Would be improved with volcengine/volcengine-go-sdk#423
		if signedTime.Add(5 * time.Minute).After(time.Now()) {
			return
		}

		fmt.Println("request expired, resigning")
		r.Sign()
	},
}

// SendHandler is a request handler to send service request using HTTP client.
var SendHandler = request.NamedHandler{
	Name: "core.SendHandler",
	Fn: func(r *request.Request) {
		sender := sendFollowRedirects
		if r.DisableFollowRedirects {
			sender = sendWithoutFollowRedirects
		}

		if request.NoBody == r.HTTPRequest.Body {
			// Strip off the request volcenginebody if the NoBody reader was used as a
			// place holder for a request volcenginebody. This prevents the SDK from
			// making requests with a request volcenginebody when it would be invalid
			// to do so.
			//
			// Use a shallow copy of the http.Request to ensure the race condition
			// of transport on Body will not trigger
			reqOrig, reqCopy := r.HTTPRequest, *r.HTTPRequest
			reqCopy.Body = nil
			r.HTTPRequest = &reqCopy
			defer func() {
				r.HTTPRequest = reqOrig
			}()
		}

		var err error
		r.HTTPResponse, err = sender(r)
		if err != nil {
			handleSendError(r, err)
		}
	},
}

func sendFollowRedirects(r *request.Request) (*http.Response, error) {
	return r.Config.HTTPClient.Do(r.HTTPRequest)
}

func sendWithoutFollowRedirects(r *request.Request) (*http.Response, error) {
	transport := r.Config.HTTPClient.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	return transport.RoundTrip(r.HTTPRequest)
}

func handleSendError(r *request.Request, err error) {
	// Prevent leaking if an HTTPResponse was returned. Clean up
	// the volcenginebody.
	if r.HTTPResponse != nil {
		r.HTTPResponse.Body.Close()
	}
	// Capture the case where url.Error is returned for error processing
	// response. e.g. 301 without location header comes back as string
	// error and r.HTTPResponse is nil. Other URL redirect errors will
	// comeback in a similar method.
	if e, ok := err.(*url.Error); ok && e.Err != nil {
		if s := reStatusCode.FindStringSubmatch(e.Err.Error()); s != nil {
			code, _ := strconv.ParseInt(s[1], 10, 64)
			r.HTTPResponse = &http.Response{
				StatusCode: int(code),
				Status:     http.StatusText(int(code)),
				Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
			}
			return
		}
	}
	if r.HTTPResponse == nil {
		// Add a dummy request response object to ensure the HTTPResponse
		// value is consistent.
		r.HTTPResponse = &http.Response{
			StatusCode: int(0),
			Status:     http.StatusText(int(0)),
			Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
		}
	}
	// Catch all request errors, and let the default retrier determine
	// if the error is retryable.
	r.Error = volcengineerr.New("RequestError", "send request failed", err)

	// Override the error with a context canceled error, if that was canceled.
	ctx := r.Context()
	select {
	case <-ctx.Done():
		r.Error = volcengineerr.New(request.CanceledErrorCode,
			"request context canceled", ctx.Err())
		r.Retryable = volcengine.Bool(false)
	default:
	}
}

// ValidateResponseHandler is a request handler to validate service response.
var ValidateResponseHandler = request.NamedHandler{Name: "core.ValidateResponseHandler", Fn: func(r *request.Request) {
	if r.HTTPResponse.StatusCode == 0 || r.HTTPResponse.StatusCode >= 300 {
		// this may be replaced by an UnmarshalError handler
		r.Error = volcengineerr.New("UnknownError", "unknown error", nil)
	}
}}

// AfterRetryHandler performs final checks to determine if the request should
// be retried and how long to delay.
var AfterRetryHandler = request.NamedHandler{
	Name: "core.AfterRetryHandler",
	Fn: func(r *request.Request) {
		// If one of the other handlers already set the retry state
		// we don't want to override it based on the service's state
		if r.Retryable == nil || volcengine.BoolValue(r.Config.EnforceShouldRetryCheck) {
			r.Retryable = volcengine.Bool(r.ShouldRetry(r))
		}

		if r.WillRetry() {
			r.RetryDelay = r.RetryRules(r)

			if sleepFn := r.Config.SleepDelay; sleepFn != nil {
				// Support SleepDelay for backwards compatibility and testing
				sleepFn(r.RetryDelay)
			} else if err := volcengine.SleepWithContext(r.Context(), r.RetryDelay); err != nil {
				r.Error = volcengineerr.New(request.CanceledErrorCode,
					"request context canceled", err)
				r.Retryable = volcengine.Bool(false)
				return
			}

			// when the expired token exception occurs the credentials
			// need to be expired locally so that the next request to
			// get credentials will trigger a credentials refresh.
			if r.IsErrorExpired() {
				r.Config.Credentials.Expire()
			}

			r.RetryCount++
			r.Error = nil
		}
	}}

// ValidateEndpointHandler is a request handler to validate a request had the
// appropriate Region and Endpoint set. Will set r.Error if the endpoint or
// region is not valid.
var ValidateEndpointHandler = request.NamedHandler{Name: "core.ValidateEndpointHandler", Fn: func(r *request.Request) {
	if r.ClientInfo.SigningRegion == "" && volcengine.StringValue(r.Config.Region) == "" && r.Config.DynamicCredentials == nil {
		r.Error = volcengine.ErrMissingRegion
	} else if r.ClientInfo.Endpoint == "" {
		r.Error = volcengine.ErrMissingEndpoint
	}
}}
