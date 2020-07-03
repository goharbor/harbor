package preheat

import (
	"context"
	"errors"
	"time"

	"github.com/goharbor/harbor/src/lib/q"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/instance"
	providerModels "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
)

var (
	// Ctl is a global preheat controller instance
	Ctl = NewController()
)

// ErrorConflict for handling conflicts
var ErrorConflict = errors.New("resource conflict")

// ErrorUnhealthy for unhealthy
var ErrorUnhealthy = errors.New("instance unhealthy")

// Controller defines related top interfaces to handle the workflow of
// the image distribution.
// TODO: Add health check API
type Controller interface {
	// Get all the supported distribution providers
	//
	// If succeed, an metadata of provider list will be returned.
	// Otherwise, a non nil error will be returned
	//
	GetAvailableProviders() ([]*provider.Metadata, error)

	// CountInstance all the setup instances of distribution providers
	//
	// params *q.Query : parameters for querying
	//
	// If succeed, matched provider instance count will be returned.
	// Otherwise, a non nil error will be returned
	//
	CountInstance(ctx context.Context, query *q.Query) (int64, error)

	// ListInstance all the setup instances of distribution providers
	//
	// params *q.Query : parameters for querying
	//
	// If succeed, matched provider instance list will be returned.
	// Otherwise, a non nil error will be returned
	//
	ListInstance(ctx context.Context, query *q.Query) ([]*providerModels.Instance, error)

	// GetInstance the metadata of the specified instance
	//
	// id string : ID of the instance being deleted
	//
	// If succeed, the metadata with nil error are returned
	// Otherwise, a non nil error is returned
	//
	GetInstance(ctx context.Context, id int64) (*providerModels.Instance, error)

	// Create a new instance for the specified provider.
	//
	// If succeed, the ID of the instance will be returned.
	// Any problems met, a non nil error will be returned.
	//
	CreateInstance(ctx context.Context, instance *providerModels.Instance) (int64, error)

	// Delete the specified provider instance.
	//
	// id string : ID of the instance being deleted
	//
	// Any problems met, a non nil error will be returned.
	//
	DeleteInstance(ctx context.Context, id int64) error

	// Update the instance with incremental way;
	// Including update the enabled flag of the instance.
	//
	// id string                     : ID of the instance being updated
	// properties ...string 				 : The properties being updated
	//
	// Any problems met, a non nil error will be returned
	//
	UpdateInstance(ctx context.Context, instance *providerModels.Instance, properties ...string) error
}

var _ Controller = (*controller)(nil)

// controller is the default implementation of Controller interface.
//
type controller struct {
	// For instance
	iManager instance.Manager
}

// NewController is constructor of controller
func NewController() Controller {
	return &controller{
		iManager: instance.Mgr,
	}
}

// GetAvailableProviders implements @Controller.GetAvailableProviders
func (cc *controller) GetAvailableProviders() ([]*provider.Metadata, error) {
	return provider.ListProviders()
}

// CountInstance implements @Controller.CountInstance
func (cc *controller) CountInstance(ctx context.Context, query *q.Query) (int64, error) {
	return cc.iManager.Count(ctx, query)
}

// List implements @Controller.ListInstance
func (cc *controller) ListInstance(ctx context.Context, query *q.Query) ([]*providerModels.Instance, error) {
	return cc.iManager.List(ctx, query)
}

// CreateInstance implements @Controller.CreateInstance
func (cc *controller) CreateInstance(ctx context.Context, instance *providerModels.Instance) (int64, error) {
	if instance == nil {
		return 0, errors.New("nil instance object provided")
	}

	// Avoid duplicated endpoint
	var query = &q.Query{
		Keywords: map[string]interface{}{
			"endpoint": instance.Endpoint,
		},
	}
	num, err := cc.iManager.Count(ctx, query)
	if err != nil {
		return 0, err
	}
	if num > 0 {
		return 0, ErrorConflict
	}

	// !WARN: Check healthy status at fronted.
	if instance.Status != "healthy" {
		return 0, ErrorUnhealthy
	}

	instance.SetupTimestamp = time.Now().Unix()

	return cc.iManager.Save(ctx, instance)
}

// Delete implements @Controller.Delete
func (cc *controller) DeleteInstance(ctx context.Context, id int64) error {
	return cc.iManager.Delete(ctx, id)
}

// Update implements @Controller.Update
func (cc *controller) UpdateInstance(ctx context.Context, instance *providerModels.Instance, properties ...string) error {
	if len(properties) == 0 {
		return errors.New("no properties provided to update")
	}

	return cc.iManager.Update(ctx, instance, properties...)
}

// Get implements @Controller.Get
func (cc *controller) GetInstance(ctx context.Context, id int64) (*providerModels.Instance, error) {
	return cc.iManager.Get(ctx, id)
}
