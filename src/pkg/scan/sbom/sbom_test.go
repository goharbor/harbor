package sbom

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sc "github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/controller/scanner"
	"github.com/goharbor/harbor/src/lib/orm"
	accessoryModel "github.com/goharbor/harbor/src/pkg/accessory/model"
	basemodel "github.com/goharbor/harbor/src/pkg/accessory/model/base"
	art "github.com/goharbor/harbor/src/pkg/artifact"
	sbomModel "github.com/goharbor/harbor/src/pkg/scan/sbom/model"
	htesting "github.com/goharbor/harbor/src/testing"
	artifactTest "github.com/goharbor/harbor/src/testing/controller/artifact"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/mock"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/task"
	scanTest "github.com/goharbor/harbor/src/testing/controller/scan"
	scannerTest "github.com/goharbor/harbor/src/testing/controller/scanner"
	"github.com/goharbor/harbor/src/testing/jobservice"
	sbomTest "github.com/goharbor/harbor/src/testing/pkg/scan/sbom"
	taskTest "github.com/goharbor/harbor/src/testing/pkg/task"
)

var registeredScanner = &scanner.Registration{
	UUID: "uuid",
	Metadata: &v1.ScannerAdapterMetadata{
		Capabilities: []*v1.ScannerCapability{
			{Type: v1.ScanTypeVulnerability, ConsumesMimeTypes: []string{v1.MimeTypeDockerArtifact}, ProducesMimeTypes: []string{v1.MimeTypeGenericVulnerabilityReport}},
			{Type: v1.ScanTypeSbom, ConsumesMimeTypes: []string{v1.MimeTypeDockerArtifact}, ProducesMimeTypes: []string{v1.MimeTypeSBOMReport}},
		},
	},
}

func Test_scanHandler_ReportURLParameter(t *testing.T) {
	type args struct {
		in0 *v1.ScanRequest
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"normal test", args{&v1.ScanRequest{}}, "sbom_media_type=application%2Fspdx%2Bjson", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &scanHandler{}
			got, err := v.URLParameter(tt.args.in0)
			if (err != nil) != tt.wantErr {
				t.Errorf("URLParameter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("URLParameter() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scanHandler_RequiredPermissions(t *testing.T) {
	tests := []struct {
		name string
		want []*types.Policy
	}{
		{"normal test", []*types.Policy{
			{
				Resource: rbac.ResourceRepository,
				Action:   rbac.ActionPull,
			},
			{
				Resource: rbac.ResourceRepository,
				Action:   rbac.ActionScannerPull,
			},
			{
				Resource: rbac.ResourceRepository,
				Action:   rbac.ActionPush,
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &scanHandler{}
			if got := v.RequiredPermissions(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RequiredPermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scanHandler_RequestProducesMineTypes(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{"normal test", []string{v1.MimeTypeSBOMReport}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &scanHandler{}
			if got := v.RequestProducesMineTypes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RequestProducesMineTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mockGetRegistry(ctx context.Context) (string, bool) {
	return "myharbor.example.com", false
}

func mockGenAccessory(scanRep v1.ScanRequest, sbomContent []byte, labels map[string]string, mediaType string, robot *model.Robot) (string, error) {
	return "sha256:1234567890", nil
}

type SBOMTestSuite struct {
	htesting.Suite
	handler           *scanHandler
	sbomManager       *sbomTest.Manager
	taskMgr           *taskTest.Manager
	artifactCtl       *artifactTest.Controller
	artifact          *artifact.Artifact
	wrongArtifact     *artifact.Artifact
	scanController    *scanTest.Controller
	scannerController *scannerTest.Controller
}

func (suite *SBOMTestSuite) SetupSuite() {
	suite.sbomManager = &sbomTest.Manager{}
	suite.taskMgr = &taskTest.Manager{}
	suite.artifactCtl = &artifactTest.Controller{}
	suite.scannerController = &scannerTest.Controller{}
	suite.scanController = &scanTest.Controller{}

	suite.handler = &scanHandler{
		GenAccessoryFunc:       mockGenAccessory,
		SBOMMgrFunc:            func() Manager { return suite.sbomManager },
		TaskMgrFunc:            func() task.Manager { return suite.taskMgr },
		ArtifactControllerFunc: func() artifact.Controller { return suite.artifactCtl },
		ScanControllerFunc:     func() sc.Controller { return suite.scanController },
		ScannerControllerFunc:  func() scanner.Controller { return suite.scannerController },
		cloneCtx: func(ctx context.Context) context.Context {
			return ctx
		},
	}

	suite.artifact = &artifact.Artifact{Artifact: art.Artifact{ID: 1}}
	suite.artifact.Type = "IMAGE"
	suite.artifact.ProjectID = 1
	suite.artifact.RepositoryName = "library/photon"
	suite.artifact.Digest = "digest-code"
	suite.artifact.ManifestMediaType = v1.MimeTypeDockerArtifact

	suite.wrongArtifact = &artifact.Artifact{Artifact: art.Artifact{ID: 2, ProjectID: 1}}
	suite.wrongArtifact.Digest = "digest-wrong"
}

func (suite *SBOMTestSuite) TearDownSuite() {
}

func (suite *SBOMTestSuite) TestPostScan() {
	req := &v1.ScanRequest{
		Registry: &v1.Registry{
			URL: "myregistry.example.com",
		},
		Artifact: &v1.Artifact{
			Repository: "library/nosql",
		},
	}
	robot := &model.Robot{
		Name:   "robot",
		Secret: "mysecret",
	}
	startTime := time.Now()
	rawReport := `{"sbom": { "key": "value" }}`
	ctx := &jobservice.MockJobContext{}
	ctx.On("GetLogger").Return(&jobservice.MockJobLogger{})
	accessory, err := suite.handler.PostScan(ctx, req, nil, rawReport, startTime, robot)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(accessory)
}

func (suite *SBOMTestSuite) TestMakeReportPlaceHolder() {
	ctx := orm.NewContext(nil, &ormtesting.FakeOrmer{})
	acc := &basemodel.Default{
		Data: accessoryModel.AccessoryData{
			ID:                1,
			ArtifactID:        2,
			SubArtifactDigest: "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180",
			Type:              accessoryModel.TypeHarborSBOM,
		},
	}
	art := &artifact.Artifact{Artifact: art.Artifact{ID: 1, Digest: "digest", ManifestMediaType: v1.MimeTypeDockerArtifact},
		Accessories: []accessoryModel.Accessory{acc}}
	r := &scanner.Registration{
		UUID: "uuid",
		Metadata: &v1.ScannerAdapterMetadata{
			Capabilities: []*v1.ScannerCapability{
				{Type: v1.ScanTypeVulnerability, ConsumesMimeTypes: []string{v1.MimeTypeDockerArtifact}, ProducesMimeTypes: []string{v1.MimeTypeGenericVulnerabilityReport}},
				{Type: v1.ScanTypeSbom, ConsumesMimeTypes: []string{v1.MimeTypeDockerArtifact}, ProducesMimeTypes: []string{v1.MimeTypeSBOMReport}},
			},
		},
	}
	mock.OnAnything(suite.sbomManager, "GetBy").Return([]*sbomModel.Report{{UUID: "uuid"}}, nil).Once()
	mock.OnAnything(suite.sbomManager, "Create").Return("uuid", nil).Once()
	mock.OnAnything(suite.sbomManager, "Delete").Return(nil).Once()
	mock.OnAnything(suite.taskMgr, "ListScanTasksByReportUUID").Return([]*task.Task{{Status: "Success"}}, nil)
	mock.OnAnything(suite.taskMgr, "IsTaskFinished").Return(true).Once()
	mock.OnAnything(suite.artifactCtl, "Get").Return(art, nil)
	mock.OnAnything(suite.artifactCtl, "Delete").Return(nil)
	rps, err := suite.handler.MakePlaceHolder(ctx, art, r)
	require.NoError(suite.T(), err)
	suite.Equal(1, len(rps))
}

func (suite *SBOMTestSuite) TestGetSBOMSummary() {
	r := registeredScanner
	rpts := []*sbomModel.Report{
		{UUID: "rp-uuid-004", MimeType: v1.MimeTypeSBOMReport, ReportSummary: `{"scan_status":"Success", "sbom_digest": "sha256:1234567890"}`},
	}
	mock.OnAnything(suite.scannerController, "GetRegistrationByProject").Return(r, nil)
	mock.OnAnything(suite.sbomManager, "GetBy").Return(rpts, nil)
	sum, err := suite.handler.GetSummary(context.TODO(), suite.artifact, []string{v1.MimeTypeSBOMReport})
	suite.Nil(err)
	suite.NotNil(sum)
	status := sum["scan_status"]
	suite.NotNil(status)
	dgst := sum["sbom_digest"]
	suite.NotNil(dgst)
	suite.Equal("Success", status)
	suite.Equal("sha256:1234567890", dgst)
	tasks := []*task.Task{{Status: "Error"}}
	suite.taskMgr.On("ListScanTasksByReportUUID", mock.Anything, "rp-uuid-004").Return(tasks, nil).Once()
	sum2, err := suite.handler.GetSummary(context.TODO(), suite.wrongArtifact, []string{v1.MimeTypeSBOMReport})
	suite.Nil(err)
	suite.NotNil(sum2)

}

func (suite *SBOMTestSuite) TestGetReportPlaceHolder() {
	mock.OnAnything(suite.sbomManager, "GetBy").Return([]*sbomModel.Report{{UUID: "uuid"}}, nil).Once()
	mock.OnAnything(suite.artifactCtl, "GetByReference").Return(suite.artifact, nil).Twice()
	rp, err := suite.handler.GetPlaceHolder(nil, "repo", "digest", "scannerUUID", "mimeType")
	require.NoError(suite.T(), err)
	suite.Equal("uuid", rp.UUID)
	mock.OnAnything(suite.sbomManager, "GetBy").Return(nil, nil).Once()
	rp, err = suite.handler.GetPlaceHolder(nil, "repo", "digest", "scannerUUID", "mimeType")
	require.Error(suite.T(), err)
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, &SBOMTestSuite{})
}
