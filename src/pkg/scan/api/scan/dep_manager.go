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

package scan

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/job/models"
	cmo "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/token"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// DepManager provides dependant functions with an interface.
type DepManager interface {
	// Generate a UUID
	//
	//   Returns:
	//     string : the uuid string
	//     error  : non nil error if any errors occurred
	UUID() (string, error)

	// Submit a job
	//
	//   Arguments:
	//     jobData models.JobData : job data model
	//
	//   Returns:
	//     string : the uuid of the job
	//     error  : non nil error if any errors occurred
	SubmitJob(jobData *models.JobData) (string, error)

	// Get the endpoint of the registry
	//
	//   Returns:
	//     string : the uuid string
	//     error  : non nil error if any errors occurred
	GetRegistryEndpoint() (string, error)

	// Get the internal address of the core
	//
	//   Returns:
	//     string : the uuid string
	//     error  : non nil error if any errors occurred
	GetInternalCoreAddr() (string, error)

	// Make a robot account
	//
	//   Arguments:
	//     pid int64 : id of the project
	//     ttl int64 : expire time of the robot account
	//
	//   Returns:
	//     string : the token encoded string
	//     error  : non nil error if any errors occurred
	MakeRobotAccount(pid int64, ttl int64) (string, error)

	// Get the job log
	//
	//   Arguments:
	//     uuid string : the job uuid
	//
	//   Returns:
	//     []byte : the log text stream
	//     error  : non nil error if any errors occurred
	GetJobLog(uuid string) ([]byte, error)
}

// basicDepManager is the default implementation of dep manager
type basicDepManager struct{}

// UUID ...
func (bdm *basicDepManager) UUID() (string, error) {
	aUUID, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	return aUUID.String(), nil
}

// SubmitJob ...
func (bdm *basicDepManager) SubmitJob(jobData *models.JobData) (string, error) {
	return job.GlobalClient.SubmitJob(jobData)
}

// GetJobLog ...
func (bdm *basicDepManager) GetJobLog(uuid string) ([]byte, error) {
	return job.GlobalClient.GetJobLog(uuid)
}

// GetRegistryEndpoint ...
func (bdm *basicDepManager) GetRegistryEndpoint() (string, error) {
	return config.ExtEndpoint()
}

// GetInternalCoreAddr ...
func (bdm *basicDepManager) GetInternalCoreAddr() (string, error) {
	return config.InternalCoreURL(), nil
}

// MakeRobotAccount ...
func (bdm *basicDepManager) MakeRobotAccount(pid int64, ttl int64) (string, error) {
	// Use uuid as name to avoid duplicated entries.
	UUID, err := uuid.NewUUID()
	if err != nil {
		return "", errors.Wrap(err, "basic dep manager: make robot account")
	}
	createdName := fmt.Sprintf("%s%s", common.RobotPrefix, UUID.String())

	expireAt := time.Now().UTC().Add(time.Duration(ttl) * time.Second).Unix()
	// First to add a robot account, and get its id.
	robot := cmo.Robot{
		Name:      createdName,
		ProjectID: pid,
		ExpiresAt: expireAt,
	}
	id, err := dao.AddRobot(&robot)
	if err != nil {
		return "", errors.Wrap(err, "basic dep manager: make robot account")
	}

	resource := fmt.Sprintf("/project/%d/repository", pid)
	access := []*rbac.Policy{{
		Resource: rbac.Resource(resource),
		Action:   "pull",
	}}

	// Generate the token, token is not stored in the database.
	jwtToken, err := token.New(id, pid, expireAt, access)
	if err != nil {
		return "", errors.Wrap(err, "basic dep manager: make robot account")
	}

	rawToken, err := jwtToken.Raw()
	if err != nil {
		return "", errors.Wrap(err, "basic dep manager: make robot account")
	}

	basic := fmt.Sprintf("%s:%s", createdName, rawToken)
	encoded := base64.StdEncoding.EncodeToString([]byte(basic))

	return fmt.Sprintf("Basic %s", encoded), nil
}
