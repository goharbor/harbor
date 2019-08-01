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

package scheduler

import (
	"testing"

	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/goharbor/harbor/src/testing/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var sch *scheduler

type schedulerTestSuite struct {
	suite.Suite
}

func (s *schedulerTestSuite) SetupTest() {
	t := s.T()
	// empty callback function registry before running every test case
	// and register a new callback function named "callback"
	registry = make(map[string]CallbackFunc)
	err := Register("callback", func(interface{}) error { return nil })
	require.Nil(t, err)

	// recreate the scheduler object
	sch = &scheduler{
		jobserviceClient: &job.MockJobClient{},
		manager:          &htesting.FakeSchedulerManager{},
	}
}

func (s *schedulerTestSuite) TestRegister() {
	t := s.T()
	var name string
	var callbackFun CallbackFunc

	// empty name
	err := Register(name, callbackFun)
	require.NotNil(t, err)

	// nil callback function
	name = "test"
	err = Register(name, callbackFun)
	require.NotNil(t, err)

	// pass
	callbackFun = func(interface{}) error { return nil }
	err = Register(name, callbackFun)
	require.Nil(t, err)

	// duplicate name
	err = Register(name, callbackFun)
	require.NotNil(t, err)
}

func (s *schedulerTestSuite) TestGetCallbackFunc() {
	t := s.T()
	// not exist
	_, err := GetCallbackFunc("not-exist")
	require.NotNil(t, err)

	// pass
	f, err := GetCallbackFunc("callback")
	require.Nil(t, err)
	assert.NotNil(t, f)
}

func (s *schedulerTestSuite) TestSchedule() {
	t := s.T()

	// callback function not exist
	_, err := sch.Schedule("0 * * * * *", "not-exist", nil)
	require.NotNil(t, err)

	// pass
	id, err := sch.Schedule("0 * * * * *", "callback", nil)
	require.Nil(t, err)
	assert.Equal(t, int64(1), id)
}

func (s *schedulerTestSuite) TestUnSchedule() {
	t := s.T()
	// schedule not exist
	err := sch.UnSchedule(1)
	require.NotNil(t, err)

	// schedule exist
	id, err := sch.Schedule("0 * * * * *", "callback", nil)
	require.Nil(t, err)
	assert.Equal(t, int64(1), id)

	err = sch.UnSchedule(id)
	require.Nil(t, err)
}

func TestScheduler(t *testing.T) {
	s := &schedulerTestSuite{}
	suite.Run(t, s)
}
