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

package reg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/encrypt"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func TestHuggingFaceCredentialRoundTrip(t *testing.T) {
	config.InitWithSettings(nil, &encrypt.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"})

	in := &model.Registry{
		ID:   1,
		Type: model.RegistryTypeHuggingFace,
		URL:  "https://huggingface.co",
		Credential: &model.Credential{
			Type:         model.CredentialTypeBasic,
			AccessKey:    "hf-user",
			AccessSecret: "hf_token",
		},
	}
	stored, err := toDaoModel(in)
	require.NoError(t, err)
	assert.Equal(t, "hf-user", stored.AccessKey)
	assert.NotEqual(t, "hf_token", stored.AccessSecret, "the token is encrypted at rest")

	out, err := fromDaoModel(stored)
	require.NoError(t, err)
	require.NotNil(t, out.Credential)
	assert.Equal(t, "hf-user", out.Credential.AccessKey)
	assert.Equal(t, "hf_token", out.Credential.AccessSecret)

	// Without an access key the manager drops the secret, which is why the Hugging Face adapter
	// rejects a secret-only credential when the registry is created or updated.
	in.Credential.AccessKey = ""
	stored, err = toDaoModel(in)
	require.NoError(t, err)
	assert.Empty(t, stored.AccessSecret)
	out, err = fromDaoModel(stored)
	require.NoError(t, err)
	assert.Equal(t, &model.Credential{}, out.Credential)
}
