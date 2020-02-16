package instance

import (
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/models"
)

// Storage is responsible for storing the instances
type Storage interface {
	// Save the instance metadata to the backend store
	//
	// inst *Metadata : a ptr of instance
	//
	// If succeed, the uuid of the saved instance is returned;
	// otherwise, a non nil error is returned
	//
	Save(inst *models.Metadata) (string, error)

	// Delete the specified instance
	//
	// id string : the uuid of the instance
	//
	// If succeed, a nil error is returned;
	// otherwise, a non nil error is returned
	//
	Delete(id string) error

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
	// id string : the uuid of the instance
	//
	// If succeed, a non nil Metadata is returned;
	// otherwise, a non nil error is returned
	//
	Get(id string) (*models.Metadata, error)

	// Query the instacnes by the param
	//
	// param *models.QueryParam : the query params
	//
	// If succeed, an instance metadata list is returned;
	// otherwise, a non nil error is returned
	//
	List(param *models.QueryParam) ([]*models.Metadata, error)
}
