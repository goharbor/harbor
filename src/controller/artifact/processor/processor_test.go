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
	"context"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/stretchr/testify/suite"
	"testing"
)

type fakeProcessor struct{}

func (f *fakeProcessor) GetArtifactType(ctx context.Context, artifact *artifact.Artifact) string {
	return ""
}
func (f *fakeProcessor) ListAdditionTypes(ctx context.Context, artifact *artifact.Artifact) []string {
	return nil
}
func (f *fakeProcessor) AbstractMetadata(ctx context.Context, artifact *artifact.Artifact, manifest []byte) error {
	return nil
}
func (f *fakeProcessor) AbstractAddition(ctx context.Context, artifact *artifact.Artifact, additionType string) (*Addition, error) {
	return nil, nil
}

type processorTestSuite struct {
	suite.Suite
}

func (p *processorTestSuite) SetupTest() {
	Registry = map[string]Processor{}
}

func (p *processorTestSuite) TestRegister() {
	// success
	mediaType := "fake_media_type"
	err := Register(nil, mediaType)
	p.Require().Nil(err)

	// conflict
	err = Register(nil, mediaType)
	p.Require().NotNil(err)
}

func (p *processorTestSuite) TestGet() {
	// register a processor
	mediaType := "fake_media_type"
	err := Register(&fakeProcessor{}, mediaType)
	p.Require().Nil(err)

	// get the processor
	processor := Get(mediaType)
	p.Require().NotNil(processor)
	_, ok := processor.(*fakeProcessor)
	p.True(ok)

	// get the not existing processor
	processor = Get("not_existing_media_type")
	p.Require().NotNil(processor)
	_, ok = processor.(*defaultProcessor)
	p.True(ok)
}

func TestProcessorTestSuite(t *testing.T) {
	suite.Run(t, &processorTestSuite{})
}
