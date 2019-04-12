package dao

import (
	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/replication/dao/models"
)

// ListRegistryQuery defines the query conditions to list registry.
type ListRegistryQuery struct {
	// Query is name query
	Query string
	// Offset specifies the offset in the registry list to return
	Offset int64
	// Limit specifies the maximum registries to return
	Limit int64
}

// AddRegistry add a new registry
func AddRegistry(registry *models.Registry) (int64, error) {
	o := dao.GetOrmer()
	return o.Insert(registry)
}

// GetRegistry gets one registry from database by id.
func GetRegistry(id int64) (*models.Registry, error) {
	o := dao.GetOrmer()
	r := models.Registry{ID: id}
	err := o.Read(&r, "ID")
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &r, err
}

// GetRegistryByName gets one registry from database by its name.
func GetRegistryByName(name string) (*models.Registry, error) {
	o := dao.GetOrmer()
	r := models.Registry{Name: name}
	err := o.Read(&r, "Name")
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &r, err
}

// GetRegistryByURL gets one registry from database by its URL.
func GetRegistryByURL(url string) (*models.Registry, error) {
	o := dao.GetOrmer()
	r := models.Registry{URL: url}
	err := o.Read(&r, "URL")
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &r, err
}

// ListRegistries lists registries. Registries returned are sorted by creation time.
// - query: query to the registry name, name query and pagination are defined.
func ListRegistries(query ...*ListRegistryQuery) (int64, []*models.Registry, error) {
	o := dao.GetOrmer()

	q := o.QueryTable(&models.Registry{})
	if len(query) > 0 && len(query[0].Query) > 0 {
		q = q.Filter("name__contains", query[0].Query)
	}

	total, err := q.Count()
	if err != nil {
		return -1, nil, err
	}

	// limit being -1 means no pagination specified.
	if len(query) > 0 && query[0].Limit != -1 {
		q = q.Offset(query[0].Offset).Limit(query[0].Limit)
	}
	var registries []*models.Registry
	_, err = q.All(&registries)
	if err != nil {
		return total, nil, err
	}
	if registries == nil {
		registries = []*models.Registry{}
	}
	return total, registries, nil
}

// UpdateRegistry updates one registry
func UpdateRegistry(registry *models.Registry, props ...string) error {
	o := dao.GetOrmer()

	_, err := o.Update(registry, props...)
	return err
}

// DeleteRegistry deletes a registry
func DeleteRegistry(id int64) error {
	o := dao.GetOrmer()
	_, err := o.Delete(&models.Registry{ID: id})
	return err
}
