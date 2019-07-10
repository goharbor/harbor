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
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
)

// Manager manager for quota
type Manager struct {
	reference   string
	referenceID string
}

func (m *Manager) addQuota(o orm.Ormer, hardLimits ResourceList, now time.Time) (int64, error) {
	quota := &models.Quota{
		Reference:    m.reference,
		ReferenceID:  m.referenceID,
		Hard:         hardLimits.String(),
		CreationTime: now,
		UpdateTime:   now,
	}

	return o.Insert(quota)
}

func (m *Manager) addUsage(o orm.Ormer, hardLimits, used ResourceList, now time.Time, ids ...int64) (int64, error) {
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

func (m *Manager) getQuotaForUpdate(o orm.Ormer) (*models.Quota, error) {
	quota := &models.Quota{Reference: m.reference, ReferenceID: m.referenceID}
	if err := o.ReadForUpdate(quota, "reference", "reference_id"); err != nil {
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

func (m *Manager) updateUsage(o orm.Ormer, resources ResourceList, calculate func(ResourceList, ResourceList) ResourceList) error {
	quota, err := m.getQuotaForUpdate(o)
	if err != nil {
		return err
	}
	hardLimits, err := NewResourceList(quota.Hard)
	if err != nil {
		return err
	}

	usage, err := m.getUsageForUpdate(o)
	if err != nil {
		return err
	}
	used, err := NewResourceList(usage.Used)
	if err != nil {
		return err
	}

	newUsed := calculate(used, resources)
	if err := isSafe(hardLimits, newUsed); err != nil {
		return err
	}

	usage.Used = newUsed.String()
	usage.UpdateTime = time.Now()

	_, err = o.Update(usage)
	return err
}

// NewQuota create new quota for (reference, reference id)
func (m *Manager) NewQuota(hardLimits ResourceList, usages ...ResourceList) (int64, error) {
	var quotaID int64

	err := dao.WithTransaction(func(o orm.Ormer) error {
		now := time.Now()

		var err error
		quotaID, err = m.addQuota(o, hardLimits, now)
		if err != nil {
			return err
		}

		var used ResourceList
		if len(usages) > 0 {
			used = usages[0]
		} else {
			used = ResourceList{}
		}

		if _, err := m.addUsage(o, hardLimits, used, now, quotaID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return quotaID, nil
}

// AddResources add resources to usage
func (m *Manager) AddResources(resources ResourceList) error {
	return dao.WithTransaction(func(o orm.Ormer) error {
		return m.updateUsage(o, resources, Add)
	})
}

// SubtractResources subtract resources from usage
func (m *Manager) SubtractResources(resources ResourceList) error {
	return dao.WithTransaction(func(o orm.Ormer) error {
		return m.updateUsage(o, resources, Subtract)
	})
}

// NewManager returns quota manager
func NewManager(reference string, referenceID string) (*Manager, error) {
	return &Manager{
		reference:   reference,
		referenceID: referenceID,
	}, nil
}
