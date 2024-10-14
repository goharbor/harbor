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

import "time"

const (
	registryEndpointTpl = "https://registry.%s.aliyuncs.com"
	endpointTpl         = "cr.%s.aliyuncs.com"

	registryACRService = "registry.aliyuncs.com"
)

type registryServiceInfo struct {
	IsACREE    bool
	RegionID   string
	InstanceID string
}

type authorizationToken struct {
	Data struct {
		ExpireDate         timeUnix `json:"expireDate"`
		AuthorizationToken string   `json:"authorizationToken"`
		TempUserName       string   `json:"tempUserName"`
	} `json:"data"`
	RequestID string `json:"requestId"`
}

type timeUnix int64

func (t timeUnix) ToTime() time.Time {
	return time.Unix(int64(t)/1000, 0)
}

func (t timeUnix) String() string {
	return t.ToTime().String()
}

type aliACRNamespaceResp struct {
	Data struct {
		Namespaces []aliACRNamespace `json:"namespaces"`
	} `json:"data"`
	RequestID string `json:"requestId"`
}

type aliACRNamespace struct {
	Namespace       string `json:"namespace"`
	AuthorizeType   string `json:"authorizeType"`
	NamespaceStatus string `json:"namespaceStatus"`
}

type aliReposResp struct {
	Data struct {
		Page     int       `json:"page"`
		Total    int       `json:"total"`
		PageSize int       `json:"pageSize"`
		Repos    []aliRepo `json:"repos"`
	} `json:"data"`
	RequestID string `json:"requestId"`
}

type aliRepo struct {
	Summary        string `json:"summary"`
	RegionID       string `json:"regionId"`
	RepoName       string `json:"repoName"`
	RepoNamespace  string `json:"repoNamespace"`
	RepoStatus     string `json:"repoStatus"`
	RepoID         int    `json:"repoId"`
	RepoType       string `json:"repoType"`
	RepoBuildType  string `json:"repoBuildType"`
	GmtCreate      int64  `json:"gmtCreate"`
	RepoOriginType string `json:"repoOriginType"`
	GmtModified    int64  `json:"gmtModified"`
	RepoDomainList struct {
		Internal string `json:"internal"`
		Public   string `json:"public"`
		Vpc      string `json:"vpc"`
	} `json:"repoDomainList"`
	Downloads         int    `json:"downloads"`
	RepoAuthorizeType string `json:"repoAuthorizeType"`
	Logo              string `json:"logo"`
	Stars             int    `json:"stars"`
}

type aliTagResp struct {
	Data struct {
		Total    int `json:"total"`
		PageSize int `json:"pageSize"`
		Page     int `json:"page"`
		Tags     []struct {
			ImageUpdate int64  `json:"imageUpdate"`
			ImageID     string `json:"imageId"`
			Digest      string `json:"digest"`
			ImageSize   int    `json:"imageSize"`
			Tag         string `json:"tag"`
			ImageCreate int64  `json:"imageCreate"`
			Status      string `json:"status"`
		} `json:"tags"`
	} `json:"data"`
	RequestID string `json:"requestId"`
}
