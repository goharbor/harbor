// Copyright 2018 Gary Burd
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// +build go1.7

package redis_test

import (
	"context"
	"testing"

	"github.com/garyburd/redigo/redis"
)

func TestWaitPoolGetAfterClose(t *testing.T) {
	d := poolDialer{t: t}
	p := &redis.Pool{
		MaxIdle:   1,
		MaxActive: 1,
		Dial:      d.dial,
		Wait:      true,
	}
	p.Close()
	_, err := p.GetContext(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWaitPoolGetCanceledContext(t *testing.T) {
	d := poolDialer{t: t}
	p := &redis.Pool{
		MaxIdle:   1,
		MaxActive: 1,
		Dial:      d.dial,
		Wait:      true,
	}
	defer p.Close()
	ctx, f := context.WithCancel(context.Background())
	f()
	c := p.Get()
	defer c.Close()
	_, err := p.GetContext(ctx)
	if err != context.Canceled {
		t.Fatalf("got error %v, want %v", err, context.Canceled)
	}
}
