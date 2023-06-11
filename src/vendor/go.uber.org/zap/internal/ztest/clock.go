// Copyright (c) 2021 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package ztest

import (
	"time"

	"github.com/benbjohnson/clock"
)

// MockClock provides control over the time.
type MockClock struct{ m *clock.Mock }

// NewMockClock builds a new mock clock that provides control of time.
func NewMockClock() *MockClock {
	return &MockClock{clock.NewMock()}
}

// Now reports the current time.
func (c *MockClock) Now() time.Time {
	return c.m.Now()
}

// NewTicker returns a time.Ticker that ticks at the specified frequency.
func (c *MockClock) NewTicker(d time.Duration) *time.Ticker {
	return &time.Ticker{C: c.m.Ticker(d).C}
}

// Add progresses time by the given duration.
func (c *MockClock) Add(d time.Duration) {
	c.m.Add(d)
}
