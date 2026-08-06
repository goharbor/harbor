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

	"github.com/goharbor/harbor/src/server/v2.0/models"
)

func TestNormalizeLegacySeverityPolicy(t *testing.T) {
	t.Run("legacy none is normalized", func(t *testing.T) {
		severity := "none"
		preventVul := "false"
		preventUnscanned := "false"
		md := &models.ProjectMetadata{
			Severity:         &severity,
			PreventVul:       &preventVul,
			PreventUnscanned: &preventUnscanned,
		}

		NormalizeLegacySeverityPolicy(md)

		assert.NotNil(t, md.Severity)
		assert.Equal(t, "low", *md.Severity)
		assert.NotNil(t, md.PreventVul)
		assert.Equal(t, "true", *md.PreventVul)
		assert.NotNil(t, md.PreventUnscanned)
		assert.Equal(t, "true", *md.PreventUnscanned)
	})

	t.Run("non legacy severity unchanged", func(t *testing.T) {
		severity := "high"
		preventVul := "false"
		preventUnscanned := "false"
		md := &models.ProjectMetadata{
			Severity:         &severity,
			PreventVul:       &preventVul,
			PreventUnscanned: &preventUnscanned,
		}

		NormalizeLegacySeverityPolicy(md)

		assert.Equal(t, "high", *md.Severity)
		assert.Equal(t, "false", *md.PreventVul)
		assert.Equal(t, "false", *md.PreventUnscanned)
	})
}
