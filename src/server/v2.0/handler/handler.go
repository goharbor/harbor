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

package handler

import (
	"context"
	"log"
	"net/http"

	rmiddleware "github.com/go-openapi/runtime/middleware"

	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/goharbor/harbor/src/server/middleware/blob"
	"github.com/goharbor/harbor/src/server/middleware/metric"
	"github.com/goharbor/harbor/src/server/middleware/quota"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
)

// New returns http handler for API V2.0
func New() http.Handler {
	h, api, err := restapi.HandlerAPI(restapi.Config{
		ArtifactAPI:           newArtifactAPI(),
		RepositoryAPI:         newRepositoryAPI(),
		AuditlogAPI:           newAuditLogAPI(),
		ScannerAPI:            newScannerAPI(),
		ScanAPI:               newScanAPI(),
		ScanAllAPI:            newScanAllAPI(),
		SearchAPI:             newSearchAPI(),
		ProjectAPI:            newProjectAPI(),
		MemberAPI:             newMemberAPI(),
		PreheatAPI:            newPreheatAPI(),
		IconAPI:               newIconAPI(),
		RobotAPI:              newRobotAPI(),
		ReplicationAPI:        newReplicationAPI(),
		RegistryAPI:           newRegistryAPI(),
		SysteminfoAPI:         newSystemInfoAPI(),
		PingAPI:               newPingAPI(),
		LdapAPI:               newLdapAPI(),
		LabelAPI:              newLabelAPI(),
		GCAPI:                 newGCAPI(),
		QuotaAPI:              newQuotaAPI(),
		RetentionAPI:          newRetentionAPI(),
		WebhookAPI:            newWebhookAPI(),
		WebhookjobAPI:         newWebhookJobAPI(),
		ImmutableAPI:          newImmutableAPI(),
		OIDCAPI:               newOIDCAPI(),
		SystemCVEAllowlistAPI: newSystemCVEAllowListAPI(),
		ConfigureAPI:          newConfigAPI(),
		UsergroupAPI:          newUserGroupAPI(),
		UserAPI:               newUsersAPI(),
		HealthAPI:             newHealthAPI(),
		StatisticAPI:          newStatisticAPI(),
		ProjectMetadataAPI:    newProjectMetadaAPI(),
		PurgeAPI:              newPurgeAPI(),
		ScanDataExportAPI:     newScanDataExportAPI(),
		JobserviceAPI:         newJobServiceAPI(),
		ScheduleAPI:           newScheduleAPI(),
		SecurityhubAPI:        newSecurityAPI(),
		PermissionsAPI:        newPermissionsAPIAPI(),
	})
	if err != nil {
		log.Fatal(err)
	}

	api.RegisterMiddleware("CopyArtifact", middleware.Chain(quota.CopyArtifactMiddleware(), blob.CopyArtifactMiddleware()))
	api.RegisterMiddleware("DeleteArtifact", quota.RefreshForProjectMiddleware())
	api.RegisterMiddleware("DeleteRepository", quota.RefreshForProjectMiddleware())

	api.BeforePrepare = beforePrepare
	api.ServeError = serveError

	return h
}

// function is called before the Prepare of the operation
func beforePrepare(ctx context.Context, operation string, _ interface{}) rmiddleware.Responder {
	metric.SetMetricOpID(ctx, operation)
	return nil
}

// Before executing operation handler, go-swagger will bind a parameters object to a request and validate the request,
// it will return directly when bind and validate failed.
// The response format of the default ServeError implementation does not match the internal error response format.
// So we needed to convert the format to the internal error response format.
func serveError(rw http.ResponseWriter, _ *http.Request, err error) {
	lib_http.SendError(rw, err)
}
