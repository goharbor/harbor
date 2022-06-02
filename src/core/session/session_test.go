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

package session

import (
	"testing"

	"github.com/beego/beego/session"
	"github.com/stretchr/testify/suite"
)

type sessionTestSuite struct {
	suite.Suite

	provider session.Provider
}

func (s *sessionTestSuite) SetupTest() {
	var err error
	s.provider, err = session.GetProvider("harbor")
	s.NoError(err, "should get harbor provider")
	s.NotNil(s.provider, "provider should not nil")

	err = s.provider.SessionInit(3600, "redis://127.0.0.1:6379/0")
	s.NoError(err, "session init should not error")
}

func (s *sessionTestSuite) TestSessionRead() {
	store, err := s.provider.SessionRead("session-001")
	s.NoError(err, "session read should not error")
	s.NotNil(store)
}

func (s *sessionTestSuite) TestSessionExist() {
	// prepare session
	store, err := s.provider.SessionRead("session-001")
	s.NoError(err, "session read should not error")
	s.NotNil(store)
	store.SessionRelease(nil)

	defer func() {
		// clean session
		err = s.provider.SessionDestroy("session-001")
		s.NoError(err)
	}()

	exist := s.provider.SessionExist("session-001")
	s.True(exist, "session-001 should exist")

	exist = s.provider.SessionExist("session-002")
	s.False(exist, "session-002 should not exist")
}

func (s *sessionTestSuite) TestSessionRegenerate() {
	// prepare session
	store, err := s.provider.SessionRead("session-001")
	s.NoError(err, "session read should not error")
	s.NotNil(store)
	store.SessionRelease(nil)

	defer func() {
		// clean session
		err = s.provider.SessionDestroy("session-001")
		s.NoError(err)

		err = s.provider.SessionDestroy("session-003")
		s.NoError(err)
	}()

	_, err = s.provider.SessionRegenerate("session-001", "session-003")
	s.NoError(err, "session regenerate should not error")

	s.True(s.provider.SessionExist("session-003"))
	s.False(s.provider.SessionExist("session-001"))
}

func (s *sessionTestSuite) TestSessionDestroy() {
	// prepare session
	store, err := s.provider.SessionRead("session-004")
	s.NoError(err, "session read should not error")
	s.NotNil(store)
	store.SessionRelease(nil)
	s.True(s.provider.SessionExist("session-004"), "session-004 should exist")

	err = s.provider.SessionDestroy("session-004")
	s.NoError(err, "session destroy should not error")
	s.False(s.provider.SessionExist("session-004"), "session-004 should not exist")
}

func (s *sessionTestSuite) TestSessionGC() {
	s.provider.SessionGC()
}

func (s *sessionTestSuite) TestSessionAll() {
	c := s.provider.SessionAll()
	s.Equal(0, c)
}

func TestSession(t *testing.T) {
	suite.Run(t, &sessionTestSuite{})
}
