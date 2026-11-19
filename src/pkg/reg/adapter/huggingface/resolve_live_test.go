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

package huggingface

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

// TestResolveLive calls the public Hub, so it only runs when HF_MODEL is set:
//
//	HF_MODEL=HuggingFaceTB/SmolLM2-135M HF_REF=main HF_FILTER='*.json' \
//	  go test -run TestResolveLive -v ./pkg/reg/adapter/huggingface/
func TestResolveLive(t *testing.T) {
	modelID := os.Getenv("HF_MODEL")
	if modelID == "" {
		t.Skip("HF_MODEL is not set")
	}
	filter, err := model.ParseFileFilter(os.Getenv("HF_FILTER"))
	require.NoError(t, err)
	a := newTestAdapter(t, &model.Registry{ID: -1, Type: model.RegistryTypeHuggingFace, URL: hub.DefaultEndpoint}, newMemoryCache())
	r, err := a.Resolve(modelID, os.Getenv("HF_REF"), filter)
	require.NoError(t, err)
	out, err := json.MarshalIndent(r, "", "  ")
	require.NoError(t, err)
	t.Logf("%s", out)
}
