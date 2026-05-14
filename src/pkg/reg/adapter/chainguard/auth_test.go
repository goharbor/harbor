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

package chainguard

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func TestLoadCredentials_AccessKeyAndSecret(t *testing.T) {
	reg := &model.Registry{
		Credential: &model.Credential{
			AccessKey:    "my-identity",
			AccessSecret: "raw-jwt-token",
		},
	}
	id, tok, err := loadCredentials(reg)
	require.NoError(t, err)
	require.Equal(t, "my-identity", id)
	require.Equal(t, "raw-jwt-token", tok)
}

func TestLoadCredentials_FilePathSecret(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "token")
	require.NoError(t, os.WriteFile(p, []byte("file-jwt\n"), 0o600))

	reg := &model.Registry{
		Credential: &model.Credential{
			AccessKey:    "id-1",
			AccessSecret: p,
		},
	}
	id, tok, err := loadCredentials(reg)
	require.NoError(t, err)
	require.Equal(t, "id-1", id)
	require.Equal(t, "file-jwt", tok)
}

func TestLoadCredentials_MissingIdentity(t *testing.T) {
	reg := &model.Registry{
		Credential: &model.Credential{
			AccessKey:    "",
			AccessSecret: "tok",
		},
	}
	_, _, err := loadCredentials(reg)
	require.Error(t, err)
}

func TestLoadCredentials_MissingSecret(t *testing.T) {
	reg := &model.Registry{
		Credential: &model.Credential{
			AccessKey:    "id",
			AccessSecret: "",
		},
	}
	_, _, err := loadCredentials(reg)
	require.Error(t, err)
}

func TestTokenExpiry(t *testing.T) {
	payload := "eyJleHAiOjIwMDAwMDAwMDB9"
	jwt := "x." + payload + ".y"
	exp := tokenExpiry(jwt)
	require.False(t, exp.IsZero())
}
