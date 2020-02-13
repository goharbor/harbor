package chart

import (
	chartserver "github.com/goharbor/harbor/src/pkg/chart"
	"github.com/stretchr/testify/mock"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

// FakeOpertaor ...
type FakeOpertaor struct {
	mock.Mock
}

// GetDetails ...
func (f *FakeOpertaor) GetDetails(content []byte) (*chartserver.VersionDetails, error) {
	args := f.Called()
	var chartDetails *chartserver.VersionDetails
	if args.Get(0) != nil {
		chartDetails = args.Get(0).(*chartserver.VersionDetails)
	}
	return chartDetails, args.Error(1)
}

// GetData ...
func (f *FakeOpertaor) GetData(content []byte) (*chart.Chart, error) {
	args := f.Called()
	var chartData *chart.Chart
	if args.Get(0) != nil {
		chartData = args.Get(0).(*chart.Chart)
	}
	return chartData, args.Error(1)
}
