package tencentcr

import (
	"reflect"
	"testing"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	tcr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tcr/v20190924"
)

func Test_filterToPatterns(t *testing.T) {
	type args struct {
		filters []*model.Filter
	}
	tests := []struct {
		name                 string
		args                 args
		wantNamespacePattern string
		wantRepoPattern      string
		wantTagsPattern      string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNamespacePattern, gotRepoPattern, gotTagsPattern := filterToPatterns(tt.args.filters)
			if gotNamespacePattern != tt.wantNamespacePattern {
				t.Errorf("filterToPatterns() gotNamespacePattern = %v, want %v", gotNamespacePattern, tt.wantNamespacePattern)
			}
			if gotRepoPattern != tt.wantRepoPattern {
				t.Errorf("filterToPatterns() gotRepoPattern = %v, want %v", gotRepoPattern, tt.wantRepoPattern)
			}
			if gotTagsPattern != tt.wantTagsPattern {
				t.Errorf("filterToPatterns() gotTagsPattern = %v, want %v", gotTagsPattern, tt.wantTagsPattern)
			}
		})
	}
}

func Test_adapter_FetchArtifacts(t *testing.T) {
	type fields struct {
		Adapter    *native.Adapter
		registryID *string
		regionName *string
		tcrClient  *tcr.Client
		pageSize   *int64
		client     *commonhttp.Client
		registry   *model.Registry
	}
	type args struct {
		filters []*model.Filter
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantResources []*model.Resource
		wantErr       bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &adapter{
				Adapter:    tt.fields.Adapter,
				registryID: tt.fields.registryID,
				regionName: tt.fields.regionName,
				tcrClient:  tt.fields.tcrClient,
				pageSize:   tt.fields.pageSize,
				client:     tt.fields.client,
				registry:   tt.fields.registry,
			}
			gotResources, err := a.FetchArtifacts(tt.args.filters)
			if (err != nil) != tt.wantErr {
				t.Errorf("adapter.FetchArtifacts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResources, tt.wantResources) {
				t.Errorf("adapter.FetchArtifacts() = %v, want %v", gotResources, tt.wantResources)
			}
		})
	}
}

func Test_adapter_listCandidateNamespaces(t *testing.T) {
	type fields struct {
		Adapter    *native.Adapter
		registryID *string
		regionName *string
		tcrClient  *tcr.Client
		pageSize   *int64
		client     *commonhttp.Client
		registry   *model.Registry
	}
	type args struct {
		namespacePattern string
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantNamespaces []string
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &adapter{
				Adapter:    tt.fields.Adapter,
				registryID: tt.fields.registryID,
				regionName: tt.fields.regionName,
				tcrClient:  tt.fields.tcrClient,
				pageSize:   tt.fields.pageSize,
				client:     tt.fields.client,
				registry:   tt.fields.registry,
			}
			gotNamespaces, err := a.listCandidateNamespaces(tt.args.namespacePattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("adapter.listCandidateNamespaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotNamespaces, tt.wantNamespaces) {
				t.Errorf("adapter.listCandidateNamespaces() = %v, want %v", gotNamespaces, tt.wantNamespaces)
			}
		})
	}
}

func Test_adapter_DeleteManifest(t *testing.T) {
	type fields struct {
		Adapter    *native.Adapter
		registryID *string
		regionName *string
		tcrClient  *tcr.Client
		pageSize   *int64
		client     *commonhttp.Client
		registry   *model.Registry
	}
	type args struct {
		repository string
		reference  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &adapter{
				Adapter:    tt.fields.Adapter,
				registryID: tt.fields.registryID,
				regionName: tt.fields.regionName,
				tcrClient:  tt.fields.tcrClient,
				pageSize:   tt.fields.pageSize,
				client:     tt.fields.client,
				registry:   tt.fields.registry,
			}
			if err := a.DeleteManifest(tt.args.repository, tt.args.reference); (err != nil) != tt.wantErr {
				t.Errorf("adapter.DeleteManifest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
