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

package wasm

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/docker/distribution/manifest/schema2"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/registry"
)

var (
	// For OCI fashion wasm artifact
	oci_manifest = `{
   "schemaVersion":2,
   "config":{
      "mediaType":"application/vnd.wasm.config.v1+json",
      "digest":"sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
      "size":2
   },
   "layers":[
      {
         "mediaType":"application/vnd.wasm.content.layer.v1+wasm",
         "digest":"sha256:d43012458290e4e2a350055bbe4a9f49fd4fb6b51d412089301e63ea4397ab4f",
         "size":3951005,
         "annotations":{
            "org.opencontainers.image.title":"test.wasm"
         }
      }
   ]
}`
	oci_config = `{}`

	// For annotation fashion wasm artifact
	annnotated_manifest = `{
   "schemaVersion":2,
   "mediaType":"application/vnd.oci.image.manifest.v1+json",
   "config":{
      "mediaType":"application/vnd.oci.image.config.v1+json",
      "digest":"sha256:6fd90b7cd05366c82ca32a3ff259e62dcb15b3b5e9672fe7d45609f29b6c1e95",
      "size":637
   },
   "layers":[
      {
         "mediaType":"application/vnd.oci.image.layer.v1.tar+gzip",
         "digest":"sha256:818ab7fdcd8d16270f01795d6aee7af0d1a06c71ce2cd3c1e8d8f946e9475450",
         "size":500361
      }
   ],
   "annotations":{
      "module.wasm.image/variant":"compat",
      "org.opencontainers.image.base.digest":"",
      "org.opencontainers.image.base.name":""
   }
}`
	annnotated_config = `{
   "created":"2022-03-02T09:02:41.01773982Z",
   "architecture":"amd64",
   "os":"linux",
   "config":{
      "Env":[
         "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
      ],
      "Cmd":[
         "/sleep.wasm"
      ],
      "Labels":{
         "io.buildah.version":"1.25.0-dev"
      }
   },
   "rootfs":{
      "type":"layers",
      "diff_ids":[
         "sha256:65885aa5fd4c157de98241de75669c4bf0d6f17d220c40069de31a572371a80d"
      ]
   },
   "history":[
      {
         "created":"2022-03-02T09:02:41.011039932Z",
         "created_by":"/bin/sh -c #(nop) COPY file:9fc00231cd29a2b8f76cfeaa9bc2355a7df54585b51d1f5dc98daf51557614b9 in / ",
         "empty_layer":true
      },
      {
         "created":"2022-03-02T09:02:41.043350231Z",
         "created_by":"/bin/sh -c #(nop) CMD [\"/sleep.wasm\"]"
      }
   ]
}`
)

type WASMProcessorTestSuite struct {
	suite.Suite
	processor *Processor
	regCli    *registry.Client
}

func (m *WASMProcessorTestSuite) SetupTest() {
	m.regCli = &registry.Client{}
	m.processor = &Processor{}
	m.processor.ManifestProcessor = &base.ManifestProcessor{RegCli: m.regCli}
}

func (m *WASMProcessorTestSuite) TestAbstractMetadataForAnnotationFashion() {
	artifact := &artifact.Artifact{}
	m.regCli.On("PullBlob", mock.Anything, mock.Anything).Return(int64(0), io.NopCloser(bytes.NewReader([]byte(annnotated_config))), nil)
	err := m.processor.AbstractMetadata(nil, artifact, []byte(annnotated_manifest))
	m.Require().Nil(err)
	m.NotNil(artifact.ExtraAttrs["created"])
	m.Equal("amd64", artifact.ExtraAttrs["architecture"])
	m.Equal("linux", artifact.ExtraAttrs["os"])
	m.NotNil(artifact.ExtraAttrs["config"])
	m.regCli.AssertExpectations(m.T())

}

func (m *WASMProcessorTestSuite) TestAbstractMetadataForOCIFashion() {
	artifact := &artifact.Artifact{}
	err := m.processor.AbstractMetadata(nil, artifact, []byte(oci_manifest))
	m.Require().Nil(err)
	m.NotNil(artifact.ExtraAttrs["org.opencontainers.image.title"])
	m.Equal(MediaType, artifact.ExtraAttrs["manifest.config.mediaType"])
	m.NotNil(artifact.ExtraAttrs["manifest.layers.mediaType"])
	m.regCli.AssertExpectations(m.T())
}

func (m *WASMProcessorTestSuite) TestAbstractAdditionForAnnotationFashion() {
	// unknown addition
	_, err := m.processor.AbstractAddition(nil, nil, "unknown_addition")
	m.True(errors.IsErr(err, errors.BadRequestCode))

	// build history
	artifact := &artifact.Artifact{}

	manifest := schema2.Manifest{}
	err = json.Unmarshal([]byte(annnotated_manifest), &manifest)
	deserializedManifest, err := schema2.FromStruct(manifest)
	m.Require().Nil(err)
	m.regCli.On("PullManifest", mock.Anything, mock.Anything).Return(deserializedManifest, "", nil)
	m.regCli.On("PullBlob", mock.Anything, mock.Anything).Return(int64(0), io.NopCloser(strings.NewReader(annnotated_config)), nil)
	addition, err := m.processor.AbstractAddition(nil, artifact, AdditionTypeBuildHistory)
	m.Require().Nil(err)
	m.Equal("application/json; charset=utf-8", addition.ContentType)
	m.Equal(`[{"created":"2022-03-02T09:02:41.011039932Z","created_by":"/bin/sh -c #(nop) COPY file:9fc00231cd29a2b8f76cfeaa9bc2355a7df54585b51d1f5dc98daf51557614b9 in / ","empty_layer":true},{"created":"2022-03-02T09:02:41.043350231Z","created_by":"/bin/sh -c #(nop) CMD [\"/sleep.wasm\"]"}]`, string(addition.Content))
}

func (m *WASMProcessorTestSuite) TestGetArtifactType() {
	m.Assert().Equal(ArtifactTypeWASM, m.processor.GetArtifactType(nil, nil))
}

func (m *WASMProcessorTestSuite) TestListAdditionTypes() {
	additions := m.processor.ListAdditionTypes(nil, nil)
	m.EqualValues([]string{AdditionTypeBuildHistory}, additions)
}

func TestManifestV2ProcessorTestSuite(t *testing.T) {
	suite.Run(t, &WASMProcessorTestSuite{})
}
