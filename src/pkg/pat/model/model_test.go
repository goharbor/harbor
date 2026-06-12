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
	"time"

	"github.com/stretchr/testify/require"
)

func TestPersonalAccessTokenTableName(t *testing.T) {
	pat := &PersonalAccessToken{}
	require.Equal(t, "personal_access_token", pat.TableName())
}

func TestPersonalAccessTokenCreation(t *testing.T) {
	now := time.Now()
	pat := &PersonalAccessToken{
		ID:          1,
		UserID:      10,
		Name:        "test-token",
		Secret:      "hashed_secret",
		Salt:        "salt_value",
		Description: "Test token",
		ExpiresAt:   now.AddDate(0, 0, 30).Unix(),
		LastUsedAt:  now.Unix(),
		Disabled:    false,
		IsLegacy:    false,
		CreationTime: now,
		UpdateTime:   now,
	}

	require.Equal(t, int64(1), pat.ID)
	require.Equal(t, 10, pat.UserID)
	require.Equal(t, "test-token", pat.Name)
	require.Equal(t, "Test token", pat.Description)
	require.False(t, pat.Disabled)
	require.False(t, pat.IsLegacy)
	require.True(t, pat.ExpiresAt > 0)
}

func TestPersonalAccessTokenLegacy(t *testing.T) {
	pat := &PersonalAccessToken{
		ID:       2,
		UserID:   20,
		Name:     "cli-secret",
		IsLegacy: true,
		ExpiresAt: -1,
	}

	require.True(t, pat.IsLegacy)
	require.Equal(t, int64(-1), pat.ExpiresAt)
}

func TestPersonalAccessTokenDisabled(t *testing.T) {
	pat := &PersonalAccessToken{
		ID:       3,
		UserID:   30,
		Name:     "disabled-token",
		Disabled: true,
	}

	require.True(t, pat.Disabled)
}
