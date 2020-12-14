package preheat

import (
	"context"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/instance"
	policyModels "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/policy"
	providerModels "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/policy"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
)

const (
	// SchedulerCallback ...
	SchedulerCallback = "P2PPreheatCallback"
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

	// GetInstance the metadata of the specified instance
	GetInstanceByName(ctx context.Context, name string) (*providerModels.Instance, error)

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

	// do not provide another policy controller, mixed in preheat controller

	// CountPolicy returns the total count of the policy.
	CountPolicy(ctx context.Context, query *q.Query) (int64, error)
	// CreatePolicy creates the policy.
	CreatePolicy(ctx context.Context, schema *policyModels.Schema) (int64, error)
	// GetPolicy gets the policy by id.
	GetPolicy(ctx context.Context, id int64) (*policyModels.Schema, error)
	// GetPolicyByName gets the policy by name.
	GetPolicyByName(ctx context.Context, projectID int64, name string) (*policyModels.Schema, error)
	// UpdatePolicy updates the policy.
	UpdatePolicy(ctx context.Context, schema *policyModels.Schema, props ...string) error
	// DeletePolicy deletes the policy by id.
	DeletePolicy(ctx context.Context, id int64) error
	// ListPolicies lists policies by query.
	ListPolicies(ctx context.Context, query *q.Query) ([]*policyModels.Schema, error)
	// ListPoliciesByProject lists policies by project.
	ListPoliciesByProject(ctx context.Context, project int64, query *q.Query) ([]*policyModels.Schema, error)
	// CheckHealth checks the instance health, for test connection
	CheckHealth(ctx context.Context, instance *providerModels.Instance) error
	// DeletePoliciesOfProject delete all policies under one project
	DeletePoliciesOfProject(ctx context.Context, project int64) error
}

var _ Controller = (*controller)(nil)

// controller is the default implementation of Controller interface.
//
type controller struct {
	// For instance
	iManager instance.Manager
	// For policy
	pManager     policy.Manager
	scheduler    scheduler.Scheduler
	executionMgr task.ExecutionManager
}

// NewController is constructor of controller
func NewController() Controller {
	return &controller{
		iManager:     instance.Mgr,
		pManager:     policy.Mgr,
		scheduler:    scheduler.Sched,
		executionMgr: task.NewExecutionManager(),
	}
}

// GetAvailableProviders implements @Controller.GetAvailableProviders
func (c *controller) GetAvailableProviders() ([]*provider.Metadata, error) {
	return provider.ListProviders()
}

// CountInstance implements @Controller.CountInstance
func (c *controller) CountInstance(ctx context.Context, query *q.Query) (int64, error) {
	return c.iManager.Count(ctx, query)
}

// ListInstance implements @Controller.ListInstance
func (c *controller) ListInstance(ctx context.Context, query *q.Query) ([]*providerModels.Instance, error) {
	return c.iManager.List(ctx, query)
}

// CreateInstance implements @Controller.CreateInstance
func (c *controller) CreateInstance(ctx context.Context, instance *providerModels.Instance) (int64, error) {
	if instance == nil {
		return 0, errors.New("nil instance object provided")
	}

	// Avoid duplicated endpoint
	var query = &q.Query{
		Keywords: map[string]interface{}{
			"endpoint": instance.Endpoint,
		},
	}
	num, err := c.iManager.Count(ctx, query)
	if err != nil {
		return 0, err
	}
	if num > 0 {
		return 0, ErrorConflict
	}

	// !WARN: We don't check the health of the instance here.
	// That is ok because the health of instance will be checked before enforcing the policy each time.

	instance.SetupTimestamp = time.Now().Unix()

	return c.iManager.Save(ctx, instance)
}

// DeleteInstance implements @Controller.Delete
func (c *controller) DeleteInstance(ctx context.Context, id int64) error {
	ins, err := c.GetInstance(ctx, id)
	if err != nil {
		return err
	}
	// delete instance should check the instance whether be used by policies
	policies, err := c.ListPolicies(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"provider_id": id,
		},
	})
	if err != nil {
		return err
	}

	if len(policies) > 0 {
		return errors.New(nil).
			WithCode(errors.PreconditionCode).
			WithMessage("Provider [%s] cannot be deleted as some preheat policies are using it", ins.Name)
	}

	return c.iManager.Delete(ctx, id)
}

// UpdateInstance implements @Controller.Update
func (c *controller) UpdateInstance(ctx context.Context, instance *providerModels.Instance, properties ...string) error {
	oldIns, err := c.GetInstance(ctx, instance.ID)
	if err != nil {
		return err
	}

	if !instance.Enabled {
		// update instance should check the instance whether be used by policies
		policies, err := c.ListPolicies(ctx, &q.Query{
			Keywords: map[string]interface{}{
				"provider_id": instance.ID,
			},
		})
		if err != nil {
			return err
		}

		if len(policies) > 0 {
			return errors.New(nil).
				WithCode(errors.PreconditionCode).
				WithMessage("Provider [%s] cannot be disabled as some preheat policies are using it", oldIns.Name)
		}
	}

	// vendor type does not support change
	if oldIns.Vendor != instance.Vendor {
		return errors.Errorf("provider [%s] vendor cannot be changed", oldIns.Name)
	}

	return c.iManager.Update(ctx, instance, properties...)
}

// GetInstance implements @Controller.Get
func (c *controller) GetInstance(ctx context.Context, id int64) (*providerModels.Instance, error) {
	return c.iManager.Get(ctx, id)
}

func (c *controller) GetInstanceByName(ctx context.Context, name string) (*providerModels.Instance, error) {
	return c.iManager.GetByName(ctx, name)
}

// CountPolicy returns the total count of the policy.
func (c *controller) CountPolicy(ctx context.Context, query *q.Query) (int64, error) {
	return c.pManager.Count(ctx, query)
}

// TriggerParam ...
type TriggerParam struct {
	PolicyID int64
}

// CreatePolicy creates the policy.
func (c *controller) CreatePolicy(ctx context.Context, schema *policyModels.Schema) (id int64, err error) {
	if schema == nil {
		return 0, errors.New("nil policy schema provided")
	}

	// Update timestamps
	now := time.Now()
	schema.CreatedAt = now
	schema.UpdatedTime = now

	// Get full model of policy schema
	err = schema.Decode()
	if err != nil {
		return 0, err
	}

	id, err = c.pManager.Create(ctx, schema)
	if err != nil {
		return
	}

	schema.ID = id

	if schema.Trigger != nil &&
		schema.Trigger.Type == policyModels.TriggerTypeScheduled &&
		len(schema.Trigger.Settings.Cron) > 0 {
		// schedule and update policy
		extras := make(map[string]interface{})
		if _, err = c.scheduler.Schedule(ctx, job.P2PPreheat, id, "", schema.Trigger.Settings.Cron,
			SchedulerCallback, TriggerParam{PolicyID: id}, extras); err != nil {
			return 0, err
		}

		if err = schema.Encode(); err == nil {
			err = c.pManager.Update(ctx, schema, "trigger")
		}

		if err != nil {
			if e := c.scheduler.UnScheduleByVendor(ctx, job.P2PPreheat, id); e != nil {
				return 0, errors.Wrap(e, err.Error())
			}

			return 0, err
		}
	}

	return
}

// GetPolicy gets the policy by id.
func (c *controller) GetPolicy(ctx context.Context, id int64) (*policyModels.Schema, error) {
	return c.pManager.Get(ctx, id)
}

// GetPolicyByName gets the policy by name.
func (c *controller) GetPolicyByName(ctx context.Context, projectID int64, name string) (*policyModels.Schema, error) {
	return c.pManager.GetByName(ctx, projectID, name)
}

// UpdatePolicy updates the policy.
func (c *controller) UpdatePolicy(ctx context.Context, schema *policyModels.Schema, props ...string) error {
	if schema == nil {
		return errors.New("nil policy schema provided")
	}

	// Get policy cache
	s0, err := c.pManager.Get(ctx, schema.ID)
	if err != nil {
		return err
	}

	// Double check trigger
	if s0.Trigger == nil {
		return errors.Errorf("missing trigger settings in preheat policy %s", s0.Name)
	}

	// Get full model of updating policy
	err = schema.Decode()
	if err != nil {
		return err
	}

	var cron = schema.Trigger.Settings.Cron
	var oldCron = s0.Trigger.Settings.Cron
	var needUn bool
	var needSch bool

	if s0.Trigger.Type != schema.Trigger.Type {
		if s0.Trigger.Type == policyModels.TriggerTypeScheduled && len(oldCron) > 0 {
			needUn = true
		}
		if schema.Trigger.Type == policyModels.TriggerTypeScheduled && len(cron) > 0 {
			needSch = true
		}
	} else {
		// not change trigger type
		if schema.Trigger.Type == policyModels.TriggerTypeScheduled && oldCron != cron {
			// unschedule old
			if len(oldCron) > 0 {
				needUn = true
			}
			// schedule new
			if len(cron) > 0 {
				// valid cron
				needSch = true
			}
		}

	}

	// unschedule old
	if needUn {
		err = c.scheduler.UnScheduleByVendor(ctx, job.P2PPreheat, schema.ID)
		if err != nil {
			return err
		}
	}

	// schedule new
	if needSch {
		extras := make(map[string]interface{})
		if _, err := c.scheduler.Schedule(ctx, job.P2PPreheat, schema.ID, "", cron, SchedulerCallback,
			TriggerParam{PolicyID: schema.ID}, extras); err != nil {
			return err
		}
	}

	// Update timestamp
	schema.UpdatedTime = time.Now()

	err = c.pManager.Update(ctx, schema, props...)
	if (err != nil) && (needSch || needUn) {
		return errors.Wrapf(err, "Update failed, but not rollback scheduler")
	}

	return err
}

// DeletePolicy deletes the policy by id.
func (c *controller) DeletePolicy(ctx context.Context, id int64) error {
	s, err := c.pManager.Get(ctx, id)
	if err != nil {
		return err
	}
	if s.Trigger != nil && s.Trigger.Type == policyModels.TriggerTypeScheduled && len(s.Trigger.Settings.Cron) > 0 {
		err = c.scheduler.UnScheduleByVendor(ctx, job.P2PPreheat, id)
		if err != nil {
			return err
		}
	}

	if err = c.deleteExecs(ctx, id); err != nil {
		return err
	}

	return c.pManager.Delete(ctx, id)
}

// DeletePoliciesOfProject deletes all the policy under project.
func (c *controller) DeletePoliciesOfProject(ctx context.Context, project int64) error {
	policies, err := c.ListPoliciesByProject(ctx, project, nil)
	if err != nil {
		return err
	}

	for _, p := range policies {
		if err = c.DeletePolicy(ctx, p.ID); err != nil {
			return err
		}
	}
	return nil
}

// deleteExecs delete executions
func (c *controller) deleteExecs(ctx context.Context, vendorID int64) error {
	executions, err := c.executionMgr.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"VendorType": job.P2PPreheat,
			"VendorID":   vendorID,
		},
	})

	if err != nil {
		return err
	}

	for _, execution := range executions {
		if err = c.executionMgr.Delete(ctx, execution.ID); err != nil {
			return err
		}
	}

	return nil
}

// ListPolicies lists policies by query.
func (c *controller) ListPolicies(ctx context.Context, query *q.Query) ([]*policyModels.Schema, error) {
	return c.pManager.ListPolicies(ctx, query)
}

// ListPoliciesByProject lists policies by project.
func (c *controller) ListPoliciesByProject(ctx context.Context, project int64, query *q.Query) ([]*policyModels.Schema, error) {
	return c.pManager.ListPoliciesByProject(ctx, project, query)
}

// CheckHealth checks the instance health, for test connection
func (c *controller) CheckHealth(ctx context.Context, instance *providerModels.Instance) error {
	if instance == nil {
		return errors.New("instance can not be nil")
	}

	fac, ok := provider.GetProvider(instance.Vendor)
	if !ok {
		return errors.Errorf("no driver registered for provider %s", instance.Vendor)
	}

	// Construct driver
	driver, err := fac(instance)
	if err != nil {
		return err
	}

	// Check health
	h, err := driver.GetHealth()
	if err != nil {
		return err
	}

	if h.Status != provider.DriverStatusHealthy {
		return errors.Errorf("preheat provider instance %s-%s:%s is not healthy", instance.Vendor, instance.Name, instance.Endpoint)
	}

	return nil
}
