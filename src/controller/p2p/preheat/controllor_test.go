package preheat

import (
	"context"
	"errors"
	"fmt"
	"github.com/goharbor/harbor/src/lib/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/policy"
	providerModel "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/auth"
	taskModel "github.com/goharbor/harbor/src/pkg/task"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/pkg/p2p/preheat/instance"
	pmocks "github.com/goharbor/harbor/src/testing/pkg/p2p/preheat/policy"
	smocks "github.com/goharbor/harbor/src/testing/pkg/scheduler"
	tmocks "github.com/goharbor/harbor/src/testing/pkg/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type preheatSuite struct {
	suite.Suite
	ctx                context.Context
	controller         Controller
	fakeInstanceMgr    *instance.FakeManager
	fakePolicyMgr      *pmocks.FakeManager
	fakeScheduler      *smocks.Scheduler
	mockInstanceServer *httptest.Server
	fakeExecutionMgr   *tmocks.ExecutionManager
}

func TestPreheatSuite(t *testing.T) {
	t.Log("Start TestPreheatSuite")
	fakeInstanceMgr := &instance.FakeManager{}
	fakePolicyMgr := &pmocks.FakeManager{}
	fakeScheduler := &smocks.Scheduler{}
	fakeExecutionMgr := &tmocks.ExecutionManager{}

	var c = &controller{
		iManager:     fakeInstanceMgr,
		pManager:     fakePolicyMgr,
		scheduler:    fakeScheduler,
		executionMgr: fakeExecutionMgr,
	}
	assert.NotNil(t, c)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	suite.Run(t, &preheatSuite{
		ctx:              ctx,
		controller:       c,
		fakeInstanceMgr:  fakeInstanceMgr,
		fakePolicyMgr:    fakePolicyMgr,
		fakeScheduler:    fakeScheduler,
		fakeExecutionMgr: fakeExecutionMgr,
	})
}

func TestNewController(t *testing.T) {
	c := NewController()
	assert.NotNil(t, c)
}

func (s *preheatSuite) SetupSuite() {
	config.Init()

	s.fakeInstanceMgr.On("List", mock.Anything, mock.Anything).Return([]*providerModel.Instance{
		{
			ID:       1,
			Vendor:   "dragonfly",
			Endpoint: "http://localhost",
			Status:   provider.DriverStatusHealthy,
			Enabled:  true,
		},
	}, nil)
	s.fakeInstanceMgr.On("Save", mock.Anything, mock.Anything).Return(int64(1), nil)
	s.fakeInstanceMgr.On("Count", mock.Anything, &q.Query{Keywords: map[string]interface{}{
		"endpoint": "http://localhost",
	}}).Return(int64(1), nil)
	s.fakeInstanceMgr.On("Count", mock.Anything, mock.Anything).Return(int64(0), nil)
	s.fakeInstanceMgr.On("Delete", mock.Anything, int64(1)).Return(nil)
	s.fakeInstanceMgr.On("Delete", mock.Anything, int64(0)).Return(errors.New("not found"))
	s.fakeInstanceMgr.On("Get", mock.Anything, int64(1)).Return(&providerModel.Instance{
		ID:       1,
		Endpoint: "http://localhost",
	}, nil)
	s.fakeInstanceMgr.On("Get", mock.Anything, int64(0)).Return(nil, errors.New("not found"))

	// mock server for check health
	s.mockInstanceServer = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/_ping":
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}

			w.WriteHeader(http.StatusOK)
		}
	}))
	s.mockInstanceServer.Start()
}

// TearDownSuite clears the env.
func (s *preheatSuite) TearDownSuite() {
	s.mockInstanceServer.Close()
}

func (s *preheatSuite) TestGetAvailableProviders() {
	providers, err := s.controller.GetAvailableProviders()
	s.Equal(2, len(providers))
	expectProviders := map[string]interface{}{}
	expectProviders["dragonfly"] = nil
	expectProviders["kraken"] = nil
	_, ok := expectProviders[providers[0].ID]
	s.True(ok)
	_, ok = expectProviders[providers[1].ID]
	s.True(ok)
	s.NoError(err)
}

func (s *preheatSuite) TestListInstance() {
	instances, err := s.controller.ListInstance(s.ctx, nil)
	s.NoError(err)
	s.Equal(1, len(instances))
	s.Equal(int64(1), instances[0].ID)
}

func (s *preheatSuite) TestCreateInstance() {
	// Case: nil instance, expect error.
	id, err := s.controller.CreateInstance(s.ctx, nil)
	s.Empty(id)
	s.Error(err)

	// Case: instance with already existed endpoint, expect conflict.
	id, err = s.controller.CreateInstance(s.ctx, &providerModel.Instance{
		Endpoint: "http://localhost",
	})
	s.Equal(ErrorConflict, err)
	s.Empty(id)

	// Case: instance with invalid provider, expect error.
	id, err = s.controller.CreateInstance(s.ctx, &providerModel.Instance{
		Endpoint: "http://foo.bar",
		Status:   "healthy",
		Vendor:   "none",
	})
	s.NoError(err)
	s.Equal(int64(1), id)

	// Case: instance with valid provider, expect ok.
	id, err = s.controller.CreateInstance(s.ctx, &providerModel.Instance{
		Endpoint: "http://foo.bar",
		Status:   "healthy",
		Vendor:   "dragonfly",
	})
	s.NoError(err)
	s.Equal(int64(1), id)

	id, err = s.controller.CreateInstance(s.ctx, &providerModel.Instance{
		Endpoint: "http://foo.bar2",
		Status:   "healthy",
		Vendor:   "kraken",
	})
	s.NoError(err)
	s.Equal(int64(1), id)
}

func (s *preheatSuite) TestDeleteInstance() {
	// instance be used should not be deleted
	s.fakeInstanceMgr.On("Get", s.ctx, int64(1)).Return(&providerModel.Instance{ID: 1}, nil)
	s.fakePolicyMgr.On("ListPolicies", s.ctx, &q.Query{Keywords: map[string]interface{}{"provider_id": int64(1)}}).Return([]*policy.Schema{
		{
			ProviderID: 1,
		},
	}, nil)
	err := s.controller.DeleteInstance(s.ctx, int64(1))
	s.Error(err, "instance should not be deleted")

	s.fakeInstanceMgr.On("Get", s.ctx, int64(2)).Return(&providerModel.Instance{ID: 2}, nil)
	s.fakePolicyMgr.On("ListPolicies", s.ctx, &q.Query{Keywords: map[string]interface{}{"provider_id": int64(2)}}).Return([]*policy.Schema{}, nil)
	s.fakeInstanceMgr.On("Delete", s.ctx, int64(2)).Return(nil)
	err = s.controller.DeleteInstance(s.ctx, int64(2))
	s.NoError(err, "instance can be deleted")
}

func (s *preheatSuite) TestUpdateInstance() {
	// normal update
	s.fakeInstanceMgr.On("Get", s.ctx, int64(1000)).Return(&providerModel.Instance{ID: 1000}, nil)
	s.fakeInstanceMgr.On("Update", s.ctx, &providerModel.Instance{ID: 1000, Enabled: true}).Return(nil)
	err := s.controller.UpdateInstance(s.ctx, &providerModel.Instance{ID: 1000, Enabled: true})
	s.NoError(err, "instance can be updated")

	// disable instance should error due to with policy used
	s.fakeInstanceMgr.On("Get", s.ctx, int64(1001)).Return(&providerModel.Instance{ID: 1001}, nil)
	s.fakeInstanceMgr.On("Update", s.ctx, &providerModel.Instance{ID: 1001}).Return(nil)
	s.fakePolicyMgr.On("ListPolicies", s.ctx, &q.Query{Keywords: map[string]interface{}{"provider_id": int64(1001)}}).Return([]*policy.Schema{
		{ProviderID: 1001},
	}, nil)
	err = s.controller.UpdateInstance(s.ctx, &providerModel.Instance{ID: 1001})
	s.Error(err, "instance should not be disabled")

	// disable instance can be deleted if no policy used
	s.fakeInstanceMgr.On("Get", s.ctx, int64(1002)).Return(&providerModel.Instance{ID: 1002}, nil)
	s.fakeInstanceMgr.On("Update", s.ctx, &providerModel.Instance{ID: 1002}).Return(nil)
	s.fakePolicyMgr.On("ListPolicies", s.ctx, &q.Query{Keywords: map[string]interface{}{"provider_id": int64(1002)}}).Return([]*policy.Schema{}, nil)
	err = s.controller.UpdateInstance(s.ctx, &providerModel.Instance{ID: 1002})
	s.NoError(err, "instance can be disabled")

	// not support change vendor type
	s.fakeInstanceMgr.On("Get", s.ctx, int64(1003)).Return(&providerModel.Instance{ID: 1003, Vendor: "dragonfly"}, nil)
	s.fakeInstanceMgr.On("Update", s.ctx, &providerModel.Instance{ID: 1003, Vendor: "kraken"}).Return(nil)
	s.fakePolicyMgr.On("ListPolicies", s.ctx, &q.Query{Keywords: map[string]interface{}{"provider_id": int64(1003)}}).Return([]*policy.Schema{}, nil)
	err = s.controller.UpdateInstance(s.ctx, &providerModel.Instance{ID: 1003, Vendor: "kraken"})
	s.Error(err, "provider vendor cannot be changed")
}

func (s *preheatSuite) TestGetInstance() {
	inst, err := s.controller.GetInstance(s.ctx, 1)
	s.NoError(err)
	s.NotNil(inst)
}

func (s *preheatSuite) TestCountPolicy() {
	s.fakePolicyMgr.On("Count", s.ctx, mock.Anything).Return(int64(1), nil)
	id, err := s.controller.CountPolicy(s.ctx, nil)
	s.NoError(err)
	s.Equal(int64(1), id)
}

func (s *preheatSuite) TestCreatePolicy() {
	policy := &policy.Schema{
		Name:       "test",
		FiltersStr: `[{"type":"repository","value":"harbor*"},{"type":"tag","value":"2*"}]`,
		TriggerStr: fmt.Sprintf(`{"type":"%s", "trigger_setting":{"cron":"* * * * */1"}}`, policy.TriggerTypeScheduled),
	}
	s.fakeScheduler.On("Schedule", s.ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	s.fakePolicyMgr.On("Create", s.ctx, policy).Return(int64(1), nil)
	s.fakePolicyMgr.On("Update", s.ctx, mock.Anything, mock.Anything).Return(nil)
	s.fakeScheduler.On("UnScheduleByVendor", s.ctx, mock.Anything, mock.Anything).Return(nil)
	id, err := s.controller.CreatePolicy(s.ctx, policy)
	s.NoError(err)
	s.Equal(int64(1), id)
	s.False(policy.CreatedAt.IsZero())
	s.False(policy.UpdatedTime.IsZero())
}

func (s *preheatSuite) TestGetPolicy() {
	s.fakePolicyMgr.On("Get", s.ctx, int64(1)).Return(&policy.Schema{Name: "test"}, nil)
	p, err := s.controller.GetPolicy(s.ctx, 1)
	s.NoError(err)
	s.Equal("test", p.Name)
}

func (s *preheatSuite) TestGetPolicyByName() {
	s.fakePolicyMgr.On("GetByName", s.ctx, int64(1), "test").Return(&policy.Schema{Name: "test"}, nil)
	p, err := s.controller.GetPolicyByName(s.ctx, 1, "test")
	s.NoError(err)
	s.Equal("test", p.Name)
}

func (s *preheatSuite) TestUpdatePolicy() {
	var p0 = &policy.Schema{Name: "test", Trigger: &policy.Trigger{Type: policy.TriggerTypeScheduled}}
	p0.Trigger.Settings.Cron = "* * * * */1"
	p0.Filters = []*policy.Filter{
		{
			Type:  policy.FilterTypeRepository,
			Value: "harbor*",
		},
		{
			Type:  policy.FilterTypeTag,
			Value: "2*",
		},
	}
	s.fakePolicyMgr.On("Get", s.ctx, int64(1)).Return(p0, nil)

	// need change to schedule
	p1 := &policy.Schema{
		ID:         1,
		Name:       "test",
		FiltersStr: `[{"type":"repository","value":"harbor*"},{"type":"tag","value":"2*"}]`,
		TriggerStr: fmt.Sprintf(`{"type":"%s", "trigger_setting":{}}`, policy.TriggerTypeManual),
	}
	s.fakePolicyMgr.On("Update", s.ctx, p1, mock.Anything).Return(nil)
	err := s.controller.UpdatePolicy(s.ctx, p1, "")
	s.NoError(err)
	s.False(p1.UpdatedTime.IsZero())

	// need update schedule
	p2 := &policy.Schema{
		ID:         1,
		Name:       "test",
		FiltersStr: `[{"type":"repository","value":"harbor*"},{"type":"tag","value":"2*"}]`,
		TriggerStr: fmt.Sprintf(`{"type":"%s", "trigger_setting":{"cron":"* * * * */2"}}`, policy.TriggerTypeScheduled),
	}
	s.fakePolicyMgr.On("Update", s.ctx, p2, mock.Anything).Return(nil)
	err = s.controller.UpdatePolicy(s.ctx, p2, "")
	s.NoError(err)
	s.False(p2.UpdatedTime.IsZero())
}

func (s *preheatSuite) TestDeletePolicy() {
	var p0 = &policy.Schema{Name: "test", Trigger: &policy.Trigger{Type: policy.TriggerTypeScheduled}}
	s.fakePolicyMgr.On("Get", s.ctx, int64(1)).Return(p0, nil)
	s.fakeExecutionMgr.On("List", s.ctx, mock.AnythingOfType("*q.Query")).Return(
		[]*taskModel.Execution{
			{ID: 1},
			{ID: 2},
		}, nil,
	)
	s.fakeExecutionMgr.On("Delete", mock.Anything, mock.Anything).Return(nil)
	s.fakePolicyMgr.On("Delete", s.ctx, int64(1)).Return(nil)
	err := s.controller.DeletePolicy(s.ctx, 1)
	s.NoError(err)
}

func (s *preheatSuite) TestListPolicies() {
	s.fakePolicyMgr.On("ListPolicies", s.ctx, &q.Query{}).Return([]*policy.Schema{}, nil)
	p, err := s.controller.ListPolicies(s.ctx, &q.Query{})
	s.NoError(err)
	s.NotNil(p)
}

func (s *preheatSuite) TestListPoliciesByProject() {
	s.fakePolicyMgr.On("ListPoliciesByProject", s.ctx, int64(1), mock.Anything).Return([]*policy.Schema{}, nil)
	p, err := s.controller.ListPoliciesByProject(s.ctx, 1, nil)
	s.NoError(err)
	s.NotNil(p)
}

func (s *preheatSuite) TestDeletePoliciesOfProject() {
	fakePolicies := []*policy.Schema{
		{ID: 1000, Name: "1-should-delete", ProjectID: 10},
		{ID: 1001, Name: "2-should-delete", ProjectID: 10},
	}
	s.fakePolicyMgr.On("ListPoliciesByProject", s.ctx, int64(10), mock.Anything).Return(fakePolicies, nil)
	for _, p := range fakePolicies {
		s.fakePolicyMgr.On("Get", s.ctx, p.ID).Return(p, nil)
		s.fakePolicyMgr.On("Delete", s.ctx, p.ID).Return(nil)
		s.fakeExecutionMgr.On("List", s.ctx, &q.Query{Keywords: map[string]interface{}{"VendorID": p.ID, "VendorType": "P2P_PREHEAT"}}).Return([]*taskModel.Execution{}, nil)
	}

	err := s.controller.DeletePoliciesOfProject(s.ctx, 10)
	s.NoError(err)
}

func (s *preheatSuite) TestCheckHealth() {
	// if instance is nil
	var instance *providerModel.Instance
	err := s.controller.CheckHealth(s.ctx, instance)
	s.Error(err)

	// unknown vendor
	instance = &providerModel.Instance{
		ID:       1,
		Name:     "test-instance",
		Vendor:   "unknown",
		Endpoint: "http://127.0.0.1",
		AuthMode: auth.AuthModeNone,
		Enabled:  true,
		Default:  true,
		Insecure: true,
		Status:   "Unknown",
	}
	err = s.controller.CheckHealth(s.ctx, instance)
	s.Error(err)

	// not health
	// health
	instance = &providerModel.Instance{
		ID:       1,
		Name:     "test-instance",
		Vendor:   provider.DriverDragonfly,
		Endpoint: "http://127.0.0.1",
		AuthMode: auth.AuthModeNone,
		Enabled:  true,
		Default:  true,
		Insecure: true,
		Status:   "Unknown",
	}
	err = s.controller.CheckHealth(s.ctx, instance)
	s.Error(err)

	// health
	instance = &providerModel.Instance{
		ID:       1,
		Name:     "test-instance",
		Vendor:   provider.DriverDragonfly,
		Endpoint: s.mockInstanceServer.URL,
		AuthMode: auth.AuthModeNone,
		Enabled:  true,
		Default:  true,
		Insecure: true,
		Status:   "Unknown",
	}
	err = s.controller.CheckHealth(s.ctx, instance)
	s.NoError(err)
}
