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

package modelimport

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/goharbor/harbor/src/jobservice/job"
	taskpkg "github.com/goharbor/harbor/src/pkg/task"
	mocktask "github.com/goharbor/harbor/src/testing/pkg/task"
)

func TestNormalizeTargetRepository(t *testing.T) {
	tests := []struct {
		name       string
		project    string
		repository string
		want       string
		wantErr    bool
	}{
		{
			name:       "relative repository",
			project:    "library",
			repository: "bge-m3",
			want:       "library/bge-m3",
		},
		{
			name:       "project qualified repository",
			project:    "library",
			repository: "library/bge-m3",
			want:       "library/bge-m3",
		},
		{
			name:       "nested repository",
			project:    "library",
			repository: "library/models/bge-m3",
			want:       "library/models/bge-m3",
		},
		{
			name:       "trim whitespace and slashes",
			project:    "library",
			repository: " /bge-m3/ ",
			want:       "library/bge-m3",
		},
		{
			name:       "reject empty repository",
			project:    "library",
			repository: " / ",
			wantErr:    true,
		},
		{
			name:       "reject another project",
			project:    "library",
			repository: "other/bge-m3",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeTargetRepository(tt.project, tt.repository)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestModelImportCheckIn(t *testing.T) {
	mgr := &mocktask.Manager{}
	original := taskpkg.Mgr
	taskpkg.Mgr = mgr
	defer func() {
		taskpkg.Mgr = original
	}()

	mgr.On("Get", mock.Anything, int64(37)).Return(&taskpkg.Task{
		ID: 37,
		ExtraAttrs: map[string]any{
			"policy_id": int64(6),
		},
	}, nil).Once()
	mgr.On("UpdateExtraAttrs", mock.Anything, int64(37), mock.MatchedBy(func(attrs map[string]any) bool {
		return attrs["policy_id"] == int64(6) &&
			attrs["progress_stage"] == "push" &&
			attrs["progress_message"] == "pushing OCI model artifact" &&
			attrs["commit_sha"] == "abc123" &&
			attrs["artifact_digest"] == "sha256:123" &&
			attrs["files"] == 30
	})).Return(nil).Once()

	err := modelImportCheckIn(context.Background(), &taskpkg.Task{ID: 37}, &job.StatusChange{
		CheckIn: `{"stage":"push","message":"pushing OCI model artifact","commit_sha":"abc123","artifact_digest":"sha256:123","files":30}`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mgr.AssertExpectations(t)
}
