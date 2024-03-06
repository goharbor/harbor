package sts

import "github.com/volcengine/volc-sdk-golang/base"

// AssumeRole
type AssumeRoleResp struct {
	ResponseMetadata base.ResponseMetadata
	Result           *AssumeRoleResult `json:",omitempty"`
}

type AssumeRoleResult struct {
	Credentials     *Credentials
	AssumedRoleUser *AssumeRoleUser
}

type AssumeRoleRequest struct {
	DurationSeconds int
	Policy          string
	RoleTrn         string
	RoleSessionName string
}

type AssumeRoleUser struct {
	Trn           string
	AssumedRoleId string
}

type Credentials struct {
	CurrentTime     string
	ExpiredTime     string
	AccessKeyId     string
	SecretAccessKey string
	SessionToken    string
}

// AssumeRole
