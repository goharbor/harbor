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
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication/model"
	trans "github.com/goharbor/harbor/src/replication/transfer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRegistry struct{}

func (f *fakeRegistry) FetchCharts(filters []*model.Filter) ([]*model.Resource, error) {
	return []*model.Resource{
		{
			Type: model.ResourceTypeChart,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/harbor",
				},
				Vtags: []string{"0.2.0"},
			},
		},
	}, nil
}
func (f *fakeRegistry) ChartExist(name, version string) (bool, error) {
	return true, nil
}
func (f *fakeRegistry) DownloadChart(name, version string) (io.ReadCloser, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte{'a'}))
	return r, nil
}
func (f *fakeRegistry) UploadChart(name, version string, chart io.Reader) error {
	return nil
}
func (f *fakeRegistry) DeleteChart(name, version string) error {
	return nil
}

func TestFactory(t *testing.T) {
	tr, err := factory(nil, nil)
	require.Nil(t, err)
	_, ok := tr.(trans.Transfer)
	assert.True(t, ok)
}

func TestShouldStop(t *testing.T) {
	// should stop
	stopFunc := func() bool { return true }
	tr := &transfer{
		logger:    log.DefaultLogger(),
		isStopped: stopFunc,
	}
	assert.True(t, tr.shouldStop())

	// should not stop
	stopFunc = func() bool { return false }
	tr = &transfer{
		isStopped: stopFunc,
	}
	assert.False(t, tr.shouldStop())
}

func TestCopy(t *testing.T) {
	stopFunc := func() bool { return false }
	transfer := &transfer{
		logger:    log.DefaultLogger(),
		isStopped: stopFunc,
		src:       &fakeRegistry{},
		dst:       &fakeRegistry{},
	}
	src := &chart{
		name:    "library/harbor",
		version: "0.2.0",
	}
	dst := &chart{
		name:    "dest/harbor",
		version: "0.2.0",
	}
	err := transfer.copy(src, dst, true)
	assert.Nil(t, err)
}

func TestDelete(t *testing.T) {
	stopFunc := func() bool { return false }
	transfer := &transfer{
		logger:    log.DefaultLogger(),
		isStopped: stopFunc,
		src:       &fakeRegistry{},
		dst:       &fakeRegistry{},
	}
	chart := &chart{
		name:    "dest/harbor",
		version: "0.2.0",
	}
	err := transfer.delete(chart)
	assert.Nil(t, err)
}
