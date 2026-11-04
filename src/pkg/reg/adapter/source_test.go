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

package adapter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/reg/model"
)

type configurable struct{ got *model.SourceOptions }

func (c *configurable) ConfigureSource(opts *model.SourceOptions) error {
	c.got = opts
	return nil
}

func TestConfigureSource(t *testing.T) {
	opts := &model.SourceOptions{FileFilter: &model.FileFilter{Include: []string{"*.json"}}}

	require.NoError(t, ConfigureSource(struct{}{}, nil), "empty options need no support")
	require.NoError(t, ConfigureSource(struct{}{}, &model.SourceOptions{}))
	require.Error(t, ConfigureSource(struct{}{}, opts), "non-empty options are never silently dropped")

	c := &configurable{}
	require.NoError(t, ConfigureSource(c, opts))
	assert.Same(t, opts, c.got)
}
