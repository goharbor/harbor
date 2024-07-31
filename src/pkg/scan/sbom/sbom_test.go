package sbom

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/testing/jobservice"

	"github.com/stretchr/testify/suite"
)

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
			got, err := v.ReportURLParameter(tt.args.in0)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReportURLParameter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReportURLParameter() got = %v, want %v", got, tt.want)
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

func mockGetRegistry(ctx context.Context) string {
	return "myharbor.example.com"
}

func mockGenAccessory(scanRep v1.ScanRequest, sbomContent []byte, labels map[string]string, mediaType string, robot *model.Robot) (string, error) {
	return "sha256:1234567890", nil
}

type ExampleTestSuite struct {
	handler *scanHandler
	suite.Suite
}

func (suite *ExampleTestSuite) SetupSuite() {
	suite.handler = &scanHandler{
		GenAccessoryFunc: mockGenAccessory,
		RegistryServer:   mockGetRegistry,
	}
}

func (suite *ExampleTestSuite) TearDownSuite() {
}

func (suite *ExampleTestSuite) TestPostScan() {
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

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, &ExampleTestSuite{})
}
