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

package model

import (
	"testing"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/stretchr/testify/assert"
)

func TestIsFinalStatus(t *testing.T) {
	assert := assert.New(t)
	flag := IsFinalStatus("")
	assert.False(flag)

	flag = IsFinalStatus(job.PendingStatus.String())
	assert.False(flag)

	flag = IsFinalStatus(job.RunningStatus.String())
	assert.False(flag)

	flag = IsFinalStatus(job.StoppedStatus.String())
	assert.True(flag)

	flag = IsFinalStatus(job.ErrorStatus.String())
	assert.True(flag)

	flag = IsFinalStatus(job.SuccessStatus.String())
	assert.True(flag)
}
