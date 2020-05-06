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
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
)

func upgradeData(ctx context.Context) error {
	abstractor := art.NewAbstractor()
	pros, err := project.Mgr.List()
	if err != nil {
		return err
	}
	for _, pro := range pros {
		repos, err := repository.Mgr.List(ctx, &q.Query{
			Keywords: map[string]interface{}{
				"ProjectID": pro.ProjectID,
			},
		})
		if err != nil {
			log.Errorf("failed to list repositories under the project %s: %v, skip", pro.Name, err)
			continue
		}
		for _, repo := range repos {
			log.Debugf("abstracting artifact metadata under repository %s ....", repo.Name)
			arts, err := artifact.Mgr.List(ctx, &q.Query{
				Keywords: map[string]interface{}{
					"RepositoryID": repo.RepositoryID,
				},
			})
			if err != nil {
				log.Errorf("failed to list artifacts under the repository %s: %v, skip", repo.Name, err)
				continue
			}
			for _, art := range arts {
				if err = abstract(ctx, abstractor, art); err != nil {
					log.Errorf("failed to abstract the artifact %s@%s: %v, skip", art.RepositoryName, art.Digest, err)
					continue
				}
				if err = artifact.Mgr.Update(ctx, art); err != nil {
					log.Errorf("failed to update the artifact %s@%s: %v, skip", repo.Name, art.Digest, err)
					continue
				}
			}
			log.Debugf("artifact metadata under repository %s abstracted", repo.Name)
		}
	}

	// update data version
	return setDataVersion(ctx, 30)
}

func abstract(ctx context.Context, abstractor art.Abstractor, art *artifact.Artifact) error {
	// abstract the children
	for _, reference := range art.References {
		child, err := artifact.Mgr.Get(ctx, reference.ChildID)
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
