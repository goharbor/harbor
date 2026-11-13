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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	regmodel "github.com/goharbor/harbor/src/pkg/reg/model"
)

func TestFileFilter(t *testing.T) {
	p := &Policy{}
	f, err := p.GetFileFilter()
	require.NoError(t, err)
	assert.Nil(t, f)

	require.NoError(t, p.SetFileFilter(&regmodel.FileFilter{Include: []string{"*.safetensors", "*.json"}, Exclude: []string{"onnx/**"}}))
	assert.Equal(t, `{"include":["*.safetensors","*.json"],"exclude":["onnx/**"]}`, p.FileFilter)
	f, err = p.GetFileFilter()
	require.NoError(t, err)
	assert.Equal(t, []string{"*.safetensors", "*.json"}, f.Include)
	assert.Equal(t, []string{"onnx/**"}, f.Exclude)

	require.NoError(t, p.SetFileFilter(nil))
	assert.Equal(t, "", p.FileFilter)
	require.NoError(t, p.SetFileFilter(&regmodel.FileFilter{}))
	assert.Equal(t, "", p.FileFilter)

	p.FileFilter = "not-json"
	_, err = p.GetFileFilter()
	assert.Error(t, err)
}

func TestDefaultDestRepository(t *testing.T) {
	assert.Equal(t, "qwen/qwen3-8b", DefaultDestRepository(" Qwen/Qwen3-8B/ "))
	assert.Equal(t, "gpt2", DefaultDestRepository("gpt2"))
}
