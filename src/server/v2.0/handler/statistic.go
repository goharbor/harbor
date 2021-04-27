// Copyright 2018 Project Harbor Authors
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

package handler

import (
	"context"
	"github.com/goharbor/harbor/src/controller/blob"

	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/statistic"
)

func newStatisticAPI() *statisticAPI {
	return &statisticAPI{
		proCtl:  project.Ctl,
		repoCtl: repository.Ctl,
		blobCtl: blob.Ctl,
	}
}

type statisticAPI struct {
	BaseAPI
	proCtl  project.Controller
	repoCtl repository.Controller
	blobCtl blob.Controller
}

func (s *statisticAPI) GetStatistic(ctx context.Context, params operation.GetStatisticParams) middleware.Responder {
	if err := s.RequireAuthenticated(ctx); err != nil {
		return s.SendError(ctx, err)
	}

	statistic := &models.Statistic{}
	pubProjs, err := s.proCtl.List(ctx, q.New(q.KeyWords{"public": true}), project.Metadata(false))
	if err != nil {
		return s.SendError(ctx, err)
	}

	statistic.PublicProjectCount = (int64)(len(pubProjs))
	if len(pubProjs) == 0 {
		statistic.PublicRepoCount = 0
	} else {
		var ids []interface{}
		for _, p := range pubProjs {
			ids = append(ids, p.ProjectID)
		}
		n, err := s.repoCtl.Count(ctx, &q.Query{
			Keywords: map[string]interface{}{
				"ProjectID": q.NewOrList(ids),
			},
		})
		if err != nil {
			return s.SendError(ctx, err)
		}
		statistic.PublicRepoCount = n
	}

	securityCtx, err := s.GetSecurityContext(ctx)
	if err != nil {
		return s.SendError(ctx, err)
	}

	if securityCtx.IsSysAdmin() {
		count, err := s.proCtl.Count(ctx, nil)
		if err != nil {
			return s.SendError(ctx, err)
		}
		statistic.TotalProjectCount = count
		statistic.PrivateProjectCount = count - statistic.PublicProjectCount

		n, err := s.repoCtl.Count(ctx, nil)
		if err != nil {
			return s.SendError(ctx, err)
		}
		statistic.TotalRepoCount = n
		statistic.PrivateRepoCount = n - statistic.PublicRepoCount

		sum, err := s.blobCtl.CalculateTotalSize(ctx, true)
		if err != nil {
			return s.SendError(ctx, err)
		}
		statistic.TotalStorageConsumption = sum

	} else {
		var privProjectIDs []interface{}
		if sc, ok := securityCtx.(*local.SecurityContext); ok && sc.IsAuthenticated() {
			user := sc.User()
			member := &project.MemberQuery{
				UserID:   user.UserID,
				GroupIDs: user.GroupIDs,
			}

			myProjects, err := s.proCtl.List(ctx, q.New(q.KeyWords{"member": member, "public": false}), project.Metadata(false))
			if err != nil {
				return s.SendError(ctx, err)
			}
			for _, p := range myProjects {
				privProjectIDs = append(privProjectIDs, p.ProjectID)
			}
		}

		statistic.PrivateProjectCount = int64(len(privProjectIDs))
		if statistic.PrivateProjectCount == 0 {
			statistic.PrivateRepoCount = 0
		} else {
			n, err := s.repoCtl.Count(ctx, &q.Query{
				Keywords: map[string]interface{}{
					"ProjectID": q.NewOrList(privProjectIDs),
				},
			})
			if err != nil {
				return s.SendError(ctx, err)
			}
			statistic.PrivateRepoCount = n
		}
	}

	return operation.NewGetStatisticOK().WithPayload(statistic)
}
