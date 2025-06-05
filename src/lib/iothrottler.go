// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lib

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
