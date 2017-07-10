// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package utils

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestTimeMarker(t *testing.T) {
	assert := assert.New(t)
	m := &TimeMarker{
		interval: 1 * time.Second,
	}
	r1 := m.Mark()
	assert.True(r1)
	r2 := m.Mark()
	assert.False(r2)
	t.Log("Sleep for 2 seconds...")
	time.Sleep(2 * time.Second)
	r3 := m.Mark()
	assert.True(r3)
}

func TestScanMarkers(t *testing.T) {
	assert := assert.New(t)
	os.Setenv("HARBOR_SCAN_ALL_INTERVAL", "5")
	sm := ScanAllMarker()
	d := sm.Next().Sub(time.Now())
	assert.True(d <= 5*time.Minute)
	som := ScanOverviewMarker()
	d = som.Next().Sub(time.Now())
	assert.True(d <= 15*time.Second)
}
