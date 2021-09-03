package transfer

import (
	"fmt"
	"io"
	"time"

	"golang.org/x/time/rate"
)

type reader struct {
	reader  io.ReadCloser
	limiter *rate.Limiter
}

type RateOpts struct {
	Rate float64
}

const KBRATE = 1024 / 8

// NewReader returns a Reader that is rate limited
func NewReader(r io.ReadCloser, kb int32) io.ReadCloser {
	l := rate.NewLimiter(rate.Limit(kb*KBRATE), 1000*1024)
	return &reader{
		reader:  r,
		limiter: l,
	}
}

func (r *reader) Read(buf []byte) (int, error) {
	n, err := r.reader.Read(buf)
	if n <= 0 {
		return n, err
	}
	now := time.Now()
	rv := r.limiter.ReserveN(now, n)
	if !rv.OK() {
		return 0, fmt.Errorf("exceeds limiter's burst")
	}
	delay := rv.DelayFrom(now)
	time.Sleep(delay)
	return n, err
}

func (r *reader) Close() error {
	return r.reader.Close()
}
