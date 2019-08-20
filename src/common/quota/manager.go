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
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/quota/driver"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/types"
)

// Manager manager for quota
type Manager struct {
	driver      driver.Driver
	reference   string
	referenceID string
}

func (m *Manager) addQuota(o orm.Ormer, hardLimits types.ResourceList, now time.Time) (int64, error) {
	quota := &models.Quota{
		Reference:    m.reference,
		ReferenceID:  m.referenceID,
		Hard:         hardLimits.String(),
		CreationTime: now,
		UpdateTime:   now,
	}

	return o.Insert(quota)
}

func (m *Manager) addUsage(o orm.Ormer, used types.ResourceList, now time.Time, ids ...int64) (int64, error) {
	usage := &models.QuotaUsage{
		Reference:    m.reference,
		ReferenceID:  m.referenceID,
		Used:         used.String(),
		CreationTime: now,
		UpdateTime:   now,
	}

	if len(ids) > 0 {
		usage.ID = ids[0]
	}

	return o.Insert(usage)
}

func (m *Manager) newQuota(o orm.Ormer, hardLimits types.ResourceList, usages ...types.ResourceList) (int64, error) {
	now := time.Now()

	id, err := m.addQuota(o, hardLimits, now)
	if err != nil {
		return 0, err
	}

	var used types.ResourceList
	if len(usages) > 0 {
		used = usages[0]
	} else {
		used = types.Zero(hardLimits)
	}

	if _, err := m.addUsage(o, used, now, id); err != nil {
		return 0, err
	}

	return id, nil
}

func (m *Manager) getQuotaForUpdate(o orm.Ormer) (*models.Quota, error) {
	quota := &models.Quota{Reference: m.reference, ReferenceID: m.referenceID}
	if err := o.ReadForUpdate(quota, "reference", "reference_id"); err != nil {
		if err == orm.ErrNoRows {
			if _, err := m.newQuota(o, m.driver.HardLimits()); err != nil {
				return nil, err
			}

			return m.getQuotaForUpdate(o)
		}

		return nil, err
	}

	return quota, nil
}

func (m *Manager) getUsageForUpdate(o orm.Ormer) (*models.QuotaUsage, error) {
	usage := &models.QuotaUsage{Reference: m.reference, ReferenceID: m.referenceID}
	if err := o.ReadForUpdate(usage, "reference", "reference_id"); err != nil {
		return nil, err
	}

	return usage, nil
}

func (m *Manager) updateUsage(o orm.Ormer, resources types.ResourceList,
	calculate func(types.ResourceList, types.ResourceList) types.ResourceList) error {

	quota, err := m.getQuotaForUpdate(o)
	if err != nil {
		return err
	}
	hardLimits, err := types.NewResourceList(quota.Hard)
	if err != nil {
		return err
	}

	usage, err := m.getUsageForUpdate(o)
	if err != nil {
		return err
	}
	used, err := types.NewResourceList(usage.Used)
	if err != nil {
		return err
	}

	newUsed := calculate(used, resources)

	// ensure that new used is never negative
	if negativeUsed := types.IsNegative(newUsed); len(negativeUsed) > 0 {
		return fmt.Errorf("quota usage is negative for resource(s): %s", prettyPrintResourceNames(negativeUsed))
	}

	if err := isSafe(hardLimits, used, newUsed); err != nil {
		return err
	}

	usage.Used = newUsed.String()
	usage.UpdateTime = time.Now()

	_, err = o.Update(usage)
	return err
}

// NewQuota create new quota for (reference, reference id)
func (m *Manager) NewQuota(hardLimit types.ResourceList, usages ...types.ResourceList) (int64, error) {
	var id int64
	err := dao.WithTransaction(func(o orm.Ormer) (err error) {
		id, err = m.newQuota(o, hardLimit, usages...)
		return err
	})

	if err != nil {
		return 0, err
	}

	return id, nil
}

// DeleteQuota delete the quota
func (m *Manager) DeleteQuota() error {
	return dao.WithTransaction(func(o orm.Ormer) error {
		quota := &models.Quota{Reference: m.reference, ReferenceID: m.referenceID}
		if _, err := o.Delete(quota, "reference", "reference_id"); err != nil {
			return err
		}

		usage := &models.QuotaUsage{Reference: m.reference, ReferenceID: m.referenceID}
		if _, err := o.Delete(usage, "reference", "reference_id"); err != nil {
			return err
		}

		return nil
	})
}

// UpdateQuota update the quota resource spec
func (m *Manager) UpdateQuota(hardLimits types.ResourceList) error {
	o := dao.GetOrmer()
	if err := m.driver.Validate(hardLimits); err != nil {
		return err
	}

	sql := `UPDATE quota SET hard = ? WHERE reference = ? AND reference_id = ?`
	_, err := o.Raw(sql, hardLimits.String(), m.reference, m.referenceID).Exec()

	return err
}

// SetResourceUsage sets the usage per resource name
func (m *Manager) SetResourceUsage(resource types.ResourceName, value int64) error {
	o := dao.GetOrmer()

	sql := fmt.Sprintf("UPDATE quota_usage SET used = jsonb_set(used, '{%s}', to_jsonb(%d::int), true) WHERE reference = ? AND reference_id = ?", resource, value)
	_, err := o.Raw(sql, m.reference, m.referenceID).Exec()

	return err
}

// EnsureQuota ensures the reference has quota and usage,
// if non-existent, will create new quota and usage.
// if existent, update the quota and usage.
func (m *Manager) EnsureQuota(usages types.ResourceList) error {
	query := &models.QuotaQuery{
		Reference:   m.reference,
		ReferenceID: m.referenceID,
	}
	quotas, err := dao.ListQuotas(query)
	if err != nil {
		return err
	}

	// non-existent: create quota and usage
	defaultHardLimit := m.driver.HardLimits()
	if len(quotas) == 0 {
		_, err := m.NewQuota(defaultHardLimit, usages)
		if err != nil {
			return err
		}
		return nil
	}

	// existent
	used := usages
	quotaUsed, err := types.NewResourceList(quotas[0].Used)
	if err != nil {
		return err
	}
	if types.Equals(quotaUsed, used) {
		return nil
	}
	dao.WithTransaction(func(o orm.Ormer) error {
		usage, err := m.getUsageForUpdate(o)
		if err != nil {
			return err
		}
		usage.Used = used.String()
		usage.UpdateTime = time.Now()
		_, err = o.Update(usage)
		if err != nil {
			return err
		}
		return nil
	})

	return nil
}

// AddResources add resources to usage
func (m *Manager) AddResources(resources types.ResourceList) error {
	return dao.WithTransaction(func(o orm.Ormer) error {
		return m.updateUsage(o, resources, types.Add)
	})
}

// SubtractResources subtract resources from usage
func (m *Manager) SubtractResources(resources types.ResourceList) error {
	return dao.WithTransaction(func(o orm.Ormer) error {
		return m.updateUsage(o, resources, types.Subtract)
	})
}

// NewManager returns quota manager
func NewManager(reference string, referenceID string) (*Manager, error) {
	d, ok := driver.Get(reference)
	if !ok {
		return nil, fmt.Errorf("quota not support for %s", reference)
	}

	if _, err := d.Load(referenceID); err != nil {
		log.Warning(fmt.Sprintf("Load quota reference object (%s, %s) failed: %v", reference, referenceID, err))
		return nil, err
	}

	return &Manager{
		driver:      d,
		reference:   reference,
		referenceID: referenceID,
	}, nil
}
