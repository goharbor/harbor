package preheat

import (
	"context"
	"errors"
	"fmt"
	"time"

	tk "github.com/docker/distribution/registry/auth/token"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/service/token"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/history"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/instance"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
)

// DefaultController is default controller
var DefaultController Controller

// ErrorConflict for handling conflicts
var ErrorConflict = errors.New("resource conflict")

// CompositePreheatingResults handle preheating results among multiple providers
// Key is the ID of the provider instance.
type CompositePreheatingResults map[int64]*[]*provider.PreheatingStatus

// Controller defines related top interfaces to handle the workflow of
// the image distribution.
type Controller interface {
	// Get all the supported distribution providers
	//
	// If succeed, an metadata of provider list will be returned.
	// Otherwise, a non nil error will be returned
	//
	GetAvailableProviders() ([]*provider.Metadata, error)

	// List all the setup instances of distribution providers
	//
	// params *models.QueryParam : parameters for querying
	//
	// If succeed, an provider instance list will be returned.
	// Otherwise, a non nil error will be returned
	//
	ListInstances(params *models.QueryParam) (int64, []*models.Metadata, error)

	// Get the metadata of the specified instance
	//
	// id string : ID of the instance being deleted
	//
	// If succeed, the metadata with nil error are returned
	// Otherwise, a non nil error is returned
	//
	GetInstance(id int64) (*models.Metadata, error)

	// Create a new instance for the specified provider.
	//
	// If succeed, the ID of the instance will be returned.
	// Any problems met, a non nil error will be returned.
	//
	CreateInstance(instance *models.Metadata) (int64, error)

	// Delete the specified provider instance.
	//
	// id string : ID of the instance being deleted
	//
	// Any problems met, a non nil error will be returned.
	//
	DeleteInstance(id int64) error

	// Update the instance with incremental way;
	// Including update the enabled flag of the instance.
	//
	// id string                     : ID of the instance being updated
	// properties models.PropertySet : The properties being updated
	//
	// Any problems met, a non nil error will be returned
	//
	UpdateInstance(id int64, properties models.PropertySet) error

	// Preheat images.
	//
	// If multiple images are provided, the status of each image will be returned respectively.
	// One preheating failure will not cause the whole process fail.
	// If meet internal problems rather than failure results returned by the providers,
	// an non nil error will be returned.
	//
	PreheatImages(images ...models.ImageRepository) (CompositePreheatingResults, error)

	// Load the history records on top of the query parameters.
	//
	// params *models.QueryParam : parameters for querying
	//
	LoadHistoryRecords(params *models.QueryParam) (int64, []*models.HistoryRecord, error)
}

// CoreController is the default implementation of Controller interface.
//
type CoreController struct {
	// For history
	hManager history.Manager

	// For instance
	iManager instance.Manager

	// Monitor and update the progress of tasks and health of instances
	monitor *Monitor
}

// NewCoreController is constructor of controller
func NewCoreController(ctx context.Context) (*CoreController, error) {
	iManager := instance.NewDefaultManager()
	if iManager == nil {
		return nil, errors.New("nil instance manager")
	}
	hManager := history.NewDefaultManager()
	if hManager == nil {
		return nil, errors.New("nil history manager")
	}

	return &CoreController{
		iManager: iManager,
		hManager: hManager,
		monitor:  NewMonitor(ctx, iManager, hManager),
	}, nil
}

// GetAvailableProviders implements @Controller.GetAvailableProviders
func (cc *CoreController) GetAvailableProviders() ([]*provider.Metadata, error) {
	return provider.ListProviders()
}

// ListInstances implements @Controller.ListInstances
func (cc *CoreController) ListInstances(params *models.QueryParam) (int64, []*models.Metadata, error) {
	return cc.iManager.List(params)
}

// CreateInstance implements @Controller.CreateInstance
func (cc *CoreController) CreateInstance(instance *models.Metadata) (int64, error) {
	if instance == nil {
		return 0, errors.New("nil instance object provided")
	}

	// Avoid duplicated endpoint
	_, allOnes, err := cc.iManager.List(nil)
	if err != nil {
		return 0, err
	}
	for _, theOne := range allOnes {
		if theOne.Endpoint == instance.Endpoint {
			return 0, ErrorConflict
		}
	}
	// Check health before saving
	f, ok := provider.GetProvider(instance.Provider)
	if !ok {
		return 0, fmt.Errorf("no provider registered with name '%s'", instance.Provider)
	}
	p, err := f(instance)
	if err != nil {
		return 0, err
	}

	status, err := p.GetHealth()
	if err != nil {
		instance.Status = provider.DriverStatusUnHealthy
		log.Errorf("Check health of new instance error: %s; set healthy status to unhealthy", err)
	} else {
		instance.Status = status.Status
	}

	instance.SetupTimestamp = time.Now().Unix()

	return cc.iManager.Save(instance)
}

// DeleteInstance implements @Controller.DeleteInstance
func (cc *CoreController) DeleteInstance(id int64) error {
	return cc.iManager.Delete(id)
}

// UpdateInstance implements @Controller.UpdateInstance
func (cc *CoreController) UpdateInstance(id int64, properties models.PropertySet) error {
	if len(properties) == 0 {
		return errors.New("no properties provided to update")
	}

	metadata, err := cc.iManager.Get(id)
	if err != nil {
		return err
	}

	if err := properties.Apply(metadata); err != nil {
		return err
	}

	return cc.iManager.Update(metadata)
}

// PreheatImages implements @Controller.PreheatImages
func (cc *CoreController) PreheatImages(images ...models.ImageRepository) (CompositePreheatingResults, error) {
	if len(images) == 0 {
		return nil, errors.New("no images provided to preheat")
	}

	// Valid the images
	for _, img := range images {
		if !img.Valid() {
			return nil, fmt.Errorf("%s is not a valid image repository", img)
		}
	}

	// Directly dispatch to all the instances
	// TODO: Use async way in future
	_, instances, err := cc.iManager.List(nil)
	if err != nil {
		return nil, err
	}

	// No instances
	if len(instances) == 0 {
		return nil, errors.New("no distribution provider instances")
	}

	// TODO: refine the logic to remove those vars
	validCount := 0
	results := make(CompositePreheatingResults)
	for _, inst := range instances {
		// Instance must be enabled and healthy
		if inst.Enabled && inst.Status != provider.DriverStatusUnHealthy {
			validCount++
			var allStatus []*provider.PreheatingStatus
			results[inst.ID] = &allStatus

			factory, ok := provider.GetProvider(inst.Provider)
			if !ok {
				// Append error
				err := fmt.Errorf("the specified provider %s for instance %d is not registered", inst.Provider, inst.ID)
				log.Errorf("get provider factory error: %s", err)

				allStatus = append(allStatus, preheatingStatus("-", models.PreheatingStatusFail, err))
				continue
			}

			p, err := factory(inst)
			if err != nil {
				// Append error
				log.Errorf("initialize provider error: %s", err)

				allStatus = append(allStatus, preheatingStatus("-", models.PreheatingStatusFail, fmt.Errorf("initialize provider error: %s", err)))
				continue
			}

			// Dispatch
			for _, img := range images {
				preheatImg, err := buildImageData(img)
				if err != nil {
					log.Errorf("build image data error: %s", err)

					allStatus = append(allStatus, preheatingStatus(string(img), models.PreheatingStatusFail, err))
					continue
				}
				log.Debugf("Preheating image %v to instance %s", preheatImg, inst.Name)

				pStatus, err := p.Preheat(preheatImg)
				if err != nil {
					log.Errorf("preheat image error: %s", err)

					allStatus = append(allStatus, preheatingStatus(string(img), models.PreheatingStatusFail, err))
					continue
				}

				// Append a new history record
				if err := cc.hManager.AppendHistory(&models.HistoryRecord{
					TaskID:     pStatus.TaskID,
					Image:      string(img),
					StartTime:  "-",
					FinishTime: "-",
					Status:     pStatus.Status,
					Provider:   inst.Provider,
					Instance:   inst.ID,
				}); err != nil {
					// Just log it
					log.Errorf("save history record error: %s", err)
				} else {
					// Monitor it
					cc.monitor.WatchProgress(inst.ID, pStatus.TaskID)
				}

				allStatus = append(allStatus, pStatus)
			}
		}
	}

	if validCount == 0 {
		return nil, errors.New("no enabled healthy instances existing")
	}

	return results, nil
}

// LoadHistoryRecords implements @Controller.LoadHistoryRecords
func (cc *CoreController) LoadHistoryRecords(params *models.QueryParam) (int64, []*models.HistoryRecord, error) {
	return cc.hManager.LoadHistories(params)
}

// GetInstance implements @Controller.GetInstance
func (cc *CoreController) GetInstance(id int64) (*models.Metadata, error) {
	return cc.iManager.Get(id)
}

// Init the distribution providers
func Init(ctx context.Context) {
	if DefaultController == nil {
		if c, err := NewCoreController(ctx); err != nil {
			log.Fatalf("initialize distribution controller error: %s", err)
		} else {
			c.monitor.Start()
			DefaultController = c

			// Sync task status
			allItemNotDone, err := syncTaskStatus()
			if err != nil {
				log.Error(err)
			}

			for _, item := range allItemNotDone {
				c.monitor.WatchProgress(item.Instance, item.TaskID)
				log.Debugf("Sync status for task %s against %d", item.TaskID, item.Instance)
			}
		}
	}
}

// Sync the task status when starting
func syncTaskStatus() ([]*models.HistoryRecord, error) {
	// Load all the tasks from storage
	// TODO: there should be a better sync way
	_, all, err := DefaultController.LoadHistoryRecords(nil)
	if err != nil {
		return nil, fmt.Errorf("sync status of preheating tasks error: %s", err)
	}

	allItemsNotDone := make([]*models.HistoryRecord, 0)
	for _, taskRecord := range all {
		status := models.TrackStatus(taskRecord.Status)
		done := status.Success() || status.Fail()
		if !done {
			allItemsNotDone = append(allItemsNotDone, taskRecord)
		}
	}

	return allItemsNotDone, nil
}

// Create a preheating status
func preheatingStatus(taskID, status string, err error) *provider.PreheatingStatus {
	return &provider.PreheatingStatus{
		TaskID: taskID,
		Status: status,
		Error:  err.Error(),
	}
}

// convert the image to preheat image by adding more required data
func buildImageData(image models.ImageRepository) (*provider.PreheatImage, error) {
	extEndpoint, err := config.ExtEndpoint()
	if err != nil {
		return nil, err
	}

	// extURL, err := config.ExtURL()
	// if err != nil {
	// 	return nil, err
	// }

	access := []*tk.ResourceActions{
		{
			Type:    "repository",
			Name:    fmt.Sprintf("%s", image.Name()),
			Actions: []string{"pull", "push", "*"},
		},
	}

	tk, err := token.MakeToken("distributor", token.Registry, access)
	if err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf("%s/v2/%s/manifests/%s", extEndpoint, image.Name(), image.Tag())

	return &provider.PreheatImage{
		Type: models.PreheatingImageTypeImage,
		URL:  fullURL,
		Headers: map[string]interface{}{
			"Authorization": fmt.Sprintf("Bearer %s", tk.Token),
		},
	}, nil
}
