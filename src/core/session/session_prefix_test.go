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
	"testing"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/config"
)

func TestSessionKeyPrefix(t *testing.T) {
	config.Init()
	t.Setenv(keyPrefixEnv, "harbor:session:")

	ctx := context.Background()
	p := &Provider{}
	require.NoError(t, p.SessionInit(ctx, 3600, "redis://127.0.0.1:6379/0"))

	raw := goredis.NewClient(&goredis.Options{Addr: "127.0.0.1:6379"})
	defer raw.Close()
	defer raw.Del(ctx, "harbor:session:prefix-001", "harbor:session:prefix-002", "prefix-001", "prefix-002")

	// SessionRelease writes through the raw client: the key must be prefixed.
	store, err := p.SessionRead(ctx, "prefix-001")
	require.NoError(t, err)
	require.NoError(t, store.Set(ctx, "k", "v"))
	store.SessionRelease(ctx, nil)

	require.EqualValues(t, 1, raw.Exists(ctx, "harbor:session:prefix-001").Val())
	require.EqualValues(t, 0, raw.Exists(ctx, "prefix-001").Val())
	exist, _ := p.SessionExist(ctx, "prefix-001")
	require.True(t, exist)

	// SessionReleaseIfPresent (SetXX) updates the prefixed key when present...
	require.NoError(t, store.Set(ctx, "k2", "v2"))
	store.SessionReleaseIfPresent(ctx, nil)
	store, err = p.SessionRead(ctx, "prefix-001")
	require.NoError(t, err)
	require.Equal(t, "v2", store.Get(ctx, "k2"))
	require.EqualValues(t, 0, raw.Exists(ctx, "prefix-001").Val())

	// ...and writes nothing when the session is absent.
	absent, err := p.SessionRead(ctx, "prefix-002")
	require.NoError(t, err)
	require.NoError(t, absent.Set(ctx, "k", "v"))
	absent.SessionReleaseIfPresent(ctx, nil)
	require.EqualValues(t, 0, raw.Exists(ctx, "harbor:session:prefix-002").Val())
	require.EqualValues(t, 0, raw.Exists(ctx, "prefix-002").Val())

	// SessionRegenerate renames through the raw client as well.
	store, err = p.SessionRegenerate(ctx, "prefix-001", "prefix-002")
	require.NoError(t, err)
	require.Equal(t, "v", store.Get(ctx, "k"))
	require.EqualValues(t, 0, raw.Exists(ctx, "harbor:session:prefix-001").Val())
	require.EqualValues(t, 1, raw.Exists(ctx, "harbor:session:prefix-002").Val())
	require.EqualValues(t, 0, raw.Exists(ctx, "prefix-002").Val())

	require.NoError(t, p.SessionDestroy(ctx, "prefix-002"))
	require.EqualValues(t, 0, raw.Exists(ctx, "harbor:session:prefix-002").Val())
}
