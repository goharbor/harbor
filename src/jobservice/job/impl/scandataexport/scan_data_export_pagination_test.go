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

package scandataexport

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/jobservice/job"
	artpkg "github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/scan/export"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
	exporttesting "github.com/goharbor/harbor/src/testing/pkg/scan/export"
)

func TestWriteCSVFilePagination(t *testing.T) {
	type page struct {
		group    int
		number   int64
		rawCount int
		rows     []export.Data
	}
	row := func(pkg string) []export.Data {
		return []export.Data{{Repository: "project/repo", CVEId: "CVE-2026-0001", Package: pkg}}
	}
	tests := []struct {
		name          string
		cveID         string
		artifactCount int
		pages         []page
		wantPackages  []string
	}{
		{
			name:  "empty filtered first page",
			cveID: "CVE-2026-0001",
			pages: []page{
				{number: 1, rawCount: export.QueryPageSize},
				{number: 2, rawCount: 1, rows: row("second-page")},
			},
			wantPackages: []string{"second-page"},
		},
		{
			name:  "partially matching first page",
			cveID: "CVE-2026-0001",
			pages: []page{
				{number: 1, rawCount: export.QueryPageSize, rows: row("first-page")},
				{number: 2, rawCount: 1, rows: row("second-page")},
			},
			wantPackages: []string{"first-page", "second-page"},
		},
		{
			name:  "empty filtered intermediate page",
			cveID: "CVE-2026-0001",
			pages: []page{
				{number: 1, rawCount: export.QueryPageSize, rows: row("first-page")},
				{number: 2, rawCount: export.QueryPageSize},
				{number: 3, rawCount: 1, rows: row("third-page")},
			},
			wantPackages: []string{"first-page", "third-page"},
		},
		{
			name:  "short final raw page without matches",
			cveID: "CVE-2026-0001",
			pages: []page{
				{number: 1, rawCount: export.QueryPageSize},
				{number: 2, rawCount: 1},
			},
		},
		{
			name:  "empty raw page",
			cveID: "CVE-2026-0001",
			pages: []page{{number: 1}},
		},
		{
			name:         "unfiltered export",
			pages:        []page{{number: 1, rawCount: 1, rows: row("unfiltered")}},
			wantPackages: []string{"unfiltered"},
		},
		{
			name:          "first matches in later artifact group",
			cveID:         "CVE-2026-0001",
			artifactCount: export.ArtifactGroupSize + 1,
			pages: []page{
				{number: 1, rawCount: 1},
				{group: 1, number: 1, rawCount: 1, rows: row("second-group")},
			},
			wantPackages: []string{"second-group"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := exporttesting.NewManager(t)
			filters := exporttesting.NewFilterProcessor(t)
			count := tt.artifactCount
			if count == 0 {
				count = 1
			}
			arts := make([]*artifact.Artifact, count)
			ids := make([]int64, count)
			for i := range arts {
				ids[i] = int64(i + 1)
				arts[i] = &artifact.Artifact{Artifact: artpkg.Artifact{ID: ids[i]}}
			}
			filters.On("ProcessRepositoryFilter", testifymock.Anything, "", []int64{1}).
				Return([]int64{1}, nil).Once()
			filters.On("ProcessTagFilter", testifymock.Anything, "", []int64{1}).
				Return(arts, nil).Once()
			filters.On("ProcessLabelFilter", testifymock.Anything, testifymock.Anything, arts).
				Return(arts, nil).Once()

			for _, pg := range tt.pages {
				start := pg.group * export.ArtifactGroupSize
				end := min(start+export.ArtifactGroupSize, len(ids))
				wantIDs := ids[start:end]
				matchesPage := testifymock.MatchedBy(func(p export.Params) bool {
					return p.PageNumber == pg.number &&
						p.PageSize == export.QueryPageSize &&
						p.CVEIds == tt.cveID &&
						reflect.DeepEqual(p.ArtifactIDs, wantIDs)
				})
				manager.On("Fetch", testifymock.Anything, matchesPage).
					Return(pg.rows, pg.rawCount, nil).Once()
			}

			exportJob := &ScanDataExport{exportMgr: manager, filterProcessor: filters}
			fileName := filepath.Join(t.TempDir(), "export.csv")
			params := job.Parameters{
				export.JobID: "pagination-test",
				export.JobRequest: map[string]any{
					"Projects": []int64{1},
					"CVEIds":   tt.cveID,
				},
			}
			require.NoError(t, exportJob.writeCsvFile(&mockjobservice.MockJobContext{}, params, fileName))

			file, err := os.Open(fileName)
			require.NoError(t, err)
			defer file.Close()
			records, err := csv.NewReader(file).ReadAll()
			require.NoError(t, err)
			if len(tt.wantPackages) == 0 {
				require.Empty(t, records)
				return
			}

			require.Len(t, records, len(tt.wantPackages)+1)
			require.Equal(t, []string{
				"Repository", "Artifact Digest", "CVE", "Package", "Current Version",
				"Fixed in version", "Severity", "CWE Ids", "Additional Data", "Scanner",
			}, records[0])
			for i, pkg := range tt.wantPackages {
				require.Equal(t, "project/repo", records[i+1][0])
				require.Equal(t, "CVE-2026-0001", records[i+1][2])
				require.Equal(t, pkg, records[i+1][3])
			}
		})
	}
}
