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

package chart

import (
	"errors"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	trans "github.com/goharbor/harbor/src/replication/transfer"
)

func init() {
	if err := trans.RegisterFactory(model.ResourceTypeChart, factory); err != nil {
		log.Errorf("failed to register transfer factory: %v", err)
	}
}

func factory(logger trans.Logger, stopFunc trans.StopFunc) (trans.Transfer, error) {
	return &transfer{
		logger:    logger,
		isStopped: stopFunc,
	}, nil
}

type chart struct {
	name    string
	version string
}

type transfer struct {
	logger    trans.Logger
	isStopped trans.StopFunc
	src       adapter.ChartRegistry
	dst       adapter.ChartRegistry
}

func (t *transfer) Transfer(src *model.Resource, dst *model.Resource) error {
	// initialize
	if err := t.initialize(src, dst); err != nil {
		return err
	}

	// delete the chart on destination registry
	if dst.Deleted {
		return t.delete(&chart{
			name:    dst.Metadata.GetResourceName(),
			version: dst.Metadata.Vtags[0],
		})
	}

	srcChart := &chart{
		name:    src.Metadata.GetResourceName(),
		version: src.Metadata.Vtags[0],
	}
	dstChart := &chart{
		name:    dst.Metadata.GetResourceName(),
		version: dst.Metadata.Vtags[0],
	}
	// copy the chart from source registry to the destination
	return t.copy(srcChart, dstChart, dst.Override)
}

func (t *transfer) initialize(src, dst *model.Resource) error {
	if t.shouldStop() {
		return nil
	}
	// create client for source registry
	srcReg, err := createRegistry(src.Registry)
	if err != nil {
		t.logger.Errorf("failed to create client for source registry: %v", err)
		return err
	}
	t.src = srcReg
	t.logger.Infof("client for source registry [type: %s, URL: %s, insecure: %v] created",
		src.Registry.Type, src.Registry.URL, src.Registry.Insecure)

	// create client for destination registry
	dstReg, err := createRegistry(dst.Registry)
	if err != nil {
		t.logger.Errorf("failed to create client for destination registry: %v", err)
		return err
	}
	t.dst = dstReg
	t.logger.Infof("client for destination registry [type: %s, URL: %s, insecure: %v] created",
		dst.Registry.Type, dst.Registry.URL, dst.Registry.Insecure)

	return nil
}

func createRegistry(reg *model.Registry) (adapter.ChartRegistry, error) {
	factory, err := adapter.GetFactory(reg.Type)
	if err != nil {
		return nil, err
	}
	ad, err := factory.Create(reg)
	if err != nil {
		return nil, err
	}
	registry, ok := ad.(adapter.ChartRegistry)
	if !ok {
		return nil, errors.New("the adapter doesn't implement the \"ChartRegistry\" interface")
	}
	return registry, nil
}

func (t *transfer) shouldStop() bool {
	isStopped := t.isStopped()
	if isStopped {
		t.logger.Info("the job is stopped")
	}
	return isStopped
}

func (t *transfer) copy(src, dst *chart, override bool) error {
	if t.shouldStop() {
		return nil
	}
	t.logger.Infof("copying %s:%s(source registry) to %s:%s(destination registry)...",
		src.name, src.version, dst.name, dst.version)

	// check the existence of the chart on the destination registry
	exist, err := t.dst.ChartExist(dst.name, dst.version)
	if err != nil {
		t.logger.Errorf("failed to check the existence of chart %s:%s on the destination registry: %v", dst.name, dst.version, err)
		return err
	}
	if exist {
		// the same name chart exists, but not allowed to override
		if !override {
			t.logger.Warningf("the same name chart %s:%s exists on the destination registry, but the \"override\" is set to false, skip",
				dst.name, dst.version)
			return nil
		}
		// the same name chart exists, but allowed to override
		t.logger.Warningf("the same name chart %s:%s exists on the destination registry and the \"override\" is set to true, continue...",
			dst.name, dst.version)
	}

	// copy the chart between the source and destination registries
	chart, err := t.src.DownloadChart(src.name, src.version)
	if err != nil {
		t.logger.Errorf("failed to download the chart %s:%s: %v", src.name, src.version, err)
		return err
	}
	defer chart.Close()

	if err = t.dst.UploadChart(dst.name, dst.version, chart); err != nil {
		t.logger.Errorf("failed to upload the chart %s:%s: %v", dst.name, dst.version, err)
		return err
	}

	t.logger.Infof("copy %s:%s(source registry) to %s:%s(destination registry) completed",
		src.name, src.version, dst.name, dst.version)

	return nil
}

func (t *transfer) delete(chart *chart) error {
	exist, err := t.dst.ChartExist(chart.name, chart.version)
	if err != nil {
		t.logger.Errorf("failed to check the existence of chart %s:%s on the destination registry: %v", chart.name, chart.version, err)
		return err
	}
	if !exist {
		t.logger.Infof("the chart %s:%s doesn't exist on the destination registry, skip",
			chart.name, chart.version)
		return nil
	}

	t.logger.Infof("deleting the chart %s:%s on the destination registry...", chart.name, chart.version)
	if err := t.dst.DeleteChart(chart.name, chart.version); err != nil {
		t.logger.Errorf("failed to delete the chart %s:%s on the destination registry: %v", chart.name, chart.version, err)
		return err
	}
	t.logger.Infof("delete the chart %s:%s on the destination registry completed", chart.name, chart.version)
	return nil
}
