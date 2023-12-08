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
	"encoding/json"
	"io"
	"strings"
	"testing"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/parser"
	"github.com/goharbor/harbor/src/testing/pkg/registry"
)

var (
	ormbConfig = `{
    "created": "2015-10-31T22:22:56.015925234Z",
    "author": "Ce Gao <gaoce@caicloud.io>",
    "description": "CNN Model",
    "tags": [
        "cv"
    ],
    "labels": {
        "tensorflow.version": "2.0.0"
    },
    "framework": "TensorFlow",
    "format": "SavedModel",
    "size": 9223372036854775807,
    "metrics": [
        {
            "name": "acc",
            "value": "0.9"
        }
    ],
    "hyperparameters": [
        {
            "name": "batch_size",
            "value": "32"
        }
    ],
    "signature": {
        "inputs": [
            {
                "name": "input_1",
                "size": [
                    224,
                    224,
                    3
                ],
                "dtype": "float64"
            }
        ],
        "outputs": [
            {
                "name": "output_1",
                "size": [
                    1,
                    1000
                ],
                "dtype": "float64"
            }
        ],
        "layers": [
            {
                "name": "conv"
            }
        ]
    },
    "training": {
        "git": {
            "repository": "git@github.com:caicloud/ormb.git",
            "revision": "22f1d8406d464b0c0874075539c1f2e96c253775"
        }
    },
    "dataset": {
        "git": {
            "repository": "git@github.com:caicloud/ormb.git",
            "revision": "22f1d8406d464b0c0874075539c1f2e96c253775"
        }
    }
}`
	ormbManifestWithoutIcon = `{
    "schemaVersion":2,
    "mediaType": "application/vnd.oci.image.manifest.v1+json",
    "config":{
        "mediaType":"application/vnd.caicloud.model.config.v1alpha1+json",
        "digest":"sha256:be948daf0e22f264ea70b713ea0db35050ae659c185706aa2fad74834455fe8c",
        "size":187,
        "annotations": {
            "io.goharbor.artifact.v1alpha1.skip-list": "metrics,git"
        }
    },
    "layers":[
        {
            "mediaType":"application/tar+gzip",
            "digest":"sha256:eb6063fecbb50a9d98268cb61746a0fd62a27a4af9e850ffa543a1a62d3948b2",
            "size":166022
        }
    ]
}`
	v2ManifestWithUnknownConfig = `{
    "schemaVersion": 2,
    "mediaType": "application/vnd.oci.image.manifest.v1+json",
    "config": {
        "mediaType": "application/vnd.nhl.peanut.butter.bagel",
        "digest": "sha256:ee29d2e91da0e5dbf6536f5b369148a83ef59b0ce96e49da65dd6c25eb1fa44f",
        "size": 33,
        "newUnspecifiedField": null
    },
    "layers": [
        {
            "mediaType": "application/vnd.oci.empty.v1+json",
            "digest": "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
            "size": 2,
            "newUnspecifiedField": "null"
        }
    ],
    "subject": {
        "mediaType": "application/vnd.oci.image.manifest.v1+json",
        "digest": "sha256:5a01bbc4ce6f52541cbc7e6af4b22bb107991a4bdd433103ff65aeb00756e906",
        "size": 714,
        "newUnspecifiedField": null
    }
 }`
	unknownConfig = `{NHL Peanut Butter on my NHL bagel}`
)

type defaultProcessorTestSuite struct {
	suite.Suite
	processor *defaultProcessor
	parser    *parser.Parser
	regCli    *registry.Client
}

func (d *defaultProcessorTestSuite) SetupTest() {
	d.regCli = &registry.Client{}
	d.processor = &defaultProcessor{
		regCli: d.regCli,
	}
	d.parser = &parser.Parser{}
}

func (d *defaultProcessorTestSuite) TestGetArtifactType() {
	mediaType := ""
	art := &artifact.Artifact{MediaType: mediaType}
	processor := &defaultProcessor{}
	typee := processor.GetArtifactType(nil, art)
	d.Equal(ArtifactTypeUnknown, typee)

	mediaType = "unknown"
	art = &artifact.Artifact{MediaType: mediaType}
	processor = &defaultProcessor{}
	typee = processor.GetArtifactType(nil, art)
	d.Equal(ArtifactTypeUnknown, typee)

	mediaType = "application/vnd.oci.empty.v1+json"
	art = &artifact.Artifact{MediaType: mediaType}
	processor = &defaultProcessor{}
	typee = processor.GetArtifactType(nil, art)
	d.Equal(ArtifactTypeUnknown, typee)

	mediaType = "application/vnd.nhl.peanut.butter.bagel"
	art = &artifact.Artifact{MediaType: mediaType}
	processor = &defaultProcessor{}
	typee = processor.GetArtifactType(nil, art)
	d.Equal(ArtifactTypeUnknown, typee)

	mediaType = "application/vnd.oci.image.config.v1+json"
	art = &artifact.Artifact{MediaType: mediaType}
	processor = &defaultProcessor{}
	typee = processor.GetArtifactType(nil, art)
	d.Equal("IMAGE", typee)

	mediaType = "application/vnd.cncf.helm.chart.config.v1+json"
	art = &artifact.Artifact{MediaType: mediaType}
	processor = &defaultProcessor{}
	typee = processor.GetArtifactType(nil, art)
	d.Equal("HELM.CHART", typee)

	mediaType = "application/vnd.sylabs.sif.config.v1+json"
	art = &artifact.Artifact{MediaType: mediaType}
	processor = &defaultProcessor{}
	typee = processor.GetArtifactType(nil, art)
	d.Equal("SIF", typee)

	mediaType = "application/vnd.caicloud.model.config.v1alpha1+json"
	art = &artifact.Artifact{MediaType: mediaType}
	processor = &defaultProcessor{}
	typee = processor.GetArtifactType(nil, art)
	d.Equal("MODEL", typee)
}

func (d *defaultProcessorTestSuite) TestAbstractMetadata() {
	manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(ormbManifestWithoutIcon))
	d.Require().Nil(err)
	manifestMediaType, content, err := manifest.Payload()
	d.Require().Nil(err)

	metadata := map[string]interface{}{}
	configBlob := io.NopCloser(strings.NewReader(ormbConfig))
	err = json.NewDecoder(configBlob).Decode(&metadata)
	d.Require().Nil(err)
	art := &artifact.Artifact{ManifestMediaType: manifestMediaType, ExtraAttrs: metadata}
	d.Len(art.ExtraAttrs, 13)

	d.regCli.On("PullBlob", mock.Anything, mock.Anything).Return(int64(0), configBlob, nil)
	d.parser.On("Parse", context.TODO(), mock.AnythingOfType("*artifact.Artifact"), mock.AnythingOfType("[]byte")).Return(nil)
	err = d.processor.AbstractMetadata(nil, art, content)
	d.Require().Nil(err)
	d.Len(art.ExtraAttrs, 12)
}

func (d *defaultProcessorTestSuite) TestAbstractMetadataWithUnknownConfig() {
	manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(v2ManifestWithUnknownConfig))
	d.Require().Nil(err)
	manifestMediaType, content, err := manifest.Payload()
	d.Require().Nil(err)

	configBlob := io.NopCloser(strings.NewReader(unknownConfig))
	d.regCli.On("PullBlob", mock.Anything, mock.Anything).Return(int64(0), configBlob, nil)
	art := &artifact.Artifact{ManifestMediaType: manifestMediaType}
	err = d.processor.AbstractMetadata(nil, art, content)
	d.Require().Nil(err)
	d.Len(art.ExtraAttrs, 0)
	d.Len(unknownConfig, 35)
}

func TestDefaultProcessorTestSuite(t *testing.T) {
	suite.Run(t, &defaultProcessorTestSuite{})
}
