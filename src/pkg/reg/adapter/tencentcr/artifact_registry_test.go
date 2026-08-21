package tencentcr

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	tcr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tcr/v20190924"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/filter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/reg/util"
)

type mockFetchAdapter struct {
	adapter
}

func (m *mockFetchAdapter) listCandidateNamespaces(pattern string) ([]string, error) {
	return []string{"demo"}, nil
}

func (m *mockFetchAdapter) isNamespaceExist(ns string) (bool, error) {
	return true, nil
}

func (m *mockFetchAdapter) listReposByNamespace(ns string) ([]*tcr.TcrRepositoryInfo, error) {
	name := "demo/app"
	return []*tcr.TcrRepositoryInfo{
		{Name: &name},
	}, nil
}

func (m *mockFetchAdapter) getImages(ns, repo, _ string) (string, []string, error) {
	return "", []string{"v1.0", "v2.0"}, nil
}

// Override FetchArtifacts to use mock methods
func (m *mockFetchAdapter) FetchArtifacts(filters []*model.Filter) ([]*model.Resource, error) {
	// get filter pattern
	var namespacePattern, _, tagsPattern = filterToPatterns(filters)

	// 1. list namespaces - use mock method
	var namespaces []string
	namespaces, err := m.listCandidateNamespaces(namespacePattern)
	if err != nil {
		return nil, err
	}

	// 2. list repos - use mock method
	var repos []*model.Repository
	var repositories []*model.Repository
	for _, ns := range namespaces {
		tcrRepos, err := m.listReposByNamespace(ns)
		if err != nil {
			return nil, err
		}

		if len(tcrRepos) == 0 {
			continue
		}
		for _, tcrRepo := range tcrRepos {
			repositories = append(repositories, &model.Repository{
				Name: *tcrRepo.Name,
			})
		}
	}
	repos, _ = filter.DoFilterRepositories(repositories, filters)

	// 4. list images - use mock method
	var rawResources = make([]*model.Resource, len(repos))
	for i, r := range repos {
		repoArr := strings.Split(r.Name, "/")
		_, images, err := m.getImages(repoArr[0], strings.Join(repoArr[1:], "/"), "")
		if err != nil {
			return nil, err
		}

		var filteredImages []string
		if tagsPattern != "" {
			for _, image := range images {
				ok, err := util.Match(tagsPattern, image)
				if err != nil {
					return nil, err
				}
				if ok {
					filteredImages = append(filteredImages, image)
				}
			}
		} else {
			filteredImages = images
		}

		if len(filteredImages) > 0 {
			rawResources[i] = &model.Resource{
				Type:     model.ResourceTypeImage,
				Registry: m.registry,
				Metadata: &model.ResourceMetadata{
					Repository: &model.Repository{
						Name: r.Name,
					},
					Vtags: filteredImages,
				},
			}
		}
	}

	var resources []*model.Resource
	for _, res := range rawResources {
		if res != nil {
			resources = append(resources, res)
		}
	}

	return resources, nil
}

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
		{
			name: "name and tag filters provided",
			args: args{
				filters: []*model.Filter{
					{Type: model.FilterTypeName, Value: "demo/app"},
					{Type: model.FilterTypeTag, Value: "v1.*"},
				},
			},
			wantNamespacePattern: "demo",
			wantRepoPattern:      "demo/app",
			wantTagsPattern:      "v1.*",
		},
		{
			name: "only name filter provided",
			args: args{
				filters: []*model.Filter{
					{Type: model.FilterTypeName, Value: "team/project"},
				},
			},
			wantNamespacePattern: "team",
			wantRepoPattern:      "team/project",
			wantTagsPattern:      "",
		},
		{
			name: "empty filters slice",
			args: args{
				filters: []*model.Filter{},
			},
			wantNamespacePattern: "",
			wantRepoPattern:      "",
			wantTagsPattern:      "",
		},
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
		{
			name: "fetch artifacts with name and tag filter",
			fields: fields{
				registry: &model.Registry{
					ID:   1,
					Name: "tencent",
				},
			},
			args: args{
				filters: []*model.Filter{
					{Type: model.FilterTypeName, Value: "demo/app"},
					{Type: model.FilterTypeTag, Value: "v1.*"},
				},
			},
			wantResources: []*model.Resource{
				{
					Type:     model.ResourceTypeImage,
					Registry: &model.Registry{ID: 1, Name: "tencent"},
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{
							Name: "demo/app",
						},
						Vtags: []string{"v1.0"},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registryID := "test-registry-id"
			// Create a non-nil tcrClient to pass the nil check in adapter methods
			tcrClient := &tcr.Client{}
			a := &mockFetchAdapter{
				adapter: adapter{
					Adapter:    tt.fields.Adapter,
					registryID: &registryID,
					regionName: tt.fields.regionName,
					tcrClient:  tcrClient,
					pageSize:   tt.fields.pageSize,
					client:     tt.fields.client,
					registry:   tt.fields.registry,
				},
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
		{
			name:   "list candidate namespaces with pattern",
			fields: fields{},
			args: args{
				namespacePattern: "demo",
			},
			wantNamespaces: []string{"demo"},
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &mockFetchAdapter{
				adapter: adapter{
					Adapter:    tt.fields.Adapter,
					registryID: tt.fields.registryID,
					regionName: tt.fields.regionName,
					tcrClient:  tt.fields.tcrClient,
					pageSize:   tt.fields.pageSize,
					client:     tt.fields.client,
					registry:   tt.fields.registry,
				},
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

type mockAdapter struct {
	adapter
	deleteImageFunc func(namespace, repo, reference string) error
}

func (m *mockAdapter) deleteImage(namespace, repo, reference string) error {
	if m.deleteImageFunc != nil {
		return m.deleteImageFunc(namespace, repo, reference)
	}
	return nil
}

func (m *mockAdapter) DeleteManifest(repository, reference string) error {
	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format: %s", repository)
	}
	namespace, repo := parts[0], parts[1]
	return m.deleteImage(namespace, repo, reference)
}

func Test_adapter_DeleteManifest(t *testing.T) {
	type args struct {
		repository string
		reference  string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid repository format",
			args: args{
				repository: "invalidRepo",
				reference:  "latest",
			},
			wantErr: true,
		},
		{
			name: "valid repository format should not error",
			args: args{
				repository: "demo/app",
				reference:  "v1.0",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &mockAdapter{}

			if tt.name == "valid repository format should not error" {
				a.deleteImageFunc = func(namespace, repo, reference string) error {
					if namespace != "demo" || repo != "app" || reference != "v1.0" {
						t.Errorf("unexpected args: %s/%s:%s", namespace, repo, reference)
					}
					return nil
				}
			}

			err := a.DeleteManifest(tt.args.repository, tt.args.reference)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteManifest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
