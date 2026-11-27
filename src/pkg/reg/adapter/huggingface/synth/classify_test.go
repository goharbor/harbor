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

package synth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassify(t *testing.T) {
	const big = weightFileSizeThreshold + 1
	cases := []struct {
		path string
		size int64
		want string
	}{
		{"config.json", 1, MediaTypeWeightConfigRaw},
		{"tokenizer.json", big, MediaTypeWeightConfigRaw},
		{"original/params.json", 1, MediaTypeWeightConfigRaw},
		{"package.json", 1, MediaTypeWeightConfigRaw}, // *.json is checked before the code list
		{"tokenizer.model", 1, MediaTypeWeightConfigRaw},
		{"tokenizer.model.v3", 1, MediaTypeWeightConfigRaw},
		{"spiece.model", 1, MediaTypeWeightConfigRaw},
		{"sentencepiece.bpe.model", 1, MediaTypeWeightConfigRaw},
		{"vocab.txt", 1, MediaTypeWeightConfigRaw},
		{"merges.txt", 1, MediaTypeWeightConfigRaw},
		{"x.meta", 1, MediaTypeWeightConfigRaw},
		{"x.xml", 1, MediaTypeWeightConfigRaw},
		{"chat_template.jinja", 1, MediaTypeWeightConfigRaw},
		{"CONFIG.JSON", 1, MediaTypeWeightConfigRaw},
		{"model.safetensors", 1, MediaTypeWeightRaw},
		{"sub/dir/Model-00001-of-00002.SafeTensors", 1, MediaTypeWeightRaw},
		{"llama-q4_k_m.gguf", 1, MediaTypeWeightRaw},
		{"pytorch_model.bin", 1, MediaTypeWeightRaw},
		{"tf_model.h5", 1, MediaTypeWeightRaw},
		{"flax_model.msgpack", 1, MediaTypeWeightRaw},
		{"model.onnx", 1, MediaTypeWeightRaw},
		{"model.onnx_data", 1, MediaTypeWeightRaw},
		{"data.parquet", 1, MediaTypeWeightRaw},
		{"tensor0_1", 1, MediaTypeWeightRaw},
		{"modeling_qwen.py", 1, MediaTypeCodeRaw},
		{"requirements.txt", 1, MediaTypeCodeRaw},
		{"Dockerfile", 1, MediaTypeCodeRaw},
		{"libfoo.so", big, MediaTypeCodeRaw},
		{"README.md", 1, MediaTypeDocRaw},
		{"USE_POLICY.md", 1, MediaTypeDocRaw},
		{"LICENSE", 1, MediaTypeDocRaw},
		{"license", 1, MediaTypeDocRaw},
		{"notes.txt", 1, MediaTypeDocRaw},
		{"figures/arch.png", big, MediaTypeDocRaw},
		{"unknown.xyz", weightFileSizeThreshold, MediaTypeCodeRaw},
		{"unknown.xyz", big, MediaTypeWeightRaw},
		{"noext", 1, MediaTypeCodeRaw},
	}
	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			assert.Equal(t, c.want, classify(c.path, c.size).mediaType())
		})
	}
}

func TestIncluded(t *testing.T) {
	assert.False(t, Included(".gitattributes"))
	assert.True(t, Included("sub/.gitattributes"))
	assert.True(t, Included(".gitignore"))
	assert.True(t, Included("config.json"))
}
