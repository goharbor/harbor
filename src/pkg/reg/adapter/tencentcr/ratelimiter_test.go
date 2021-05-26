package tencentcr

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_newRateLimitedTransport(t *testing.T) {
	tests := []struct {
		name      string
		rate      int
		transport http.RoundTripper
	}{
		{"1qps", 1, http.DefaultTransport},
	}
	var req = &http.Request{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newRateLimitedTransport(tt.rate, tt.transport)
			start := time.Now()
			for i := 0; i <= tt.rate; i++ {
				got.RoundTrip(req)
			}
			used := int64(time.Since(start).Milliseconds()) / int64(tt.rate)
			assert.GreaterOrEqualf(t, used/int64(tt.rate), int64(1e3/tt.rate), "used %d ms per req", used/int64(tt.rate))
		})
	}
}
