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

package testing

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/types"
	"github.com/stretchr/testify/suite"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Suite ...
type Suite struct {
	suite.Suite
}

// SetupSuite ...
func (suite *Suite) SetupSuite() {
	config.Init()
	dao.PrepareTestForPostgresSQL()
}

// RandString ...
func (suite *Suite) RandString(n int, letters ...string) string {
	if len(letters) == 0 || len(letters[0]) == 0 {
		letters = []string{"abcdefghijklmnopqrstuvwxyz"}
	}

	letterBytes := []byte(letters[0])

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// WithProject ...
func (suite *Suite) WithProject(f func(int64, string), projectNames ...string) {
	var projectName string
	if len(projectNames) > 0 {
		projectName = projectNames[0]
	} else {
		projectName = suite.RandString(5)
	}

	projectID, err := dao.AddProject(models.Project{
		Name:    projectName,
		OwnerID: 1,
	})
	if err != nil {
		panic(err)
	}

	defer func() {
		dao.DeleteProject(projectID)
	}()

	f(projectID, projectName)
}

// AssertResourceUsage ...
func (suite *Suite) AssertResourceUsage(expected int64, resource types.ResourceName, projectID int64) {
	usage := models.QuotaUsage{Reference: "project", ReferenceID: strconv.FormatInt(projectID, 10)}
	err := dao.GetOrmer().Read(&usage, "reference", "reference_id")
	suite.Nil(err, fmt.Sprintf("Failed to get resource %s usage of project %d, error: %v", resource, projectID, err))

	used, err := types.NewResourceList(usage.Used)
	suite.Nil(err, "Bad resource usage of project %d", projectID)
	suite.Equal(expected, used[resource])
}
