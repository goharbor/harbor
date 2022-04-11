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

package internal

import (
	"context"
	"github.com/goharbor/harbor/src/lib/q"
	accel "github.com/goharbor/harbor/src/pkg/acceleration"
	accelModel "github.com/goharbor/harbor/src/pkg/acceleration/model"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/acceleration/adapter"
)

// autoScan scan artifact when the project of the artifact enable auto scan
func autoScan(ctx context.Context, a *artifact.Artifact, tags ...string) error {
	proj, err := project.Ctl.Get(ctx, a.ProjectID)
	if err != nil {
		return err
	}
	if !proj.AutoScan() {
		return nil
	}

	// transaction here to work with the image index
	return orm.WithTransaction(func(ctx context.Context) error {
		options := []scan.Option{}
		if len(tags) > 0 {
			options = append(options, scan.WithTag(tags[0]))
		}

		return scan.DefaultController.Scan(ctx, a, options...)
	})(orm.SetTransactionOpNameToContext(ctx, "tx-auto-scan"))
}

// autoAcc accelerate artifact
func autoAcc(ctx context.Context, a *artifact.Artifact, tags ...string) error {
	proj, err := project.Ctl.Get(ctx, a.ProjectID)
	if err != nil {
		return err
	}
	if !proj.AutoAcc() {
		return nil
	}

	// convert
	acceleration, err := adapter.GetFactory(accelModel.AccelerationTypeNydus)
	if err != nil {
		return err
	}
	accs, err := accel.Mgr.List(ctx, q.New(q.KeyWords{"type": accelModel.AccelerationTypeNydus}))
	if err != nil {
		return err
	}
	adapter, err := acceleration.Create(accs[0])
	if err != nil {
		return err
	}
	var tagName string
	if len(tags) > 0 {
		tagName = tags[0]
	}
	err = adapter.Convert(&a.Artifact, tagName)
	if err != nil {
		return err
	}
	return nil
}
