package instance

import (
	"encoding/json"
	"errors"

	"github.com/goharbor/harbor/src/lib/q"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/dao"
	daomodels "github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/models"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models"
)

// Manager is responsible for storing the instances
type Manager interface {
	// Save the instance metadata to the backend store
	//
	// inst *Metadata : a ptr of instance
	//
	// If succeed, the uuid of the saved instance is returned;
	// otherwise, a non nil error is returned
	//
	Save(inst *models.Metadata) (int64, error)

	// Delete the specified instance
	//
	// id int64 : the id of the instance
	//
	// If succeed, a nil error is returned;
	// otherwise, a non nil error is returned
	//
	Delete(id int64) error

	// Update the specified instance
	//
	// inst *Metadata : a ptr of instance
	//
	// If succeed, a nil error is returned;
	// otherwise, a non nil error is returned
	//
	Update(inst *models.Metadata) error

	// Get the instance with the ID
	//
	// id int64 : the id of the instance
	//
	// If succeed, a non nil Metadata is returned;
	// otherwise, a non nil error is returned
	//
	Get(id int64) (*models.Metadata, error)

	// Query the instances by the param
	//
	// query *q.Query : the query params
	//
	// If succeed, an instance metadata list is returned;
	// otherwise, a non nil error is returned
	//
	List(query *q.Query) (int64, []*models.Metadata, error)
}

// DefaultManager implement the Manager interface
type DefaultManager struct{}

// NewDefaultManager returns an instance of DefaultManger
func NewDefaultManager() *DefaultManager {
	return &DefaultManager{}
}

// Ensure *DefaultManager has implemented Manager interface.
var _ Manager = (*DefaultManager)(nil)

var (
	errNilMetadataModel = errors.New("nil instance metadata model")
)

// Save implements @Manager.Save
func (dm *DefaultManager) Save(inst *models.Metadata) (int64, error) {
	if inst == nil {
		return 0, errors.New("nil instance metadata")
	}

	instance, err := convertToDaoModel(inst)
	if err != nil {
		return 0, err
	}

	id, err := dao.AddInstance(instance)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func convertToDaoModel(inst *models.Metadata) (*daomodels.Instance, error) {
	if inst == nil {
		return nil, errNilMetadataModel
	}

	instance := &daomodels.Instance{
		ID:             inst.ID,
		Name:           inst.Name,
		Description:    inst.Description,
		Provider:       inst.Provider,
		Endpoint:       inst.Endpoint,
		AuthMode:       inst.AuthMode,
		AuthData:       mapToString(inst.AuthData),
		Status:         inst.Status,
		Enabled:        inst.Enabled,
		SetupTimestamp: inst.SetupTimestamp,
		Extensions:     mapToString(inst.Extensions),
	}

	return instance, nil
}

func convertFromDaoModel(inst *daomodels.Instance) (*models.Metadata, error) {
	if inst == nil {
		return nil, errNilMetadataModel
	}

	instance := &models.Metadata{
		ID:             inst.ID,
		Name:           inst.Name,
		Description:    inst.Description,
		Provider:       inst.Provider,
		Endpoint:       inst.Endpoint,
		AuthMode:       inst.AuthMode,
		AuthData:       mapFromString(inst.AuthData),
		Status:         inst.Status,
		Enabled:        inst.Enabled,
		SetupTimestamp: inst.SetupTimestamp,
		Extensions:     mapFromString(inst.Extensions),
	}

	return instance, nil
}

// Delete implements @Manager.Delete
func (dm *DefaultManager) Delete(id int64) error {
	return dao.DeleteInstance(id)
}

// Update implements @Manager.Update
func (dm *DefaultManager) Update(inst *models.Metadata) error {
	if inst == nil {
		return errors.New("nil instance metadata")
	}

	instance, err := convertToDaoModel(inst)
	if err != nil {
		return err
	}

	return dao.UpdateInstance(instance)
}

// Get implements @Manager.Get
func (dm *DefaultManager) Get(id int64) (*models.Metadata, error) {
	inst, err := dao.GetInstance(id)
	if err != nil {
		return nil, err
	}

	instance, err := convertFromDaoModel(inst)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

// List implements @Manager.List
func (dm *DefaultManager) List(query *q.Query) (int64, []*models.Metadata, error) {
	total, instances, err := dao.ListInstances(query)
	if err != nil {
		return 0, nil, err
	}

	var results []*models.Metadata
	for _, inst := range instances {
		if ins, err := convertFromDaoModel(inst); err == nil {
			results = append(results, ins)
		}
	}

	return total, results, nil
}

func mapToString(m map[string]string) string {
	result, _ := json.Marshal(m)
	return string(result)
}

func mapFromString(s string) map[string]string {
	result := make(map[string]string)
	json.Unmarshal([]byte(s), &result)
	return result
}
