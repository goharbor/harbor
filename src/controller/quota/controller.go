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

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"

	// quota driver
	_ "github.com/goharbor/harbor/src/controller/quota/driver"
	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/gtask"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	libredis "github.com/goharbor/harbor/src/lib/redis"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/quota"
	"github.com/goharbor/harbor/src/pkg/quota/driver"
	"github.com/goharbor/harbor/src/pkg/quota/types"

	// init the db config
	_ "github.com/goharbor/harbor/src/pkg/config/db"
)

func init() {
	// register the async task for flushing quota to db when enable update quota by redis
	if provider := config.GetQuotaUpdateProvider(); provider == updateQuotaProviderRedis.String() {
		gtask.DefaultPool().AddTask(flushQuota, 30*time.Second)
	}
}

type updateQuotaProviderType string

func (t updateQuotaProviderType) String() string {
	return string(t)
}

var (
	defaultRetryTimeout = time.Minute * 5
	// quotaExpireTimeout is the expire time for quota when update quota by redis
	quotaExpireTimeout = time.Minute * 5

	updateQuotaProviderRedis updateQuotaProviderType = "redis"
	updateQuotaProviderDB    updateQuotaProviderType = "db"
)

var (
	// Ctl is a global quota controller instance
	Ctl = NewController()
)

// Controller defines the operations related with quotas
type Controller interface {
	// Count returns the total count of quotas according to the query.
	Count(ctx context.Context, query *q.Query) (int64, error)

	// Create ensure quota for the reference object
	Create(ctx context.Context, reference, referenceID string, hardLimits types.ResourceList, used ...types.ResourceList) (int64, error)

	// Delete delete quota by id
	Delete(ctx context.Context, id int64) error

	// Get returns quota by id
	Get(ctx context.Context, id int64, options ...Option) (*quota.Quota, error)

	// GetByRef returns quota by reference object
	GetByRef(ctx context.Context, reference, referenceID string, options ...Option) (*quota.Quota, error)

	// IsEnabled returns true when quota enabled for reference object
	IsEnabled(ctx context.Context, reference, referenceID string) (bool, error)

	// List list quotas
	List(ctx context.Context, query *q.Query, options ...Option) ([]*quota.Quota, error)

	// Refresh refresh quota for the reference object
	Refresh(ctx context.Context, reference, referenceID string, options ...Option) error

	// Request request resources to run f
	// Before run the function, it reserves the resources,
	// then runs f and refresh quota when f success，
	// in the finally it releases the resources which reserved at the beginning.
	Request(ctx context.Context, reference, referenceID string, resources types.ResourceList, f func() error) error

	// Update update quota
	Update(ctx context.Context, q *quota.Quota) error
}

// NewController creates an instance of the default quota controller
func NewController() Controller {
	return &controller{
		quotaMgr: quota.Mgr,
	}
}

type controller struct {
	reservedExpiration time.Duration

	quotaMgr quota.Manager
	g        singleflight.Group
}

// flushQuota flushes the quota info from redis to db asynchronously.
func flushQuota(ctx context.Context) {
	iter, err := cache.Default().Scan(ctx, "quota:*")
	if err != nil {
		log.Errorf("failed to scan out the quota records from redis")
		return
	}

	for iter.Next(ctx) {
		key := iter.Val()
		q := &quota.Quota{}
		err = cache.Default().Fetch(ctx, key, q)
		if err != nil {
			log.Errorf("failed to fetch quota: %s, error: %v", key, err)
			continue
		}

		if err = Ctl.Update(ctx, q); err != nil {
			log.Errorf("failed to refresh quota: %s, error: %v", key, err)
		} else {
			log.Debugf("successfully refreshed quota: %s", key)
		}
	}
}

func (c *controller) Count(ctx context.Context, query *q.Query) (int64, error) {
	return c.quotaMgr.Count(ctx, query)
}

func (c *controller) Create(ctx context.Context, reference, referenceID string, hardLimits types.ResourceList, used ...types.ResourceList) (int64, error) {
	return c.quotaMgr.Create(ctx, reference, referenceID, hardLimits, used...)
}

func (c *controller) Delete(ctx context.Context, id int64) error {
	return c.quotaMgr.Delete(ctx, id)
}

func (c *controller) Get(ctx context.Context, id int64, options ...Option) (*quota.Quota, error) {
	q, err := c.quotaMgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return c.assembleQuota(ctx, q, newOptions(options...))
}

func (c *controller) GetByRef(ctx context.Context, reference, referenceID string, options ...Option) (*quota.Quota, error) {
	q, err := c.quotaMgr.GetByRef(ctx, reference, referenceID)
	if err != nil {
		return nil, err
	}

	return c.assembleQuota(ctx, q, newOptions(options...))
}

func (c *controller) assembleQuota(ctx context.Context, q *quota.Quota, opts *Options) (*quota.Quota, error) {
	if opts.WithReferenceObject {
		driver, err := Driver(ctx, q.Reference)
		if err != nil {
			return nil, err
		}

		ref, err := driver.Load(ctx, q.ReferenceID)
		if err != nil {
			log.G(ctx).Warningf("failed to load referenced %s object %s for quota %d, error %v",
				q.Reference, q.ReferenceID, q.ID, err)
		} else {
			q.Ref = ref
		}
	}

	return q, nil
}

func (c *controller) IsEnabled(ctx context.Context, reference, referenceID string) (bool, error) {
	d, err := Driver(ctx, reference)
	if err != nil {
		return false, err
	}

	return d.Enabled(ctx, referenceID)
}

func (c *controller) List(ctx context.Context, query *q.Query, options ...Option) ([]*quota.Quota, error) {
	quotas, err := c.quotaMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}

	opts := newOptions(options...)
	for _, q := range quotas {
		if _, err := c.assembleQuota(ctx, q, opts); err != nil {
			return nil, err
		}
	}

	return quotas, nil
}

// updateUsageByDB updates the quota usage by the database which updates the quota usage immediately.
func (c *controller) updateUsageByDB(ctx context.Context, reference, referenceID string, op func(hardLimits, used types.ResourceList) (types.ResourceList, error)) error {
	q, err := c.quotaMgr.GetByRef(ctx, reference, referenceID)
	if err != nil {
		return retry.Abort(err)
	}

	hardLimits, err := q.GetHard()
	if err != nil {
		return retry.Abort(err)
	}

	used, err := q.GetUsed()
	if err != nil {
		return retry.Abort(err)
	}

	newUsed, err := op(hardLimits, used)
	if err != nil {
		return retry.Abort(err)
	}

	// The PR https://github.com/goharbor/harbor/pull/17392 optimized the logic for post upload blob which use size 0
	// for checking quota, this will increase the pressure of optimistic lock, so here return earlier
	// if the quota usage has not changed to reduce the probability of optimistic lock.
	if types.Equals(used, newUsed) {
		return nil
	}

	q.SetUsed(newUsed)

	err = c.quotaMgr.Update(ctx, q)
	if err != nil && !errors.Is(err, orm.ErrOptimisticLock) {
		return retry.Abort(err)
	}

	return err
}

// updateUsageByRedis updates the quota usage by the redis and flush the quota usage to db asynchronously.
func (c *controller) updateUsageByRedis(ctx context.Context, reference, referenceID string, op func(hardLimits, used types.ResourceList) (types.ResourceList, error)) error {
	// earlier abort if context is error such as context canceled
	if ctx.Err() != nil {
		return retry.Abort(ctx.Err())
	}

	client, err := libredis.GetHarborClient()
	if err != nil {
		return retry.Abort(err)
	}
	// normally use cache.Save will append prefix "cache:", in order to keep consistent
	// here adopts raw redis client should also pad the prefix manually.
	key := fmt.Sprintf("%s:quota:%s:%s", "cache", reference, referenceID)
	return client.Watch(ctx, func(tx *redis.Tx) error {
		data, err := tx.Get(ctx, key).Result()
		if err != nil && err != redis.Nil {
			return retry.Abort(err)
		}

		q := &quota.Quota{}
		// calc the quota usage in real time if no key found
		if err == redis.Nil {
			// use singleflight to prevent cache penetration and cause pressure on the database.
			realQuota, err, _ := c.g.Do(key, func() (any, error) {
				return c.calcQuota(ctx, reference, referenceID)
			})
			if err != nil {
				return retry.Abort(err)
			}

			q = realQuota.(*quota.Quota)
		} else {
			if err = cache.DefaultCodec().Decode([]byte(data), q); err != nil {
				return retry.Abort(err)
			}
		}

		hardLimits, err := q.GetHard()
		if err != nil {
			return retry.Abort(err)
		}

		used, err := q.GetUsed()
		if err != nil {
			return retry.Abort(err)
		}

		newUsed, err := op(hardLimits, used)
		if err != nil {
			return retry.Abort(err)
		}

		q.SetUsed(newUsed)

		val, err := cache.DefaultCodec().Encode(q)
		if err != nil {
			return retry.Abort(err)
		}

		_, err = tx.TxPipelined(ctx, func(p redis.Pipeliner) error {
			_, err = p.Set(ctx, key, val, quotaExpireTimeout).Result()
			return err
		})

		if err != nil && err != redis.TxFailedErr {
			return retry.Abort(err)
		}

		return err
	}, key)
}

func (c *controller) updateUsageWithRetry(ctx context.Context, reference, referenceID string, op func(hardLimits, used types.ResourceList) (types.ResourceList, error), provider updateQuotaProviderType, retryOpts ...retry.Option) error {
	var f func() error
	switch provider {
	case updateQuotaProviderDB:
		f = func() error {
			return c.updateUsageByDB(ctx, reference, referenceID, op)
		}
	case updateQuotaProviderRedis:
		f = func() error {
			return c.updateUsageByRedis(ctx, reference, referenceID, op)
		}
	default:
		// by default is update quota by db
		f = func() error {
			return c.updateUsageByDB(ctx, reference, referenceID, op)
		}
	}

	options := []retry.Option{
		retry.Timeout(defaultRetryTimeout),
		retry.Backoff(false),
		retry.Callback(func(err error, _ time.Duration) {
			log.G(ctx).Debugf("failed to update the quota usage for %s %s, error: %v", reference, referenceID, err)
		}),
	}
	// append for override default retry options
	if len(retryOpts) > 0 {
		options = append(options, retryOpts...)
	}

	return retry.Retry(f, options...)
}

func (c *controller) Refresh(ctx context.Context, reference, referenceID string, options ...Option) error {
	driver, err := Driver(ctx, reference)
	if err != nil {
		return err
	}

	opts := newOptions(options...)

	calculateUsage := func() (types.ResourceList, error) {
		newUsed, err := driver.CalculateUsage(ctx, referenceID)
		if err != nil {
			log.G(ctx).Errorf("failed to calculate quota usage for %s %s, error: %v", reference, referenceID, err)
			return nil, err
		}

		return newUsed, err
	}

	// update quota usage by db for refresh operation
	return c.updateUsageWithRetry(ctx, reference, referenceID, refreshResources(calculateUsage, opts.IgnoreLimitation), updateQuotaProviderType(config.GetQuotaUpdateProvider()), opts.RetryOptions...)
}

func (c *controller) Request(ctx context.Context, reference, referenceID string, resources types.ResourceList, f func() error) error {
	if len(resources) == 0 {
		return f()
	}

	provider := updateQuotaProviderType(config.GetQuotaUpdateProvider())
	if err := c.updateUsageWithRetry(ctx, reference, referenceID, reserveResources(resources), provider); err != nil {
		log.G(ctx).Errorf("reserve resources %s for %s %s failed, error: %v", resources.String(), reference, referenceID, err)
		return err
	}

	err := f()

	if err != nil {
		if er := c.updateUsageWithRetry(ctx, reference, referenceID, rollbackResources(resources), provider); er != nil {
			// ignore this error, the quota usage will be correct when users do operations which will call refresh quota
			log.G(ctx).Warningf("rollback resources %s for %s %s failed, error: %v", resources.String(), reference, referenceID, er)
		}
	}

	return err
}

// calcQuota calculates the quota and usage in real time.
func (c *controller) calcQuota(ctx context.Context, reference, referenceID string) (*quota.Quota, error) {
	// get quota and usage from db
	q, err := c.quotaMgr.GetByRef(ctx, reference, referenceID)
	if err != nil {
		return nil, err
	}
	// the usage in the db maybe outdated, calc it in real time
	driver, err := Driver(ctx, reference)
	if err != nil {
		return nil, err
	}

	newUsed, err := driver.CalculateUsage(ctx, referenceID)
	if err != nil {
		log.G(ctx).Errorf("failed to calculate quota usage for %s %s, error: %v", reference, referenceID, err)
		return nil, err
	}

	q.SetUsed(newUsed)
	return q, nil
}

func (c *controller) Update(ctx context.Context, u *quota.Quota) error {
	f := func() error {
		q, err := c.quotaMgr.GetByRef(ctx, u.Reference, u.ReferenceID)
		if err != nil {
			return err
		}

		if oldHard, err := q.GetHard(); err == nil {
			if newHard, err := u.GetHard(); err == nil {
				if !types.Equals(oldHard, newHard) {
					q.SetHard(newHard)
				}
			}
		}

		if oldUsed, err := q.GetUsed(); err == nil {
			if newUsed, err := u.GetUsed(); err == nil {
				if !types.Equals(oldUsed, newUsed) {
					q.SetUsed(newUsed)
				}
			}
		}

		return c.quotaMgr.Update(ctx, q)
	}

	options := []retry.Option{
		retry.Timeout(defaultRetryTimeout),
		retry.Backoff(false),
	}

	return retry.Retry(f, options...)
}

// Driver returns quota driver for the reference
func Driver(_ context.Context, reference string) (driver.Driver, error) {
	d, ok := driver.Get(reference)
	if !ok {
		return nil, fmt.Errorf("quota not support for %s", reference)
	}

	return d, nil
}

// Validate validate hard limits
func Validate(ctx context.Context, reference string, hardLimits types.ResourceList) error {
	d, err := Driver(ctx, reference)
	if err != nil {
		return err
	}

	return d.Validate(hardLimits)
}

func reserveResources(resources types.ResourceList) func(hardLimits, used types.ResourceList) (types.ResourceList, error) {
	return func(hardLimits, used types.ResourceList) (types.ResourceList, error) {
		newUsed := types.Add(used, resources)

		if err := quota.IsSafe(hardLimits, used, newUsed, false); err != nil {
			return nil, errors.DeniedError(err).WithMessagef("Quota exceeded when processing the request of %v", err)
		}

		return newUsed, nil
	}
}

func rollbackResources(resources types.ResourceList) func(hardLimits, used types.ResourceList) (types.ResourceList, error) {
	return func(_, used types.ResourceList) (types.ResourceList, error) {
		newUsed := types.Subtract(used, resources)
		// ensure that new used is never negative
		if negativeUsed := types.IsNegative(newUsed); len(negativeUsed) > 0 {
			return nil, fmt.Errorf("resources is negative for resource(s): %s", quota.PrettyPrintResourceNames(negativeUsed))
		}

		return newUsed, nil
	}
}

func refreshResources(calculateUsage func() (types.ResourceList, error), ignoreLimitation bool) func(hardLimits, used types.ResourceList) (types.ResourceList, error) {
	return func(hardLimits, used types.ResourceList) (types.ResourceList, error) {
		newUsed, err := calculateUsage()
		if err != nil {
			return nil, err
		}

		// ensure that new used is never negative
		if negativeUsed := types.IsNegative(newUsed); len(negativeUsed) > 0 {
			return nil, fmt.Errorf("quota usage is negative for resource(s): %s", quota.PrettyPrintResourceNames(negativeUsed))
		}

		if err := quota.IsSafe(hardLimits, used, newUsed, ignoreLimitation); err != nil {
			return nil, err
		}

		return newUsed, nil
	}
}
