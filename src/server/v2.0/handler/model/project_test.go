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
	t.Run("legacy none with unset prevent flags fills defaults", func(t *testing.T) {
		severity := "none"
		md := &models.ProjectMetadata{
			Severity: &severity,
		}

		NormalizeLegacySeverityPolicy(md)

		assert.NotNil(t, md.Severity)
		assert.Equal(t, "low", *md.Severity)
		assert.NotNil(t, md.PreventVul)
		assert.Equal(t, "true", *md.PreventVul)
		assert.NotNil(t, md.PreventUnscanned)
		assert.Equal(t, "true", *md.PreventUnscanned)
	})

	t.Run("legacy none does not override explicit prevent flags", func(t *testing.T) {
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
		assert.Equal(t, "false", *md.PreventVul)
		assert.NotNil(t, md.PreventUnscanned)
		assert.Equal(t, "false", *md.PreventUnscanned)
	})

	t.Run("mixed explicit and unset prevent flags are handled independently", func(t *testing.T) {
		severity := "none"
		preventVul := "false"
		md := &models.ProjectMetadata{
			Severity:   &severity,
			PreventVul: &preventVul,
			// PreventUnscanned intentionally left unset
		}

		NormalizeLegacySeverityPolicy(md)

		assert.NotNil(t, md.Severity)
		assert.Equal(t, "low", *md.Severity)
		assert.NotNil(t, md.PreventVul)
		assert.Equal(t, "false", *md.PreventVul, "explicit prevent_vul must be preserved")
		assert.NotNil(t, md.PreventUnscanned)
		assert.Equal(t, "true", *md.PreventUnscanned, "unset prevent_unscanned must be filled with the legacy default")
	})

	t.Run("nil metadata is a no-op", func(t *testing.T) {
		assert.NotPanics(t, func() {
			NormalizeLegacySeverityPolicy(nil)
		})
	})

	t.Run("missing severity is a no-op", func(t *testing.T) {
		preventVul := "false"
		md := &models.ProjectMetadata{
			PreventVul: &preventVul,
		}

		NormalizeLegacySeverityPolicy(md)

		assert.Nil(t, md.Severity)
		assert.Equal(t, "false", *md.PreventVul)
		assert.Nil(t, md.PreventUnscanned)
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
