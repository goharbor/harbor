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

package aliacr

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/reg/model"
)

var acrAdapter *adapter
var acreeAdapter *adapter

func init() {
	accessKey := os.Getenv("ALIYUN_ACCESS_KEY")
	accessSecret := os.Getenv("ALIYUN_ACCESS_SECRET")
	acrEndpoint := os.Getenv("ALIYUN_ACR_ENDPOINT")
	acreeEndpoint := os.Getenv("ALIYUN_ACREE_ENDPOINT")

	if accessKey == "" || accessSecret == "" {
		return
	}
	if acrEndpoint != "" {
		acrAdapter, _ = newAdapter(&model.Registry{
			URL: acrEndpoint,
			Credential: &model.Credential{
				AccessKey:    accessKey,
				AccessSecret: accessSecret,
			},
		})
	}

	if acreeEndpoint != "" {
		acreeAdapter, _ = newAdapter(&model.Registry{
			URL: acreeEndpoint,
			Credential: &model.Credential{
				AccessKey:    accessKey,
				AccessSecret: accessSecret,
			},
		})
	}
}

func skipCheck(adapter *adapter) bool {
	return adapter == nil
}

func Test_ACREE_ListNamespace(t *testing.T) {
	if skipCheck(acreeAdapter) {
		t.Skip("skip test acree ListNamespace")
	}
	testcases := []struct {
		name            string
		wantedNamespace []string
	}{
		{
			name:            "test list namespace",
			wantedNamespace: []string{"ut_acree_namespace"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			namespaces, err := acreeAdapter.acrAPI.ListNamespace()
			if err != nil {
				t.Errorf("ListNamespace error: %v", err)
			}
			if !reflect.DeepEqual(namespaces, tc.wantedNamespace) {
				t.Errorf("ListNamespace error, wants=%v, actual=%v", tc.wantedNamespace, namespaces)
			}
		})
	}
}

func Test_ACREE_ListRepositoryAndTag(t *testing.T) {
	if skipCheck(acreeAdapter) {
		t.Skip("skip test acree ListRepository")
	}
	testcases := []struct {
		name       string
		namespace  string
		wantedRepo []string
		wantedTag  []string
	}{
		{
			name:       "test list repository",
			namespace:  "ut_acree_namespace",
			wantedRepo: []string{"ut_acree_repo"},
			wantedTag:  []string{"latest"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// List repository
			repos, err := acreeAdapter.acrAPI.ListRepository(tc.namespace)
			if err != nil {
				t.Errorf("ListRepository error: %v", err)
			}
			var actualRepo []string
			for _, repo := range repos {
				actualRepo = append(actualRepo, repo.Name)
			}
			if !reflect.DeepEqual(actualRepo, tc.wantedRepo) {
				t.Errorf("ListRepository error, wants=%v, actual=%v", tc.wantedRepo, actualRepo)
			}

			// List tag
			var actualTag []string
			for _, repo := range repos {
				tags, err := acreeAdapter.acrAPI.ListRepoTag(repo)
				if err != nil {
					t.Errorf("ListRepoTag error: %v", err)
				}
				actualTag = append(actualTag, tags...)
			}
			if !reflect.DeepEqual(actualTag, tc.wantedTag) {
				t.Errorf("ListRepoTag error, wants=%v, actual=%v", tc.wantedTag, actualTag)
			}
		})
	}
}

func Test_ACREE_GetAuthorizationToken(t *testing.T) {
	if skipCheck(acreeAdapter) {
		t.Skip("skip test acree GetAuthorizationToken")
	}
	testcases := []struct {
		name string
	}{
		{
			name: "test get authorization token",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := acreeAdapter.acrAPI.GetAuthorizationToken()
			if err != nil {
				t.Errorf("GetAuthorizationToken error: %v", err)
			}
			if token == nil {
				t.Errorf("GetAuthorizationToken error, token is nil")
			}
			if !time.Now().Before(token.expiresAt) {
				t.Errorf("GetAuthorizationToken error, token expiresAt is not valid")
			}
		})
	}
}

func Test_ACR_ListNamespace(t *testing.T) {
	if skipCheck(acrAdapter) {
		t.Skip("skip test acr ListNamespace")
	}
	testcases := []struct {
		name            string
		wantedNamespace []string
	}{
		{
			name:            "test list namespace",
			wantedNamespace: []string{"ut_acr_namespace"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			namespaces, err := acrAdapter.acrAPI.ListNamespace()
			if err != nil {
				t.Errorf("ListNamespace error: %v", err)
			}
			if !reflect.DeepEqual(namespaces, tc.wantedNamespace) {
				t.Errorf("ListNamespace error, wants=%v, actual=%v", tc.wantedNamespace, namespaces)
			}
		})
	}
}

func Test_ACR_ListRepositoryAndTag(t *testing.T) {
	if skipCheck(acrAdapter) {
		t.Skip("skip test acr ListRepository")
	}
	testcases := []struct {
		name       string
		namespace  string
		wantedRepo []string
		wantedTag  []string
	}{
		{
			name:       "test list repository",
			namespace:  "ut_acr_namespace",
			wantedRepo: []string{"ut_acr_repo"},
			wantedTag:  []string{"latest"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// List repository
			repos, err := acrAdapter.acrAPI.ListRepository(tc.namespace)
			if err != nil {
				t.Errorf("ListRepository error: %v", err)
			}
			var actualRepo []string
			for _, repo := range repos {
				actualRepo = append(actualRepo, repo.Name)
			}
			if !reflect.DeepEqual(actualRepo, tc.wantedRepo) {
				t.Errorf("ListRepository error, wants=%v, actual=%v", tc.wantedRepo, actualRepo)
			}
			// List tag
			var actualTag []string
			for _, repo := range repos {
				tags, err := acrAdapter.acrAPI.ListRepoTag(repo)
				if err != nil {
					t.Errorf("ListRepoTag error: %v", err)
				}
				actualTag = append(actualTag, tags...)
			}
			if !reflect.DeepEqual(actualTag, tc.wantedTag) {
				t.Errorf("ListRepoTag error, wants=%v, actual=%v", tc.wantedTag, actualTag)
			}
		})
	}
}

func Test_ACR_GetAuthorizationToken(t *testing.T) {
	if skipCheck(acrAdapter) {
		t.Skip("skip test acr GetAuthorizationToken")
	}
	testcases := []struct {
		name string
	}{
		{
			name: "test get authorization token",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := acrAdapter.acrAPI.GetAuthorizationToken()
			if err != nil {
				t.Errorf("GetAuthorizationToken error: %v", err)
			}
			if token == nil {
				t.Errorf("GetAuthorizationToken error, token is nil")
			}
			if !time.Now().Before(token.expiresAt) {
				t.Errorf("GetAuthorizationToken error, token expiresAt is not valid")
			}
		})
	}
}
