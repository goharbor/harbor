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

// +build go1.9

package redis

import "testing"

func TestPoolList(t *testing.T) {
	var idle idleList
	var a, b, c idleConn

	check := func(ics ...*idleConn) {
		if idle.count != len(ics) {
			t.Fatal("idle.count != len(ics)")
		}
		if len(ics) == 0 {
			if idle.front != nil {
				t.Fatalf("front not nil")
			}
			if idle.back != nil {
				t.Fatalf("back not nil")
			}
			return
		}
		if idle.front != ics[0] {
			t.Fatal("front != ics[0]")
		}
		if idle.back != ics[len(ics)-1] {
			t.Fatal("back != ics[len(ics)-1]")
		}
		if idle.front.prev != nil {
			t.Fatal("front.prev != nil")
		}
		if idle.back.next != nil {
			t.Fatal("back.next != nil")
		}
		for i := 1; i < len(ics)-1; i++ {
			if ics[i-1].next != ics[i] {
				t.Fatal("ics[i-1].next != ics[i]")
			}
			if ics[i+1].prev != ics[i] {
				t.Fatal("ics[i+1].prev != ics[i]")
			}
		}
	}

	idle.pushFront(&c)
	check(&c)
	idle.pushFront(&b)
	check(&b, &c)
	idle.pushFront(&a)
	check(&a, &b, &c)
	idle.popFront()
	check(&b, &c)
	idle.popFront()
	check(&c)
	idle.popFront()
	check()

	idle.pushFront(&c)
	check(&c)
	idle.pushFront(&b)
	check(&b, &c)
	idle.pushFront(&a)
	check(&a, &b, &c)
	idle.popBack()
	check(&a, &b)
	idle.popBack()
	check(&a)
	idle.popBack()
	check()
}
