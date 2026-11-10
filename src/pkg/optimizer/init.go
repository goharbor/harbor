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

package optimizer

import (
	"context"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/optimizer/dao"
)

// EnsureOptimizers ensures that the optimizers with the specified endpoint URLs exist in the system.
func EnsureOptimizers(ctx context.Context, wantedOptimizers []dao.Registration) (err error) {
	if len(wantedOptimizers) == 0 {
		return
	}
	names := make([]string, len(wantedOptimizers))
	for i, wo := range wantedOptimizers {
		names[i] = wo.Name
	}

	list, err := Mgr.List(ctx, q.New(q.KeyWords{"name__in": names}))
	if err != nil {
		return errors.Errorf("listing optimizers: %v", err)
	}
	existing := make(map[string]*dao.Registration)
	for _, li := range list {
		existing[li.Name] = li
	}

	for _, wo := range wantedOptimizers {
		reg, exists := existing[wo.Name]
		if !exists {
			if _, err := Mgr.Create(ctx, &wo); err != nil {
				return errors.Errorf("creating optimizer registration %s at %s failed: %v", wo.Name, wo.URL, err)
			}
			log.Infof("Successfully registered %s optimizer at %s", wo.Name, wo.URL)
		} else if reg.URL != wo.URL {
			reg.URL = wo.URL
			if err := Mgr.Update(ctx, reg); err != nil {
				return errors.Errorf("updating optimizer registration %s to %s failed: %v", wo.Name, wo.URL, err)
			}
			log.Infof("Successfully updated %s optimizer to %s", wo.Name, wo.URL)
		} else {
			log.Infof("Optimizer registration already exists: %s", wo.URL)
		}
	}

	return
}

// EnsureDefaultOptimizer ensures that the optimizer with the specified name is set as
// default in the system, unless a default is already configured.
func EnsureDefaultOptimizer(ctx context.Context, optimizerName string) (err error) {
	defaultOptimizer, err := Mgr.GetDefault(ctx)
	if err != nil {
		err = errors.Errorf("getting default optimizer: %v", err)
		return
	}
	if defaultOptimizer != nil {
		log.Infof("Skipped setting %s as the default optimizer. The default optimizer is already set to %s", optimizerName, defaultOptimizer.URL)
		return
	}
	optimizers, err := Mgr.List(ctx, q.New(q.KeyWords{"name": optimizerName}))
	if err != nil {
		err = errors.Errorf("listing optimizers: %v", err)
		return
	}
	if len(optimizers) != 1 {
		return errors.Errorf("expected only one optimizer with name %v but got %d", optimizerName, len(optimizers))
	}
	err = Mgr.SetAsDefault(ctx, optimizers[0].UUID)
	if err != nil {
		err = errors.Errorf("setting %s as default optimizer: %v", optimizerName, err)
	}
	return
}

// RemoveImmutableOptimizers removes immutable optimizer registrations with the specified names.
func RemoveImmutableOptimizers(ctx context.Context, names []string) error {
	if len(names) == 0 {
		return nil
	}
	query := q.New(q.KeyWords{"immutable": true, "name__in": names})

	registrations, err := Mgr.List(ctx, query)
	if err != nil {
		return errors.Errorf("listing optimizers: %v", err)
	}

	for _, reg := range registrations {
		if err := Mgr.Delete(ctx, reg.UUID); err != nil {
			return errors.Errorf("deleting optimizer: %s: %v", reg.UUID, err)
		}
	}

	return nil
}
