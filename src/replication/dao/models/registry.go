package models

import (
	"time"

	"github.com/astaxie/beego/validation"
	"github.com/goharbor/harbor/src/common/utils"
)

const (
	// RegistryTable is the table name for registry
	RegistryTable = "registry"
)

// Registry is the model for a registry, which wraps the endpoint URL and credential of a remote registry.
type Registry struct {
	ID             int64     `orm:"pk;auto;column(id)" json:"id"`
	URL            string    `orm:"column(url)" json:"endpoint"`
	Name           string    `orm:"column(name)" json:"name"`
	CredentialType string    `orm:"column(credential_type);default(basic)" json:"credential_type"`
	AccessKey      string    `orm:"column(access_key)" json:"access_key"`
	AccessSecret   string    `orm:"column(access_secret)" json:"access_secret"`
	Type           string    `orm:"column(type)" json:"type"`
	Insecure       bool      `orm:"column(insecure)" json:"insecure"`
	Description    string    `orm:"column(description)" json:"description"`
	Health         string    `orm:"column(health)" json:"health"`
	CreationTime   time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime     time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName is required by by beego orm to map Registry to table registry
func (r *Registry) TableName() string {
	return RegistryTable
}

// Valid ...
func (r *Registry) Valid(v *validation.Validation) {
	if len(r.Name) == 0 {
		v.SetError("name", "can not be empty")
	}

	if len(r.Name) > 64 {
		v.SetError("name", "max length is 64")
	}

	url, err := utils.ParseEndpoint(r.URL)
	if err != nil {
		v.SetError("endpoint", err.Error())
	} else {
		// Prevent SSRF security issue #3755
		r.URL = url.Scheme + "://" + url.Host + url.Path
		if len(r.URL) > 64 {
			v.SetError("endpoint", "max length is 64")
		}
	}
}
