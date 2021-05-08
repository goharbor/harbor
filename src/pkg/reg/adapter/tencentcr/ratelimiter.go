package tencentcr

import (
	"net/http"
	"sync"

	"go.uber.org/ratelimit"
)

type limitTransport struct {
	http.RoundTripper
	limiter ratelimit.Limiter
}

var _ http.RoundTripper = limitTransport{}

func (t limitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.limiter.Take()
	return t.RoundTripper.RoundTrip(req)
}

var limiterOnce sync.Once
var limiter ratelimit.Limiter

func newLimiter(rate int) ratelimit.Limiter {
	limiterOnce.Do(func() {
		limiter = ratelimit.New(rate)
	})
	return limiter
}

func newRateLimitedTransport(rate int, transport http.RoundTripper) http.RoundTripper {
	return &limitTransport{
		RoundTripper: transport,
		limiter:      newLimiter(rate),
	}
}
