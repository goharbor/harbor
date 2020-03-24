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

package metadata

import (
	event2 "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/stretchr/testify/suite"
	"testing"
)

type scanEventTestSuite struct {
	suite.Suite
}

func (r *scanEventTestSuite) TestResolveOfScanImageEventMetadata() {
	e := &event.Event{}
	metadata := &ScanImageMetaData{
		Artifact: &v1.Artifact{
			NamespaceID: 0,
			Repository:  "library/hello-world",
			Tag:         "latest",
			Digest:      "sha256:absdfd87123",
			MimeType:    "docker.chart",
		},
		Status: "finished",
	}
	err := metadata.Resolve(e)
	r.Require().Nil(err)
	r.Equal(event2.TopicScanningCompleted, e.Topic)
	r.Require().NotNil(e.Data)
	data, ok := e.Data.(*event2.ScanImageEvent)
	r.Require().True(ok)
	r.Equal("library/hello-world", data.Artifact.Repository)
}

func TestScanEventTestSuite(t *testing.T) {
	suite.Run(t, &scanEventTestSuite{})
}
