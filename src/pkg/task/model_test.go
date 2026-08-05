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

package task

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/pkg/task/dao"
)

func TestTaskFrom_LargeIDInExtraAttrs(t *testing.T) {
	// 2^53 + 1: intentionally larger than the robot ID sequence cap (2^53 - 1);
	// this verifies the decoder itself is lossless for any int64 in extra_attrs
	// (e.g. execution/artifact IDs of other vendors), not that robot IDs can reach this value
	largeID := int64(9007199254740993)
	daoTask := &dao.Task{
		ExtraAttrs: fmt.Sprintf(`{"robot_id": %d, "artifact_id": %d}`, largeID, largeID),
	}

	tsk := &Task{}
	tsk.From(daoTask)

	assert.NotNil(t, tsk.ExtraAttrs)

	// Verify UseNumber preserved exact json.Number
	robotVal, ok := tsk.ExtraAttrs["robot_id"].(json.Number)
	assert.True(t, ok)
	assert.Equal(t, "9007199254740993", robotVal.String())

	// Verify GetInt64FromExtraAttrs extracts exact int64
	assert.Equal(t, largeID, tsk.GetInt64FromExtraAttrs("robot_id"))
	assert.Equal(t, largeID, tsk.GetInt64FromExtraAttrs("artifact_id"))
}

func TestInt64FromAny(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		wantVal  int64
		wantOk   bool
	}{
		{"int64", int64(12345), 12345, true},
		{"int", int(12345), 12345, true},
		{"int32", int32(12345), 12345, true},
		{"json.Number valid", json.Number("9007199254740993"), 9007199254740993, true},
		{"json.Number invalid", json.Number("abc"), 0, false},
		{"string valid", "12345", 12345, true},
		{"string invalid", "invalid", 0, false},
		{"float64 exact int", float64(12345.0), 12345, true},
		{"float64 fractional", float64(12.34), 0, false},
		{"float64 NaN", math.NaN(), 0, false},
		{"float64 +Inf", math.Inf(1), 0, false},
		{"float64 -Inf", math.Inf(-1), 0, false},
		{"float64 out of range positive", 1e20, 0, false},
		{"float64 out of range negative", -1e20, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOk := Int64FromAny(tt.input)
			assert.Equal(t, tt.wantOk, gotOk)
			assert.Equal(t, tt.wantVal, gotVal)
		})
	}
}
