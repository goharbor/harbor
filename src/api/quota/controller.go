// Copyright Project Harbor Authors
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

package quota

import (
	"context"
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/common/utils/log"
	util "github.com/goharbor/harbor/src/common/utils/redis"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/internal/orm"
	"github.com/goharbor/harbor/src/pkg/quota"
	"github.com/goharbor/harbor/src/pkg/quota/driver"
	"github.com/goharbor/harbor/src/pkg/types"

	// quota driver
	_ "github.com/goharbor/harbor/src/api/quota/driver"
)

var (
	// expire reserved resources when no actions on the key of the reserved resources in redis during 1 hour
	defaultReservedExpiration = time.Hour
)

var (
	// Ctl is a global quota controller instance
	Ctl = NewController()
)

// Controller defines the operations related with quotas
type Controller interface {
	// Create ensure quota for the reference object
	Create(ctx context.Context, reference, referenceID string, hardLimits types.ResourceList, used ...types.ResourceList) (int64, error)

	// Delete delete quota by id
	Delete(ctx context.Context, id int64) error

	// Get returns quota by id
	Get(ctx context.Context, id int64) (*quota.Quota, error)

	// IsEnabled returns true when quota enabled for reference object
	IsEnabled(ctx context.Context, reference, referenceID string) (bool, error)

	// Refresh refresh quota for the reference object
	Refresh(ctx context.Context, reference, referenceID string) error

	// Request request resources to run f
	// Before run the function, it reserves the resources,
	// then runs f and refresh quota when f success，
	// in the finally it releases the resources which reserved at the beginning.
	Request(ctx context.Context, reference, referenceID string, resources types.ResourceList, f func() error) error
}

// NewController creates an instance of the default quota controller
func NewController() Controller {
	return &controller{
		logPrefix:          "[controller][quota]",
		reservedExpiration: defaultReservedExpiration,
		quotaMgr:           quota.Mgr,
	}
}

type controller struct {
	logPrefix          string
	reservedExpiration time.Duration

	quotaMgr quota.Manager
}

func (c *controller) Create(ctx context.Context, reference, referenceID string, hardLimits types.ResourceList, used ...types.ResourceList) (int64, error) {
	return c.quotaMgr.Create(ctx, reference, referenceID, hardLimits, used...)
}

func (c *controller) Delete(ctx context.Context, id int64) error {
	return c.quotaMgr.Delete(ctx, id)
}

func (c *controller) Get(ctx context.Context, id int64) (*quota.Quota, error) {
	return c.quotaMgr.Get(ctx, id)
}

func (c *controller) IsEnabled(ctx context.Context, reference, referenceID string) (bool, error) {
	d, err := quotaDriver(ctx, reference, referenceID)
	if err != nil {
		return false, err
	}

	return d.Enabled(ctx, referenceID)
}

func (c *controller) getReservedResources(ctx context.Context, reference, referenceID string) (types.ResourceList, error) {
	conn := util.DefaultPool().Get()
	defer conn.Close()

	key := reservedResourcesKey(reference, referenceID)

	str, err := redis.String(conn.Do("GET", key))
	if err == redis.ErrNil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return types.NewResourceList(str)
}

func (c *controller) setReservedResources(ctx context.Context, reference, referenceID string, resources types.ResourceList) error {
	conn := util.DefaultPool().Get()
	defer conn.Close()

	key := reservedResourcesKey(reference, referenceID)

	reply, err := redis.String(conn.Do("SET", key, resources.String(), "EX", int64(c.reservedExpiration/time.Second)))
	if err != nil {
		return err
	}

	if reply != "OK" {
		return fmt.Errorf("bad reply value")
	}

	return nil
}

func (c *controller) reserveResources(ctx context.Context, reference, referenceID string, resources types.ResourceList) error {
	reserve := func(ctx context.Context) error {
		q, err := c.quotaMgr.GetForUpdate(ctx, reference, referenceID)
		if err != nil {
			return err
		}

		hardLimits, err := q.GetHard()
		if err != nil {
			return err
		}

		used, err := q.GetUsed()
		if err != nil {
			return err
		}

		reserved, err := c.getReservedResources(ctx, reference, referenceID)
		if err != nil {
			log.Errorf("failed to get reserved resources for %s %s, error: %v", reference, referenceID, err)
			return err
		}

		newReserved := types.Add(reserved, resources)

		newUsed := types.Add(used, newReserved)
		if err := quota.IsSafe(hardLimits, used, newUsed); err != nil {
			return ierror.DeniedError(nil).WithMessage("Quota exceeded when processing the request of %v", err)
		}

		if err := c.setReservedResources(ctx, reference, referenceID, newReserved); err != nil {
			log.Errorf("failed to set reserved resources for %s %s, error: %v", reference, referenceID, err)
			return err
		}

		return nil
	}

	return orm.WithTransaction(reserve)(ctx)
}

func (c *controller) unreserveResources(ctx context.Context, reference, referenceID string, resources types.ResourceList) error {
	unreserve := func(ctx context.Context) error {
		if _, err := c.quotaMgr.GetForUpdate(ctx, reference, referenceID); err != nil {
			return err
		}

		reserved, err := c.getReservedResources(ctx, reference, referenceID)
		if err != nil {
			log.Errorf("failed to get reserved resources for %s %s, error: %v", reference, referenceID, err)
			return err
		}

		newReserved := types.Subtract(reserved, resources)
		// ensure that new used is never negative
		if negativeUsed := types.IsNegative(newReserved); len(negativeUsed) > 0 {
			return fmt.Errorf("reserved resources is negative for resource(s): %s", quota.PrettyPrintResourceNames(negativeUsed))
		}

		if err := c.setReservedResources(ctx, reference, referenceID, newReserved); err != nil {
			log.Errorf("failed to set reserved resources for %s %s, error: %v", reference, referenceID, err)
			return err
		}

		return nil
	}

	return orm.WithTransaction(unreserve)(ctx)
}

func (c *controller) Refresh(ctx context.Context, reference, referenceID string) error {
	driver, err := quotaDriver(ctx, reference, referenceID)
	if err != nil {
		return err
	}

	refresh := func(ctx context.Context) error {
		q, err := c.quotaMgr.GetForUpdate(ctx, reference, referenceID)
		if err != nil {
			return err
		}

		hardLimits, err := q.GetHard()
		if err != nil {
			return err
		}

		used, err := q.GetUsed()
		if err != nil {
			return err
		}

		newUsed, err := driver.CalculateUsage(ctx, referenceID)
		if err != nil {
			log.Errorf("failed to calculate quota usage for %s %s, error: %v", reference, referenceID, err)
			return err
		}

		// ensure that new used is never negative
		if negativeUsed := types.IsNegative(newUsed); len(negativeUsed) > 0 {
			return fmt.Errorf("quota usage is negative for resource(s): %s", quota.PrettyPrintResourceNames(negativeUsed))
		}

		if err := quota.IsSafe(hardLimits, used, newUsed); err != nil {
			return err
		}

		q.SetUsed(newUsed)
		q.UpdateTime = time.Now()

		return c.quotaMgr.Update(ctx, q)
	}

	return orm.WithTransaction(refresh)(ctx)
}

func (c *controller) Request(ctx context.Context, reference, referenceID string, resources types.ResourceList, f func() error) error {
	if len(resources) == 0 {
		return f()
	}

	if err := c.reserveResources(ctx, reference, referenceID, resources); err != nil {
		return err
	}

	defer func() {
		if err := c.unreserveResources(ctx, reference, referenceID, resources); err != nil {
			// ignore this error because reserved resources will be expired
			// when no actions on the key of the reserved resources in redis during sometimes
			log.Warningf("unreserve resources %s for %s %s failed, error: %v", resources.String(), reference, referenceID, err)
		}
	}()

	if err := f(); err != nil {
		return err
	}

	return c.Refresh(ctx, reference, referenceID)
}

func quotaDriver(ctx context.Context, reference, referenceID string) (driver.Driver, error) {
	d, ok := driver.Get(reference)
	if !ok {
		return nil, fmt.Errorf("quota not support for %s", reference)
	}

	return d, nil
}

func reservedResourcesKey(reference, referenceID string) string {
	return fmt.Sprintf("quota:%s:%s:reserved", reference, referenceID)
}
