package aliacr

import "time"

const (
	defaultTemporaryTokenExpiredTime = time.Hour * 1
	registryEndpointTpl              = "https://registry.%s.aliyuncs.com"
	endpointTpl                      = "cr.%s.aliyuncs.com"
)

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

type aliRepoResp struct {
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
