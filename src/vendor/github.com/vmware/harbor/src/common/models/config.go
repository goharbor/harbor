// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package models

/*
// Authentication ...
type Authentication struct {
	Mode             string `json:"mode"`
	SelfRegistration bool   `json:"self_registration"`
	LDAP             *LDAP  `json:"ldap,omitempty"`
}
*/

// Database ...
type Database struct {
	Type       string      `json:"type"`
	PostGreSQL *PostGreSQL `json:"postgresql,omitempty"`
}

// MySQL ...
type MySQL struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Database string `json:"database"`
}

// SQLite ...
type SQLite struct {
	File string `json:"file"`
}

// PostGreSQL ...
type PostGreSQL struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Database string `json:"database"`
}

// Email ...
type Email struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSL      bool   `json:"ssl"`
	Identity string `json:"identity"`
	From     string `json:"from"`
	Insecure bool   `json:"insecure"`
}

/*
// Registry ...
type Registry struct {
	URL string `json:"url"`
}

// TokenService ...
type TokenService struct {
	URL string `json:"url"`
}

// SystemCfg holds all configurations of system
type SystemCfg struct {
	DomainName                 string          `json:"domain_name"` // Harbor external URL: protocal://host:port
	Authentication             *Authentication `json:"authentication"`
	Database                   *Database       `json:"database"`
	TokenService               *TokenService   `json:"token_service"`
	Registry                   *Registry       `json:"registry"`
	Email                      *Email          `json:"email"`
	VerifyRemoteCert           bool            `json:"verify_remote_cert"`
	ProjectCreationRestriction string          `json:"project_creation_restriction"`
	MaxJobWorkers              int             `json:"max_job_workers"`
	JobLogDir                  string          `json:"job_log_dir"`
	InitialAdminPwd            string          `json:"initial_admin_pwd,omitempty"`
	TokenExpiration            int             `json:"token_expiration"` // in minute
	SecretKey                  string          `json:"secret_key,omitempty"`
	CfgExpiration              int             `json:"cfg_expiration"`
}
*/

// ConfigEntry ...
type ConfigEntry struct {
	ID    int64  `orm:"pk;auto;column(id)" json:"-"`
	Key   string `orm:"column(k)" json:"k"`
	Value string `orm:"column(v)" json:"v"`
}

// TableName ...
func (ce *ConfigEntry) TableName() string {
	return "properties"
}
