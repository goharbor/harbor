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

package rds

import (
	"encoding/json"
	"github.com/goharbor/harbor/src/jobservice/tests"
	"testing"
	"time"
)

var (
	pool      = tests.GiveMeRedisPool()
	namespace = tests.GiveMeTestNamespace()
)

// For testing
type simpleStatusChange struct {
	JobID string
}

func TestZPopMin(t *testing.T) {
	conn := pool.Get()
	defer conn.Close()

	s1 := &simpleStatusChange{"a"}
	s2 := &simpleStatusChange{"b"}

	raw1, _ := json.Marshal(s1)
	raw2, _ := json.Marshal(s2)

	key := KeyStatusUpdateRetryQueue(namespace)
	_, err := conn.Do("ZADD", key, time.Now().Unix(), raw1)
	_, err = conn.Do("ZADD", key, time.Now().Unix()+5, raw2)
	if err != nil {
		t.Fatal(err)
	}

	v, err := ZPopMin(conn, key)
	if err != nil {
		t.Fatal(err)
	}

	change1 := &simpleStatusChange{}
	json.Unmarshal(v.([]byte), change1)
	if change1.JobID != "a" {
		t.Errorf("expect min element 'a' but got '%s'", change1.JobID)
	}

	v, err = ZPopMin(conn, key)
	if err != nil {
		t.Fatal(err)
	}

	change2 := &simpleStatusChange{}
	json.Unmarshal(v.([]byte), change2)
	if change2.JobID != "b" {
		t.Errorf("expect min element 'b' but got '%s'", change2.JobID)
	}
}
