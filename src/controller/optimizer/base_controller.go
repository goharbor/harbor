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

package optimizer

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/goharbor/harbor/src/common/rbac"
	ar "github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/robot"
	sc "github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/cache"
	_ "github.com/goharbor/harbor/src/lib/cache/memory" // memory cache
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	dockerfileoptdao "github.com/goharbor/harbor/src/pkg/dockerfileoptimization/dao"
	optimizerpkg "github.com/goharbor/harbor/src/pkg/optimizer"
	"github.com/goharbor/harbor/src/pkg/optimizer/dao"
	v1 "github.com/goharbor/harbor/src/pkg/optimizer/rest/v1"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	scanv1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/task"
)

const (
	// StatusUnhealthy indicates the adapter is not reachable
	StatusUnhealthy = "unhealthy"
	// StatusHealthy indicates the adapter is reachable
	StatusHealthy = "healthy"
	// RetrieveCapFailMsg the message indicate failed to retrieve the optimizer capabilities
	RetrieveCapFailMsg = "failed to retrieve optimizer capabilities, error %v"

	artifactIDKey = "artifact_id"
	robotIDKey    = "robot_id"
)

// DefaultController is a singleton api controller for optimizer adapters
var DefaultController = New()

var reservedNames = []string{"Sleeko"}

// New a basic controller
func New() Controller {
	return &basicController{
		manager: optimizerpkg.Mgr,
		optDAO:  dockerfileoptdao.New(),
		rc:      robot.Ctl,
		scanCtl: sc.DefaultController,
		execMgr: task.ExecMgr,
		taskMgr: task.Mgr,
		uuid: func() (string, error) {
			aUUID, err := uuid.NewUUID()
			if err != nil {
				return "", err
			}
			return aUUID.String(), nil
		},
		clientPool: v1.DefaultClientPool,
	}
}

// basicController is the default implementation of the Controller interface
type basicController struct {
	sync.Once

	// Manager for the optimizer registrations
	manager optimizerpkg.Manager
	// DAO for the dockerfile_optimization records
	optDAO dockerfileoptdao.DAO
	// Robot account controller
	rc robot.Controller
	// Scan controller, used to look up the artifact's vulnerability scan report
	scanCtl sc.Controller
	// Execution and task managers for the async job
	execMgr task.ExecutionManager
	taskMgr task.Manager
	// UUID generator
	uuid func() (string, error)
	// Client pool for talking to adapters
	clientPool v1.ClientPool
	// Cache of the adapter metadata
	cache cache.Cache
}

func (bc *basicController) Cache() cache.Cache {
	bc.Do(func() {
		bc.cache, _ = cache.New(cache.Memory, cache.Expiration(time.Second*30))
	})

	return bc.cache
}

// ListRegistrations ...
func (bc *basicController) ListRegistrations(ctx context.Context, query *q.Query) ([]*dao.Registration, error) {
	l, err := bc.manager.List(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "optimizer controller: list registrations")
	}
	return l, nil
}

// GetTotalOfRegistrations ...
func (bc *basicController) GetTotalOfRegistrations(ctx context.Context, query *q.Query) (int64, error) {
	return bc.manager.Count(ctx, query)
}

// CreateRegistration ...
func (bc *basicController) CreateRegistration(ctx context.Context, registration *dao.Registration) (string, error) {
	if isReservedName(registration.Name) {
		return "", errors.BadRequestError(nil).WithMessagef(`name "%s" is reserved, please try a different name`, registration.Name)
	}

	// Check if the registration is available
	if _, err := bc.Ping(ctx, registration); err != nil {
		return "", errors.Wrap(err, "optimizer controller: create registration")
	}

	// Check if there are any registrations already existing.
	l, err := bc.manager.List(ctx, &q.Query{
		PageSize:   1,
		PageNumber: 1,
	})
	if err != nil {
		return "", errors.Wrap(err, "optimizer controller: create registration")
	}

	if len(l) == 0 && !registration.IsDefault {
		// Mark the 1st as default automatically
		registration.IsDefault = true
	}

	return bc.manager.Create(ctx, registration)
}

// GetRegistration ...
func (bc *basicController) GetRegistration(ctx context.Context, registrationUUID string) (*dao.Registration, error) {
	r, err := bc.manager.Get(ctx, registrationUUID)
	if err != nil {
		return nil, errors.Wrap(err, "optimizer controller: get registration")
	}
	return r, nil
}

// UpdateRegistration ...
func (bc *basicController) UpdateRegistration(ctx context.Context, registration *dao.Registration) error {
	if registration.IsDefault && registration.Disabled {
		return errors.Errorf("default registration %s can not be marked to deactivated", registration.UUID)
	}

	if isReservedName(registration.Name) {
		return errors.BadRequestError(nil).WithMessagef(`name "%s" is reserved, please try a different name`, registration.Name)
	}

	return bc.manager.Update(ctx, registration)
}

// DeleteRegistration ...
func (bc *basicController) DeleteRegistration(ctx context.Context, registrationUUID string) (*dao.Registration, error) {
	registration, err := bc.manager.Get(ctx, registrationUUID)
	if err != nil {
		return nil, errors.Wrap(err, "optimizer controller: delete registration")
	}

	if registration == nil {
		// Not found
		return nil, nil
	}

	if err := bc.manager.Delete(ctx, registrationUUID); err != nil {
		return nil, errors.Wrap(err, "optimizer controller: delete registration")
	}

	return registration, nil
}

// SetDefaultRegistration ...
func (bc *basicController) SetDefaultRegistration(ctx context.Context, registrationUUID string) error {
	return bc.manager.SetAsDefault(ctx, registrationUUID)
}

// GetDefaultRegistration ...
func (bc *basicController) GetDefaultRegistration(ctx context.Context) (*dao.Registration, error) {
	return bc.manager.GetDefault(ctx)
}

// Ping ...
func (bc *basicController) Ping(ctx context.Context, registration *dao.Registration) (*v1.OptimizerAdapterMetadata, error) {
	if registration == nil {
		return nil, errors.New("nil registration to ping")
	}

	var (
		err  error
		meta *v1.OptimizerAdapterMetadata
	)

	if registration.ID > 0 {
		meta, err = bc.getAdapterMetadataWithCache(ctx, registration)
	} else {
		meta, err = bc.getAdapterMetadata(registration)
	}

	if err != nil {
		log.G(ctx).WithField("error", err).Error("failed to ping optimizer")

		return nil, errors.Wrap(err, "optimizer controller: ping")
	}

	if err := meta.Validate(); err != nil {
		return nil, err
	}

	return meta, nil
}

// GetMetadata ...
func (bc *basicController) GetMetadata(ctx context.Context, registrationUUID string) (*v1.OptimizerAdapterMetadata, error) {
	if len(registrationUUID) == 0 {
		return nil, errors.New("empty registration uuid")
	}

	r, err := bc.manager.Get(ctx, registrationUUID)
	if err != nil {
		return nil, errors.Wrap(err, "optimizer controller: get metadata")
	}

	if r == nil {
		return nil, errors.NotFoundError(nil).WithMessagef("registration %s not found", registrationUUID)
	}

	return bc.Ping(ctx, r)
}

// resolveDefault runs the checks shared by Optimize and CheckAvailability: the default
// registration exists, is enabled, is reachable, and supports the artifact's mime type.
func (bc *basicController) resolveDefault(ctx context.Context, artifact *ar.Artifact) (*dao.Registration, error) {
	r, err := bc.manager.GetDefault(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "optimizer controller: resolve default")
	}

	if r == nil {
		return nil, errors.PreconditionFailedError(nil).WithMessage("no optimizer adapter is configured")
	}

	if r.Disabled {
		return nil, errors.PreconditionFailedError(nil).WithMessagef("optimizer %s is deactivated", r.Name)
	}

	// Negotiate capability against the artifact manifest mime type
	meta, err := bc.Ping(ctx, r)
	if err != nil {
		return nil, errors.PreconditionFailedError(err).WithMessagef("optimizer %s is not reachable", r.Name)
	}
	r.Metadata = meta

	if !r.HasCapability(artifact.ManifestMediaType) {
		return nil, errors.BadRequestError(nil).
			WithMessagef("the configured optimizer %s does not support optimizing artifact with mime type %s", r.Name, artifact.ManifestMediaType)
	}

	return r, nil
}

// CheckAvailability ...
func (bc *basicController) CheckAvailability(ctx context.Context, artifact *ar.Artifact) error {
	if artifact == nil {
		return errors.New("nil artifact to check")
	}
	_, err := bc.resolveDefault(ctx, artifact)
	return err
}

// Optimize ...
func (bc *basicController) Optimize(ctx context.Context, artifact *ar.Artifact, tag string) (int64, error) {
	if artifact == nil {
		return 0, errors.New("nil artifact to optimize")
	}

	r, err := bc.resolveDefault(ctx, artifact)
	if err != nil {
		return 0, err
	}

	// Create the execution tracking this optimization
	extraAttrs := map[string]any{
		"artifact": map[string]any{
			"id":              artifact.ID,
			"project_id":      artifact.ProjectID,
			"repository_name": artifact.RepositoryName,
			"digest":          artifact.Digest,
		},
		"registration": map[string]any{
			"id":   r.ID,
			"name": r.Name,
		},
	}
	executionID, err := bc.execMgr.Create(ctx, job.OptimizeArtifactVendorType, artifact.ID, task.ExecutionTriggerManual, extraAttrs)
	if err != nil {
		return 0, errors.Wrap(err, "optimizer controller: create execution")
	}

	// Create/reset the Pending record the portal polls
	if err := bc.optDAO.Upsert(ctx, &dockerfileoptdao.DockerfileOptimization{
		RepositoryName:   artifact.RepositoryName,
		ArtifactDigest:   artifact.Digest,
		Status:           dockerfileoptdao.StatusPending,
		RegistrationUUID: r.UUID,
		ExecutionID:      executionID,
	}); err != nil {
		return 0, errors.Wrap(err, "optimizer controller: create pending record")
	}

	if err := bc.launchOptimizeJob(ctx, executionID, r, artifact, tag); err != nil {
		// mark the record errored so the portal does not poll forever
		if derr := bc.optDAO.UpdateStatus(ctx, artifact.RepositoryName, artifact.Digest, dockerfileoptdao.StatusError, err.Error()); derr != nil {
			log.G(ctx).Warningf("failed to persist error status: %v", derr)
		}
		return 0, errors.Wrap(err, "optimizer controller: launch optimize job")
	}

	return executionID, nil
}

// requiredPermissions returns the permission set of the single-use robot account the
// adapter uses to pull the artifact. scanner-pull is required in addition to pull:
// it bypasses the vulnerability-blocking middleware which would otherwise deny
// pulls of images that violate the project vulnerability policy.
func requiredPermissions() []*types.Policy {
	return []*types.Policy{
		{
			Resource: rbac.ResourceRepository,
			Action:   rbac.ActionPull,
		},
		{
			Resource: rbac.ResourceRepository,
			Action:   rbac.ActionScannerPull,
		},
	}
}

// makeRobotAccount creates a single-use project-level robot account the adapter uses
// to pull the artifact from the registry.
func (bc *basicController) makeRobotAccount(ctx context.Context, projectID int64, repository string, registration *dao.Registration) (*robot.Robot, error) {
	// Use uuid as name to avoid duplicated entries.
	UUID, err := bc.uuid()
	if err != nil {
		return nil, errors.Wrap(err, "optimizer controller: make robot account")
	}

	projectName := strings.Split(repository, "/")[0]
	robotPrefix := config.ScannerRobotPrefix(ctx)

	robotReq := &robot.Robot{
		Robot: model.Robot{
			Name:        fmt.Sprintf("%s-%s-%s", robotPrefix, registration.Name, UUID),
			Description: "for optimize",
			ProjectID:   projectID,
			Duration:    -1,
			CreatorType: "local",
			CreatorRef:  int64(0),
		},
		ProjectName: projectName,
		Level:       robot.LEVELPROJECT,
		Permissions: []*robot.Permission{
			{
				Kind:      "project",
				Namespace: projectName,
				Access:    requiredPermissions(),
			},
		},
	}

	rb, pwd, err := bc.rc.Create(ctx, robotReq)
	if err != nil {
		return nil, errors.Wrap(err, "optimizer controller: make robot account")
	}

	r, err := bc.rc.Get(ctx, rb, &robot.Option{WithPermission: false})
	if err != nil {
		return nil, errors.Wrap(err, "optimizer controller: make robot account")
	}
	r.Secret = pwd
	return r, nil
}

// launchOptimizeJob creates the jobservice task that submits the optimize request to
// the adapter and polls for the report.
func (bc *basicController) launchOptimizeJob(ctx context.Context, executionID int64, registration *dao.Registration, artifact *ar.Artifact, tag string) error {
	var registryAddr string
	if registration.UseInternalAddr {
		registryAddr = config.InternalCoreURL()
	} else {
		addr, err := config.ExtEndpoint()
		if err != nil {
			return errors.Wrap(err, "get registry endpoint")
		}
		registryAddr = addr
	}

	robotAcct, err := bc.makeRobotAccount(ctx, artifact.ProjectID, artifact.RepositoryName, registration)
	if err != nil {
		return err
	}

	scanReport, err := bc.fetchScanReport(ctx, artifact)
	if err != nil {
		return errors.Wrap(err, "get scan report")
	}

	optimizeReq := &v1.OptimizeRequest{
		Registry: &v1.Registry{
			URL: registryAddr,
		},
		Artifact: &v1.Artifact{
			Repository: artifact.RepositoryName,
			Digest:     artifact.Digest,
			Tag:        tag,
			MimeType:   artifact.ManifestMediaType,
		},
		ScanReport: scanReport,
	}

	rJSON, err := registration.ToJSON()
	if err != nil {
		return errors.Wrap(err, "marshal registration")
	}

	sJSON, err := optimizeReq.ToJSON()
	if err != nil {
		return errors.Wrap(err, "marshal optimize request")
	}

	robotJSON, err := robotAcct.ToJSON()
	if err != nil {
		return errors.Wrap(err, "marshal robot account")
	}

	params := make(map[string]any)
	params[optimizerpkg.JobParamRegistration] = rJSON
	params[optimizerpkg.JobParameterAuthType] = registration.GetRegistryAuthorizationType()
	params[optimizerpkg.JobParameterRequest] = sJSON
	params[optimizerpkg.JobParameterRobot] = robotJSON

	j := &task.Job{
		Name: job.OptimizeArtifactVendorType,
		Metadata: &job.Metadata{
			JobKind: job.KindGeneric,
		},
		Parameters: params,
	}

	extraAttrs := map[string]any{
		artifactIDKey: artifact.ID,
		robotIDKey:    robotAcct.ID,
	}

	_, err = bc.taskMgr.Create(ctx, executionID, j, extraAttrs)
	return err
}

// fetchScanReport returns the raw vulnerability scan report JSON for the artifact, or
// "" if the artifact has no completed scan yet (no scanner configured, not scanned,
// or the scan is still running/failed). A missing report is not treated as an error
// here: the adapter decides what to do about it after extracting the Dockerfile.
func (bc *basicController) fetchScanReport(ctx context.Context, artifact *ar.Artifact) (string, error) {
	reports, err := bc.scanCtl.GetReport(ctx, artifact, []string{scanv1.MimeTypeGenericVulnerabilityReport})
	if err != nil {
		if errors.IsNotFoundErr(err) {
			return "", nil
		}
		return "", err
	}

	for _, r := range reports {
		if r.Status == job.SuccessStatus.String() && len(r.Report) > 0 {
			return r.Report, nil
		}
	}

	return "", nil
}

func (bc *basicController) getAdapterMetadata(registration *dao.Registration) (*v1.OptimizerAdapterMetadata, error) {
	client, err := registration.Client(bc.clientPool)
	if err != nil {
		return nil, err
	}

	return client.GetMetadata()
}

func (bc *basicController) getAdapterMetadataWithCache(ctx context.Context, registration *dao.Registration) (*v1.OptimizerAdapterMetadata, error) {
	key := fmt.Sprintf("optimizer:reg:%d:metadata", registration.ID)

	var result MetadataResult
	err := cache.FetchOrSave(ctx, bc.Cache(), key, &result, func() (any, error) {
		meta, err := bc.getAdapterMetadata(registration)
		if err != nil {
			return &MetadataResult{Error: err.Error()}, nil
		}

		return &MetadataResult{Metadata: meta}, nil
	})

	if err != nil {
		return nil, err
	}

	return result.Unpack()
}

func isReservedName(name string) bool {
	return slices.Contains(reservedNames, name)
}

// MetadataResult metadata or error saved in cache
type MetadataResult struct {
	Metadata *v1.OptimizerAdapterMetadata
	Error    string
}

// Unpack get OptimizerAdapterMetadata and error from the result
func (m *MetadataResult) Unpack() (*v1.OptimizerAdapterMetadata, error) {
	var err error
	if m.Error != "" {
		err = errors.New(nil).WithMessage(m.Error)
	}

	return m.Metadata, err
}
