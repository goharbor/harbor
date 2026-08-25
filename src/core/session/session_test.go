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
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/beego/beego/v2/server/web/session"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/config"
	_ "github.com/goharbor/harbor/src/pkg/config/db"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
)

type sessionTestSuite struct {
	suite.Suite

	provider session.Provider
}

func (s *sessionTestSuite) SetupSuite() {
	config.Init()

	var err error
	s.provider, err = session.GetProvider("harbor")
	s.NoError(err, "should get harbor provider")
	s.NotNil(s.provider, "provider should not nil")

	err = s.provider.SessionInit(context.Background(), 3600, "redis://127.0.0.1:6379/0")
	s.NoError(err, "session init should not error")
}

func (s *sessionTestSuite) TestSessionRead() {
	store, err := s.provider.SessionRead(context.Background(), "session-001")
	s.NoError(err, "session read should not error")
	s.NotNil(store)
}

func (s *sessionTestSuite) TestSessionExist() {
	// prepare session
	ctx := context.Background()
	store, err := s.provider.SessionRead(ctx, "session-001")
	s.NoError(err, "session read should not error")
	s.NotNil(store)
	store.SessionRelease(context.Background(), nil)

	defer func() {
		// clean session
		err = s.provider.SessionDestroy(ctx, "session-001")
		s.NoError(err)
	}()

	exist, _ := s.provider.SessionExist(ctx, "session-001")
	s.True(exist, "session-001 should exist")

	exist, _ = s.provider.SessionExist(ctx, "session-002")
	s.False(exist, "session-002 should not exist")
}

func (s *sessionTestSuite) TestSessionRegenerate() {
	// prepare session
	ctx := context.Background()
	store, err := s.provider.SessionRead(ctx, "session-001")
	s.NoError(err, "session read should not error")
	s.NotNil(store)
	store.SessionRelease(ctx, nil)

	defer func() {
		// clean session
		err = s.provider.SessionDestroy(ctx, "session-001")
		s.NoError(err)

		err = s.provider.SessionDestroy(ctx, "session-003")
		s.NoError(err)

		err = s.provider.SessionDestroy(ctx, "session-004")
		s.NoError(err)
	}()

	_, err = s.provider.SessionRegenerate(ctx, "session-001", "session-003")
	s.NoError(err, "session regenerate should not error")

	s.True(s.provider.SessionExist(ctx, "session-003"))
	s.False(s.provider.SessionExist(ctx, "session-001"))

	// session-001 was renamed above, so it no longer exists. Regenerating from a
	// missing session must not leave an empty one behind under the new sid.
	store, err = s.provider.SessionRegenerate(ctx, "session-001", "session-004")
	s.NoError(err, "session regenerate should not error")
	s.NotNil(store, "the caller should still get a usable store")

	exist, _ := s.provider.SessionExist(ctx, "session-004")
	s.False(exist, "an empty session should not be persisted")
}

func (s *sessionTestSuite) TestSessionRegenerateConcurrent() {
	ctx := context.Background()
	const workers = 20

	store, err := s.provider.SessionRead(ctx, "session-concurrent")
	s.NoError(err, "session read should not error")
	s.NoError(store.Set(ctx, "user", "alice"))
	store.SessionRelease(ctx, nil)

	sids := make([]string, workers)
	for i := range sids {
		sids[i] = fmt.Sprintf("session-concurrent-%d", i)
	}

	defer func() {
		s.NoError(s.provider.SessionDestroy(ctx, "session-concurrent"))
		for _, sid := range sids {
			s.NoError(s.provider.SessionDestroy(ctx, sid))
		}
	}()

	// Every worker regenerates the same session, which is what the UI does when
	// several requests are in flight at once.
	errs := make([]error, workers)
	var wg sync.WaitGroup
	for i, sid := range sids {
		wg.Add(1)
		go func(i int, sid string) {
			defer wg.Done()
			_, errs[i] = s.provider.SessionRegenerate(ctx, "session-concurrent", sid)
		}(i, sid)
	}
	wg.Wait()

	for _, err := range errs {
		s.NoError(err, "session regenerate should not error")
	}

	// Only one worker can win the rename. The others must not persist anything:
	// whoever receives one of those sids in a cookie would otherwise be logged out.
	persisted := 0
	for _, sid := range sids {
		exist, _ := s.provider.SessionExist(ctx, sid)
		if !exist {
			continue
		}
		persisted++

		regenerated, err := s.provider.SessionRead(ctx, sid)
		s.NoError(err, "session read should not error")
		s.Equal("alice", regenerated.Get(ctx, "user"), "the regenerated session should keep its data")
	}
	s.Equal(1, persisted, "exactly one regenerated session should be persisted")
}

func (s *sessionTestSuite) TestSessionDestroy() {
	// prepare session
	ctx := context.Background()
	store, err := s.provider.SessionRead(ctx, "session-004")
	s.NoError(err, "session read should not error")
	s.NotNil(store)
	store.SessionRelease(ctx, nil)
	isExist, _ := s.provider.SessionExist(ctx, "session-004")
	s.True(isExist, "session-004 should exist")

	err = s.provider.SessionDestroy(ctx, "session-004")
	s.NoError(err, "session destroy should not error")
	isExist, _ = s.provider.SessionExist(ctx, "session-004")
	s.False(isExist, "session-004 should not exist")
}

func (s *sessionTestSuite) TestSessionGC() {
	s.provider.SessionGC(context.Background())
}

func (s *sessionTestSuite) TestSessionAll() {
	c := s.provider.SessionAll(context.Background())
	s.Equal(0, c)
}

func TestSession(t *testing.T) {
	suite.Run(t, &sessionTestSuite{})
}
