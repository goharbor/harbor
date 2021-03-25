package tencentcr

import (
	"reflect"
	"testing"
)

func Test_adapter_createPrivateNamespace(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		wantErr   bool
	}{
		{namespace: "ut_ns_123", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &adapter{tcrClient: tcrClient}
			if err := a.createPrivateNamespace(tt.namespace); (err != nil) != tt.wantErr {
				t.Errorf("adapter.createPrivateNamespace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_adapter_createRepository(t *testing.T) {
	tests := []struct {
		name       string
		namespace  string
		repository string
		wantErr    bool
	}{
		{namespace: "ut_ns_123", repository: "ut_repo_123", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &adapter{tcrClient: tcrClient}
			if err := a.createRepository(tt.namespace, tt.repository); (err != nil) != tt.wantErr {
				t.Errorf("adapter.createRepository() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_adapter_listNamespaces(t *testing.T) {
	tests := []struct {
		name           string
		wantNamespaces []string
		wantErr        bool
	}{
		{wantNamespaces: []string{}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &adapter{tcrClient: tcrClient}
			_, err := a.listNamespaces()
			if (err != nil) != tt.wantErr {
				t.Errorf("adapter.listNamespaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_adapter_isNamespaceExist(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		wantExist bool
		wantErr   bool
	}{
		{namespace: "ut_ns_123", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &adapter{tcrClient: tcrClient}
			gotExist, err := a.isNamespaceExist(tt.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("adapter.isNamespaceExist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotExist != tt.wantExist {
				t.Errorf("adapter.isNamespaceExist() = %v, want %v", gotExist, tt.wantExist)
			}
		})
	}
}

func Test_adapter_listReposByNamespace(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		wantErr   bool
	}{
		{namespace: "ut_ns_123", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &adapter{tcrClient: tcrClient}
			_, err := a.listReposByNamespace(tt.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("adapter.listReposByNamespace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_adapter_getImages(t *testing.T) {
	tests := []struct {
		name       string
		namespace  string
		repo       string
		tag        string
		wantImages []string
		wantErr    bool
	}{
		{namespace: "ut_ns_123", repo: "ut_repo_123", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &adapter{tcrClient: tcrClient}
			_, gotImages, err := a.getImages(tt.namespace, tt.repo, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("adapter.getImages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotImages, tt.wantImages) {
				t.Errorf("adapter.getImages() = %v, want %v", gotImages, tt.wantImages)
			}
		})
	}
}
