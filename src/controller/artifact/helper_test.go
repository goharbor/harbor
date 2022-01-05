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

package artifact

import (
	"context"
	accessorymodel "github.com/goharbor/harbor/src/pkg/accessory/model"
	"testing"

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/testing/pkg/accessory"
	accessorytesting "github.com/goharbor/harbor/src/testing/pkg/accessory"
	artifacttesting "github.com/goharbor/harbor/src/testing/pkg/artifact"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type IteratorTestSuite struct {
	suite.Suite

	artMgr *artifacttesting.Manager
	accMgr *accessory.Manager

	ctl         *controller
	originalCtl Controller
}

func (suite *IteratorTestSuite) SetupSuite() {
	suite.artMgr = &artifacttesting.Manager{}
	suite.accMgr = &accessorytesting.Manager{}

	suite.originalCtl = Ctl
	suite.ctl = &controller{
		artMgr:       suite.artMgr,
		accessoryMgr: suite.accMgr,
	}
	Ctl = suite.ctl
}

func (suite *IteratorTestSuite) TeardownSuite() {
	Ctl = suite.originalCtl
}

func (suite *IteratorTestSuite) TestIterator() {
	suite.accMgr.On("List", mock.Anything, mock.Anything).Return([]accessorymodel.Accessory{}, nil)
	q1 := &q.Query{PageNumber: 1, PageSize: 5, Keywords: map[string]interface{}{}}
	suite.artMgr.On("List", mock.Anything, q1).Return([]*artifact.Artifact{
		{ID: 1},
		{ID: 2},
		{ID: 3},
		{ID: 4},
		{ID: 5},
	}, nil)

	q2 := &q.Query{PageNumber: 2, PageSize: 5, Keywords: map[string]interface{}{}}
	suite.artMgr.On("List", mock.Anything, q2).Return([]*artifact.Artifact{
		{ID: 6},
		{ID: 7},
		{ID: 8},
	}, nil)

	var artifacts []*Artifact
	for art := range Iterator(context.TODO(), 5, nil, nil) {
		artifacts = append(artifacts, art)
	}

	suite.Len(artifacts, 8)
}

func TestIteratorTestSuite(t *testing.T) {
	suite.Run(t, &IteratorTestSuite{})
}
