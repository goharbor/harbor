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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/errors"
)

func TestParseFileFilter(t *testing.T) {
	f, err := ParseFileFilter("*.json, *.safetensors\n!onnx/** !*.bin")
	require.NoError(t, err)
	assert.Equal(t, &FileFilter{Include: []string{"*.json", "*.safetensors"}, Exclude: []string{"onnx/**", "*.bin"}}, f)

	var value any
	require.NoError(t, json.Unmarshal([]byte(`{"include":["*.json"],"exclude":["x/**"]}`), &value))
	f, err = ParseFileFilter(value)
	require.NoError(t, err)
	assert.Equal(t, &FileFilter{Include: []string{"*.json"}, Exclude: []string{"x/**"}}, f)

	for _, empty := range []any{nil, "", " , ", map[string]any{}, (*FileFilter)(nil), FileFilter{}} {
		f, err := ParseFileFilter(empty)
		require.NoError(t, err, "%#v", empty)
		assert.Nil(t, f, "%#v", empty)
	}

	for _, bad := range []any{"[a", "/abs", 42, []any{"*.json"}, map[string]any{"includes": []any{"*.json"}}, map[string]any{"include": "*.json"}} {
		_, err := ParseFileFilter(bad)
		require.Error(t, err, "%#v", bad)
		assert.True(t, errors.IsErr(err, errors.BadRequestCode), "%#v", bad)
	}
}

func TestFileFilterMatch(t *testing.T) {
	var none *FileFilter
	assert.True(t, none.Match("anything"))

	f := &FileFilter{Include: []string{"*.json", "onnx/**"}, Exclude: []string{"onnx/*.bin"}}
	assert.True(t, f.Match("config.json"))
	assert.True(t, f.Match("onnx/config.json"), "a pattern without slash matches the base name at any depth")
	assert.True(t, f.Match("onnx/model.onnx"))
	assert.False(t, f.Match("onnx/model.bin"))
	assert.False(t, f.Match("model.safetensors"))

	excludeOnly := &FileFilter{Exclude: []string{"*.bin", "original/**"}}
	assert.True(t, excludeOnly.Match("model.safetensors"))
	assert.False(t, excludeOnly.Match("pytorch_model.bin"))
	assert.False(t, excludeOnly.Match("original/consolidated.pth"))
}

func TestFileFilterKey(t *testing.T) {
	var none *FileFilter
	assert.Empty(t, none.Key())
	assert.Empty(t, (&FileFilter{}).Key())
	a := &FileFilter{Include: []string{"b", "a", "a"}}
	b := &FileFilter{Include: []string{"a", "b"}}
	assert.Equal(t, a.Key(), b.Key())
	assert.NotEqual(t, a.Key(), (&FileFilter{Exclude: []string{"a", "b"}}).Key())
	assert.Len(t, a.Key(), 16)
}

func TestFilterValidateFile(t *testing.T) {
	require.NoError(t, (&Filter{Type: FilterTypeFile, Value: "*.json"}).Validate())
	require.NoError(t, (&Filter{Type: FilterTypeFile, Value: map[string]any{"exclude": []any{"*.bin"}}}).Validate())
	require.Error(t, (&Filter{Type: FilterTypeFile, Value: "[a"}).Validate())
	require.Error(t, (&Filter{Type: FilterTypeFile, Value: "*.json", Decoration: Excludes}).Validate())
}

func TestSourceOptionsIsEmpty(t *testing.T) {
	var none *SourceOptions
	assert.True(t, none.IsEmpty())
	assert.True(t, (&SourceOptions{FileFilter: &FileFilter{}}).IsEmpty())
	assert.False(t, (&SourceOptions{FileFilter: &FileFilter{Include: []string{"*"}}}).IsEmpty())
}
