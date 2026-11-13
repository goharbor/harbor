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
	"context"
	"testing"

	"github.com/docker/distribution/manifest/schema2"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact"
	accessoryModel "github.com/goharbor/harbor/src/pkg/accessory/model"
	accessorybase "github.com/goharbor/harbor/src/pkg/accessory/model/base"
	artmodel "github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	scannertesting "github.com/goharbor/harbor/src/testing/controller/scanner"
	"github.com/goharbor/harbor/src/testing/mock"
	accessorytesting "github.com/goharbor/harbor/src/testing/pkg/accessory"
)

type CheckerTestSuite struct {
	suite.Suite
}

func (suite *CheckerTestSuite) new() *checker {
	artifactCtl := &artifacttesting.Controller{}
	scannerCtl := &scannertesting.Controller{}
	accessoryMgr := &accessorytesting.Manager{}

	return &checker{
		artifactCtl:   artifactCtl,
		scannerCtl:    scannerCtl,
		accMgr:        accessoryMgr,
		registrations: map[int64]*scanner.Registration{},
	}
}

func (suite *CheckerTestSuite) TestScannerNotFound() {
	c := suite.new()

	{
		mock.OnAnything(c.scannerCtl, "GetRegistrationByProject").Return(nil, nil)

		isScannable, err := c.IsScannable(context.TODO(), &artifact.Artifact{})
		suite.Nil(err)
		suite.False(isScannable)
	}
}

func (suite *CheckerTestSuite) TestIsScannable() {
	c := suite.new()

	supportMimeType := "support mime type"

	mock.OnAnything(c.scannerCtl, "GetRegistrationByProject").Return(&scanner.Registration{
		Metadata: &v1.ScannerAdapterMetadata{
			Capabilities: []*v1.ScannerCapability{
				{ConsumesMimeTypes: []string{supportMimeType}},
			},
		},
	}, nil)

	{

		art := &artifact.Artifact{}

		mock.OnAnything(c.accMgr, "List").Return([]accessoryModel.Accessory{}, nil).Once()
		mock.OnAnything(c.artifactCtl, "Walk").Return(nil).Once().Run(func(args mock.Arguments) {
			walkFn := args.Get(2).(func(*artifact.Artifact) error)
			walkFn(art)
		})
		mock.OnAnything(c.artifactCtl, "HasUnscannableLayer").Return(false, nil).Once()
		isScannable, err := c.IsScannable(context.TODO(), art)
		suite.Nil(err)
		suite.False(isScannable)
	}

	{
		art := &artifact.Artifact{}
		art.Type = "IMAGE"
		art.ManifestMediaType = supportMimeType

		mock.OnAnything(c.accMgr, "List").Return([]accessoryModel.Accessory{}, nil).Once()
		mock.OnAnything(c.artifactCtl, "Walk").Return(nil).Once().Run(func(args mock.Arguments) {
			walkFn := args.Get(2).(func(*artifact.Artifact) error)
			walkFn(art)
		})
		mock.OnAnything(c.artifactCtl, "HasUnscannableLayer").Return(false, nil).Once()

		isScannable, err := c.IsScannable(context.TODO(), art)
		suite.Nil(err)
		suite.True(isScannable)
	}
}

func (suite *CheckerTestSuite) TestRunnableSignatureAccessoryIsScannable() {
	for _, mediaType := range []string{ocispec.MediaTypeImageConfig, schema2.MediaTypeImageConfig} {
		suite.Run(mediaType, func() {
			c := suite.new()
			art := &artifact.Artifact{Artifact: artmodel.Artifact{MediaType: mediaType, ManifestMediaType: ocispec.MediaTypeImageManifest}}
			registration := &scanner.Registration{Metadata: &v1.ScannerAdapterMetadata{
				Capabilities: []*v1.ScannerCapability{{ConsumesMimeTypes: []string{ocispec.MediaTypeImageManifest}}},
			}}
			mock.OnAnything(c.scannerCtl, "GetRegistrationByProject").Return(registration, nil).Once()
			mock.OnAnything(c.accMgr, "List").Return([]accessoryModel.Accessory{
				&accessorybase.Default{Data: accessoryModel.AccessoryData{Type: accessoryModel.TypeNotationSignature}},
			}, nil).Once()
			mock.OnAnything(c.artifactCtl, "HasUnscannableLayer").Return(false, nil).Once()
			mock.OnAnything(c.artifactCtl, "Walk").Return(nil).Once().Run(func(args mock.Arguments) {
				walkFn := args.Get(2).(func(*artifact.Artifact) error)
				walkFn(art)
			})
			scannable, err := c.IsScannable(context.Background(), art)
			suite.Require().NoError(err)
			suite.True(scannable)
		})
	}
}

func (suite *CheckerTestSuite) TestRunnableAccessoryWithExcludedLayerIsNotScannable() {
	c := suite.new()
	art := &artifact.Artifact{Artifact: artmodel.Artifact{MediaType: ocispec.MediaTypeImageConfig, ManifestMediaType: ocispec.MediaTypeImageManifest}}
	registration := &scanner.Registration{Metadata: &v1.ScannerAdapterMetadata{
		Capabilities: []*v1.ScannerCapability{{ConsumesMimeTypes: []string{ocispec.MediaTypeImageManifest}}},
	}}
	mock.OnAnything(c.scannerCtl, "GetRegistrationByProject").Return(registration, nil).Once()
	mock.OnAnything(c.accMgr, "List").Return([]accessoryModel.Accessory{&accessorybase.Default{}}, nil).Once()
	mock.OnAnything(c.artifactCtl, "HasUnscannableLayer").Return(true, nil).Once()
	mock.OnAnything(c.artifactCtl, "Walk").Return(nil).Once().Run(func(args mock.Arguments) {
		walkFn := args.Get(2).(func(*artifact.Artifact) error)
		suite.Require().NoError(walkFn(art))
	})
	scannable, err := c.IsScannable(context.Background(), art)
	suite.Require().NoError(err)
	suite.False(scannable)
}

func (suite *CheckerTestSuite) TestAccessoryClassification() {
	tests := []struct {
		name              string
		mediaType         string
		accessoryType     string
		skip              bool
		runnableSignature bool
	}{
		{"notation signature", "application/vnd.oci.empty.v1+json", accessoryModel.TypeNotationSignature, true, false},
		{"nydus image", ocispec.MediaTypeImageConfig, accessoryModel.TypeNydusAccelerator, false, true},
		{"non-runnable nydus", "application/vnd.oci.empty.v1+json", accessoryModel.TypeNydusAccelerator, true, false},
		{"image claiming signature", ocispec.MediaTypeImageConfig, accessoryModel.TypeCosignSignature, false, true},
		{"image with subject", ocispec.MediaTypeImageConfig, accessoryModel.TypeSubject, false, true},
		{"image claiming sbom", ocispec.MediaTypeImageConfig, accessoryModel.TypeHarborSBOM, false, true},
		{"wasm claiming signature", "application/vnd.wasm.config.v1+json", accessoryModel.TypeCosignSignature, false, true},
	}
	for _, tc := range tests {
		suite.Run(tc.name, func() {
			c := suite.new()
			art := &artifact.Artifact{Artifact: artmodel.Artifact{MediaType: tc.mediaType}}
			mock.OnAnything(c.accMgr, "List").Return([]accessoryModel.Accessory{
				&accessorybase.Default{Data: accessoryModel.AccessoryData{Type: tc.accessoryType}},
			}, nil).Once()
			skip, runnable, err := scanAccessoryStatus(context.Background(), c.accMgr, art)
			suite.Require().NoError(err)
			suite.Equal(tc.skip, skip)
			suite.Equal(tc.runnableSignature, runnable)
		})
	}
}

func (suite *CheckerTestSuite) TestRunnableSignatureCollectedForScanning() {
	c := suite.new()
	art := &artifact.Artifact{Artifact: artmodel.Artifact{MediaType: schema2.MediaTypeImageConfig, ManifestMediaType: ocispec.MediaTypeImageManifest}}
	registration := &scanner.Registration{Metadata: &v1.ScannerAdapterMetadata{
		Capabilities: []*v1.ScannerCapability{{ConsumesMimeTypes: []string{ocispec.MediaTypeImageManifest}}},
	}}
	mock.OnAnything(c.accMgr, "List").Return([]accessoryModel.Accessory{
		&accessorybase.Default{Data: accessoryModel.AccessoryData{Type: accessoryModel.TypeCosignSignature}},
	}, nil).Once()
	mock.OnAnything(c.artifactCtl, "HasUnscannableLayer").Return(false, nil).Once()
	mock.OnAnything(c.artifactCtl, "Walk").Return(nil).Once().Run(func(args mock.Arguments) {
		walkFn := args.Get(2).(func(*artifact.Artifact) error)
		walkFn(art)
	})
	bc := &basicController{ar: c.artifactCtl, acc: c.accMgr}
	artifacts, scannable, err := bc.collectScanningArtifacts(context.Background(), registration, art)
	suite.Require().NoError(err)
	suite.True(scannable)
	suite.Equal([]*artifact.Artifact{art}, artifacts)
}

func TestCheckerTestSuite(t *testing.T) {
	suite.Run(t, &CheckerTestSuite{})
}
