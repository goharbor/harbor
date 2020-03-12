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

package processor

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type defaultProcessorTestSuite struct {
	suite.Suite
}

func (d *defaultProcessorTestSuite) TestGetArtifactType() {
	mediaType := ""
	processor := &defaultProcessor{mediaType: mediaType}
	typee := processor.GetArtifactType()
	d.Equal(ArtifactTypeUnknown, typee)

	mediaType = "unknown"
	processor = &defaultProcessor{mediaType: mediaType}
	typee = processor.GetArtifactType()
	d.Equal(ArtifactTypeUnknown, typee)

	mediaType = "application/vnd.oci.image.config.v1+json"
	processor = &defaultProcessor{mediaType: mediaType}
	typee = processor.GetArtifactType()
	d.Equal("IMAGE", typee)

	mediaType = "application/vnd.cncf.helm.chart.config.v1+json"
	processor = &defaultProcessor{mediaType: mediaType}
	typee = processor.GetArtifactType()
	d.Equal("HELM.CHART", typee)

	mediaType = "application/vnd.sylabs.sif.config.v1+json"
	processor = &defaultProcessor{mediaType: mediaType}
	typee = processor.GetArtifactType()
	d.Equal("SIF", typee)
}

func TestDefaultProcessorTestSuite(t *testing.T) {
	suite.Run(t, &defaultProcessorTestSuite{})
}
