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

package artifact

import (
	"context"
	"testing"
	"time"

	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/project"
	repctl "github.com/goharbor/harbor/src/controller/replication"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/notification"
	policy_model "github.com/goharbor/harbor/src/pkg/notification/policy/model"
	reppkg "github.com/goharbor/harbor/src/pkg/replication"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/model"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	replicationtesting "github.com/goharbor/harbor/src/testing/controller/replication"
	"github.com/goharbor/harbor/src/testing/mock"
	testingnotification "github.com/goharbor/harbor/src/testing/pkg/notification/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakedReplicationRegistryMgr struct {
}

// Add new registry
func (f *fakedReplicationRegistryMgr) Add(*model.Registry) (int64, error) {
	return 0, nil
}

// List registries, returns total count, registry list and error
func (f *fakedReplicationRegistryMgr) List(query *q.Query) (int64, []*model.Registry, error) {
	return 0, nil, nil
}

// Get the specified registry
func (f *fakedReplicationRegistryMgr) Get(int64) (*model.Registry, error) {
	return &model.Registry{
		Type: "harbor",
		Credential: &model.Credential{
			Type: "local",
		},
	}, nil
}

// GetByName gets registry by name
func (f *fakedReplicationRegistryMgr) GetByName(name string) (*model.Registry, error) {
	return nil, nil
}

// Update the registry, the "props" are the properties of registry
// that need to be updated
func (f *fakedReplicationRegistryMgr) Update(registry *model.Registry, props ...string) error {
	return nil
}

// Remove the registry with the specified ID
func (f *fakedReplicationRegistryMgr) Remove(int64) error {
	return nil
}

// HealthCheck checks health status of all registries and update result in database
func (f *fakedReplicationRegistryMgr) HealthCheck() error {
	return nil
}

func TestReplicationHandler_Handle(t *testing.T) {
	common_dao.PrepareTestForPostgresSQL()
	config.Init()

	PolicyMgr := notification.PolicyMgr
	rpRegistry := replication.RegistryMgr
	prj := project.Ctl
	repCtl := repctl.Ctl

	defer func() {
		notification.PolicyMgr = PolicyMgr
		replication.RegistryMgr = rpRegistry
		project.Ctl = prj
		repctl.Ctl = repCtl
	}()
	policyMgrMock := &testingnotification.Manager{}
	notification.PolicyMgr = policyMgrMock
	replication.RegistryMgr = &fakedReplicationRegistryMgr{}
	projectCtl := &projecttesting.Controller{}
	project.Ctl = projectCtl
	mockRepCtl := &replicationtesting.Controller{}
	repctl.Ctl = mockRepCtl
	mockRepCtl.On("GetPolicy", mock.Anything, mock.Anything).Return(&reppkg.Policy{ID: 1}, nil)
	mockRepCtl.On("GetTask", mock.Anything, mock.Anything).Return(&repctl.Task{}, nil)
	mockRepCtl.On("GetExecution", mock.Anything, mock.Anything).Return(&repctl.Execution{}, nil)
	policyMgrMock.On("GetRelatedPolices", mock.Anything, mock.Anything, mock.Anything).Return([]*policy_model.Policy{
		{
			ID: 0,
		},
	}, nil)

	mock.OnAnything(projectCtl, "GetByName").Return(&models.Project{ProjectID: 1}, nil)

	handler := &ReplicationHandler{}

	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ReplicationHandler Want Error 1",
			args: args{
				data: "",
			},
			wantErr: true,
		},
		{
			name: "ReplicationHandler 1",
			args: args{
				data: &event.ReplicationEvent{
					OccurAt: time.Now(),
				},
			},
			wantErr: false,
		},
		{
			name: "ReplicationHandler with docker registry",
			args: args{
				data: &event.ReplicationEvent{
					OccurAt:           time.Now(),
					ReplicationTaskID: 1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Handle(context.TODO(), tt.args.data)
			if tt.wantErr {
				require.NotNil(t, err, "Error: %s", err)
				return
			}
			assert.Nil(t, err)
		})
	}

}

func TestReplicationHandler_IsStateful(t *testing.T) {
	handler := &ReplicationHandler{}
	assert.False(t, handler.IsStateful())
}

func TestReplicationHandler_Name(t *testing.T) {
	handler := &ReplicationHandler{}
	assert.Equal(t, "ReplicationWebhook", handler.Name())
}
