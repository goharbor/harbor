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

package registry

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	beegocontext "github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/require"

	commondao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/orm"
	artdao "github.com/goharbor/harbor/src/pkg/artifact/dao"
	repodao "github.com/goharbor/harbor/src/pkg/repository/dao"
	"github.com/goharbor/harbor/src/pkg/repository/model"
	tagdao "github.com/goharbor/harbor/src/pkg/tag/dao"
	tagmodel "github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"github.com/goharbor/harbor/src/server/router"
)

var tagListDatabase sync.Once

func tagListFixture(tb testing.TB, names []string) context.Context {
	tb.Helper()
	tagListDatabase.Do(commondao.PrepareTestForPostgresSQL)
	ctx := orm.Context()
	repos := repodao.New()
	repo := &model.RepoRecord{ProjectID: 1, Name: "library/tag-list-test"}
	repoID, err := repos.Create(ctx, repo)
	require.NoError(tb, err)
	tb.Cleanup(func() { require.NoError(tb, repos.Delete(ctx, repoID)) })
	arts := artdao.New()
	artID, err := arts.Create(ctx, &artdao.Artifact{
		ProjectID: 1, RepositoryID: repoID, Digest: "sha256:tag-list-test", Type: "IMAGE",
		MediaType: "application/vnd.oci.image.config.v1+json", ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
	})
	require.NoError(tb, err)
	tb.Cleanup(func() { require.NoError(tb, arts.Delete(ctx, artID)) })
	tags := tagdao.New()
	tb.Cleanup(func() { require.NoError(tb, tags.DeleteOfArtifact(ctx, artID)) })
	if len(names) > 0 {
		rows := make([]*tagmodel.Tag, len(names))
		for i, name := range names {
			rows[i] = &tagmodel.Tag{RepositoryID: repoID, ArtifactID: artID, Name: name}
		}
		o, err := orm.FromContext(ctx)
		require.NoError(tb, err)
		_, err = o.InsertMulti(100, rows)
		require.NoError(tb, err)
	}
	input := &beegocontext.BeegoInput{}
	input.SetParam(":splat", repo.Name)
	return context.WithValue(ctx, router.ContextKeyInput{}, input)
}

// BenchmarkTagList measures the local registry handler with PostgreSQL, including
// repository lookup, tag retrieval and response encoding. It requires the same
// POSTGRESQL_* and POSTGRES_MIGRATION_SCRIPTS_PATH settings as the DAO tests.
func BenchmarkTagList(b *testing.B) {
	for _, total := range []int{2, 10000} {
		b.Run(fmt.Sprintf("tags=%d", total), func(b *testing.B) {
			names := make([]string, total)
			for i := range names {
				names[i] = fmt.Sprintf("v%d", i+1)
			}
			ctx := tagListFixture(b, names)
			h := newTagHandler()
			req := httptest.NewRequest(http.MethodGet, "/v2/library/tag-list-test/tags/list?n=100", nil).WithContext(ctx)
			b.ReportAllocs()
			for b.Loop() {
				w := httptest.NewRecorder()
				h.ServeHTTP(w, req)
				if w.Code != http.StatusOK {
					b.Fatalf("status %d: %s", w.Code, w.Body.String())
				}
			}
		})
	}
}
