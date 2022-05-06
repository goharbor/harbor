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

package migration

import (
	"context"

	art "github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

func abstractArtData(ctx context.Context) error {
	abstractor := art.NewAbstractor()
	pros, err := pkg.ProjectMgr.List(ctx, nil)
	if err != nil {
		return err
	}
	for _, pro := range pros {
		repos, err := pkg.RepositoryMgr.List(ctx, &q.Query{
			Keywords: map[string]interface{}{
				"ProjectID": pro.ProjectID,
			},
		})
		if err != nil {
			log.Errorf("failed to list repositories under the project %s: %v, skip", pro.Name, err)
			continue
		}
		for _, repo := range repos {
			log.Infof("abstracting artifact metadata under repository %s ....", repo.Name)
			arts, err := pkg.ArtifactMgr.List(ctx, &q.Query{
				Keywords: map[string]interface{}{
					"RepositoryID": repo.RepositoryID,
				},
			})
			if err != nil {
				log.Errorf("failed to list artifacts under the repository %s: %v, skip", repo.Name, err)
				continue
			}
			for _, a := range arts {
				if err = abstract(ctx, abstractor, a); err != nil {
					log.Errorf("failed to abstract the artifact %s@%s: %v, skip", a.RepositoryName, a.Digest, err)
					continue
				}
				if err = pkg.ArtifactMgr.Update(ctx, a); err != nil {
					log.Errorf("failed to update the artifact %s@%s: %v, skip", repo.Name, a.Digest, err)
					continue
				}
			}
			log.Infof("artifact metadata under repository %s abstracted", repo.Name)
		}
	}

	// update data version
	return setDataVersion(ctx, dataversionV2_0_0)
}

func abstract(ctx context.Context, abstractor art.Abstractor, art *artifact.Artifact) error {
	// abstract the children
	for _, reference := range art.References {
		child, err := pkg.ArtifactMgr.Get(ctx, reference.ChildID)
		if err != nil {
			log.Errorf("failed to get the artifact %d: %v, skip", reference.ChildID, err)
			continue
		}
		if err = abstract(ctx, abstractor, child); err != nil {
			log.Errorf("failed to abstract the artifact %s@%s: %v, skip", child.RepositoryName, child.Digest, err)
			continue
		}
	}
	// abstract the parent
	return abstractor.AbstractMetadata(ctx, art)
}
