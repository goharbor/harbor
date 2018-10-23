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
package period

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestPeriodicJobPolicy(t *testing.T) {
	p := createPolicy("")

	data, err := p.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	if err := p.DeSerialize(data); err != nil {
		t.Fatal(err)
	}
}

func TestPeriodicJobPolicyStore(t *testing.T) {
	ps := &periodicJobPolicyStore{
		lock:     new(sync.RWMutex),
		policies: make(map[string]*PeriodicJobPolicy),
	}

	ps.add(createPolicy("fake_ID_Steven"))
	if ps.size() != 1 {
		t.Errorf("expect size 1 but got '%d'\n", ps.size())
	}
	pl := make([]*PeriodicJobPolicy, 0)
	pl = append(pl, createPolicy(""))
	pl = append(pl, createPolicy(""))
	ps.addAll(pl)
	if ps.size() != 3 {
		t.Fatalf("expect size 3 but got '%d'\n", ps.size())
	}

	l := ps.list()
	if l == nil || len(l) != 3 {
		t.Fatal("expect a policy list with 3 items but got invalid list")
	}

	rp := ps.remove("fake_ID_Steven")
	if rp == nil {
		t.Fatal("expect none nil policy object but got nil")
	}
}

func createPolicy(id string) *PeriodicJobPolicy {
	theID := id
	if theID == "" {
		theID = fmt.Sprintf("fake_ID_%d", time.Now().UnixNano()+int64(rand.Intn(1000)))
	}
	p := &PeriodicJobPolicy{
		PolicyID:      theID,
		JobName:       "fake_job",
		JobParameters: make(map[string]interface{}),
		CronSpec:      "5 * * * * *",
	}
	p.JobParameters["image"] = "testing:v1"

	return p
}
