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
package clair

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"runtime"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
)

func TestParseServerity(t *testing.T) {
	newAssert := assert.New(t)
	in := map[string]models.Severity{
		"negligible": models.SevNone,
		"whatever":   models.SevUnknown,
		"LOW":        models.SevLow,
		"Medium":     models.SevMedium,
		"high":       models.SevHigh,
		"Critical":   models.SevHigh,
	}
	for k, v := range in {
		newAssert.Equal(v, ParseClairSev(k))
	}
}

func TestTransformVuln(t *testing.T) {
	var clairVuln = &models.ClairLayerEnvelope{}
	newAssert := assert.New(t)
	empty := []byte(`{"Layer":{"Features":[]}}`)
	loadVuln(empty, clairVuln)
	output, o := transformVuln(clairVuln)
	newAssert.Equal(0, output.Total)
	newAssert.Equal(models.SevNone, o)
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current directory")
	}
	curDir := path.Dir(f)
	fileData, err := ioutil.ReadFile(path.Join(curDir, "test/total-12.json"))
	if err != nil {
		panic(err)
	}
	loadVuln(fileData, clairVuln)
	output, o = transformVuln(clairVuln)
	newAssert.Equal(12, output.Total)
	newAssert.Equal(models.SevHigh, o)
	hit := false
	for _, s := range output.Summary {
		if s.Sev == int(models.SevHigh) {
			newAssert.Equal(3, s.Count, "There should be 3 components with High severity")
			hit = true
		}
	}
	newAssert.True(hit, "Not found entry for high severity in summary list")
}

func loadVuln(input []byte, data *models.ClairLayerEnvelope) {
	err := json.Unmarshal(input, data)
	if err != nil {
		panic(err)
	}
}
