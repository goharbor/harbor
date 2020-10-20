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

package annotation

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/distribution"
	reg "github.com/goharbor/harbor/src/testing/pkg/registry"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"
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
	ormbManifest = `{
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
            "mediaType": "image/png",
            "digest": "sha256:d923b93eadde0af5c639a972710a4d919066aba5d0dfbf4b9385099f70272da0",
            "size": 166015,
            "annotations": { 
                "io.goharbor.artifact.v1alpha1.icon": ""
            }
        },
        {
            "mediaType":"application/tar+gzip",
            "digest":"sha256:eb6063fecbb50a9d98268cb61746a0fd62a27a4af9e850ffa543a1a62d3948b2",
            "size":166022
        }
    ]
}`
	ormbManifestWithoutSkipList = `{
    "schemaVersion":2,
    "mediaType": "application/vnd.oci.image.manifest.v1+json",
    "config":{
        "mediaType":"application/vnd.caicloud.model.config.v1alpha1+json",
        "digest":"sha256:be948daf0e22f264ea70b713ea0db35050ae659c185706aa2fad74834455fe8c",
        "size":187
    },
    "layers":[
        {
            "mediaType": "image/png",
            "digest": "sha256:d923b93eadde0af5c639a972710a4d919066aba5d0dfbf4b9385099f70272da0",
            "size": 166015,
            "annotations": { 
                "io.goharbor.artifact.v1alpha1.icon": ""
            }
        },
        {
            "mediaType":"application/tar+gzip",
            "digest":"sha256:eb6063fecbb50a9d98268cb61746a0fd62a27a4af9e850ffa543a1a62d3948b2",
            "size":166022
        }
    ]
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
	ormbIcon = "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LQAAgGBCCwAAIJjQAgAACCa0AAAAggktAACAYEILAAAgmNACAAAIJrQAAACCCS0AAIBgQgsAACCY0AIAAAgmtAAAAIIJLQAAgGBCCwAAIJjQAgAACCa0AAAAggktAACAYEILAAAgmNACAAAIJrQAAACCCS0AAIBgQgsAACCY0AIAAAgmtAAAAIIJLQAAgGBCCwAAIJjQAgAACCa0AAAAggktAACAYEILAAAgmNACAAAIJrQAAACCCS0AAIBgQgsAACCY0AIAAAgmtAAAAIIJLQAAgGBCCwAAIJjQAgAACCa0AAAAggktAACAYEILAAAgmNACAAAIJrQAAACCCS0AAIBgQgsAACCY0AIAAAgmtAAAAIIJLQAAgGBCCwAAIJjQAgAACCa0AAAAggktAACAYEILAAAgmNACAAAIJrQAAACCCS0AAIBgQgsAACCY0AIAAAgmtAAAAIIJLQAAgGBCCwAAIJjQAgAACCa0AAAAggktAACAYEILAAAgmNACAAAIJrQAAACCCS0AAIBgQgsAACCY0AIAAAgmtAAAAIIJLQAAgGBCCwAAIJjQAgAACCa0AAAAggktAACAYEILAAAgmNACAAAIJrQAAACCCS0AAIBgQgsAACCY0AIAAAgmtAAAAIIJLQAAgGBCCwAAIJjQAgAACCa0AAAAggktAACAYEILAAAgmNACAAAIJrQAAACCCS0AAIBgQgsAACCY0AIAAAgmtAAAAIIJLQAAgGBCCwAAIJjQAgAACCa0AAAAggktAACAYEILAAAgmNACAAAIJrQAAACCCS0AAIBgQgsAACBS0zT/H29PQcL+62hzAAAAAElFTkSuQmCC"
)

// v1alpha1TestSuite is a test suite of testing v1alpha1 parser
type v1alpha1TestSuite struct {
	suite.Suite
	regCli         *reg.FakeClient
	v1alpha1Parser *v1alpha1Parser
}

func (p *v1alpha1TestSuite) SetupTest() {
	p.regCli = &reg.FakeClient{}
	p.v1alpha1Parser = &v1alpha1Parser{
		regCli: p.regCli,
	}
}

func (p *v1alpha1TestSuite) TestParse() {
	manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(ormbManifest))
	p.Require().Nil(err)
	manifestMediaType, content, err := manifest.Payload()
	p.Require().Nil(err)

	metadata := map[string]interface{}{}
	configBlob := ioutil.NopCloser(strings.NewReader(ormbConfig))
	err = json.NewDecoder(configBlob).Decode(&metadata)
	p.Require().Nil(err)
	art := &artifact.Artifact{ManifestMediaType: manifestMediaType, ExtraAttrs: metadata}

	blob := ioutil.NopCloser(base64.NewDecoder(base64.StdEncoding, strings.NewReader(ormbIcon)))
	p.regCli.On("PullBlob").Return(0, blob, nil)
	err = p.v1alpha1Parser.Parse(nil, art, content)
	p.Require().Nil(err)
	p.Len(art.ExtraAttrs, 12)
	p.Equal("CNN Model", art.ExtraAttrs["description"])
	p.Equal("TensorFlow", art.ExtraAttrs["framework"])
	p.Equal([]interface{}{map[string]interface{}{"name": "batch_size", "value": "32"}}, art.ExtraAttrs["hyperparameters"])
	p.Equal("sha256:d923b93eadde0af5c639a972710a4d919066aba5d0dfbf4b9385099f70272da0", art.Icon)

	// reset the mock
	p.SetupTest()
	manifest, _, err = distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(ormbManifestWithoutSkipList))
	p.Require().Nil(err)
	manifestMediaType, content, err = manifest.Payload()
	p.Require().Nil(err)

	metadata = map[string]interface{}{}
	configBlob = ioutil.NopCloser(strings.NewReader(ormbConfig))
	err = json.NewDecoder(configBlob).Decode(&metadata)
	p.Require().Nil(err)
	art = &artifact.Artifact{ManifestMediaType: manifestMediaType, ExtraAttrs: metadata}

	blob = ioutil.NopCloser(base64.NewDecoder(base64.StdEncoding, strings.NewReader(ormbIcon)))
	p.regCli.On("PullBlob").Return(0, blob, nil)
	err = p.v1alpha1Parser.Parse(nil, art, content)
	p.Require().Nil(err)
	p.Len(art.ExtraAttrs, 13)
	p.Equal("CNN Model", art.ExtraAttrs["description"])
	p.Equal("TensorFlow", art.ExtraAttrs["framework"])
	p.Equal([]interface{}{map[string]interface{}{"name": "batch_size", "value": "32"}}, art.ExtraAttrs["hyperparameters"])
	p.Equal("sha256:d923b93eadde0af5c639a972710a4d919066aba5d0dfbf4b9385099f70272da0", art.Icon)

	// reset the mock
	p.SetupTest()
	manifest, _, err = distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(ormbManifestWithoutIcon))
	p.Require().Nil(err)
	manifestMediaType, content, err = manifest.Payload()
	p.Require().Nil(err)

	metadata = map[string]interface{}{}
	configBlob = ioutil.NopCloser(strings.NewReader(ormbConfig))
	err = json.NewDecoder(configBlob).Decode(&metadata)
	p.Require().Nil(err)
	art = &artifact.Artifact{ManifestMediaType: manifestMediaType, ExtraAttrs: metadata}

	err = p.v1alpha1Parser.Parse(nil, art, content)
	p.Require().Nil(err)
	p.Len(art.ExtraAttrs, 12)
	p.Equal("CNN Model", art.ExtraAttrs["description"])
	p.Equal("TensorFlow", art.ExtraAttrs["framework"])
	p.Equal([]interface{}{map[string]interface{}{"name": "batch_size", "value": "32"}}, art.ExtraAttrs["hyperparameters"])
	p.Equal("", art.Icon)
}

func TestDefaultProcessorTestSuite(t *testing.T) {
	suite.Run(t, &v1alpha1TestSuite{})
}
