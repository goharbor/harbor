package preheat

import (
	"context"
	"errors"
	"testing"

	"github.com/goharbor/harbor/src/core/config"
	providerModel "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
	"github.com/goharbor/harbor/src/testing/pkg/p2p/preheat/instance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type preheatSuite struct {
	suite.Suite
	ctx         context.Context
	controller  Controller
	fackManager *instance.FakeManager
}

func TestPreheatSuite(t *testing.T) {
	t.Log("Start TestPreheatSuite")
	fackManager := &instance.FakeManager{}

	var c = &controller{
		iManager: fackManager,
	}
	assert.NotNil(t, c)

	suite.Run(t, &preheatSuite{
		ctx:         context.Background(),
		controller:  c,
		fackManager: fackManager,
	})
}

func TestNewController(t *testing.T) {
	c := NewController()
	assert.NotNil(t, c)
}

func (s *preheatSuite) SetupSuite() {
	config.Init()

	s.fackManager.On("List", mock.Anything, mock.Anything).Return([]*providerModel.Instance{
		{
			ID:       1,
			Vendor:   "dragonfly",
			Endpoint: "http://localhost",
			Status:   provider.DriverStatusHealthy,
			Enabled:  true,
		},
	}, nil)
	s.fackManager.On("Save", mock.Anything, mock.Anything).Return(int64(1), nil)
	s.fackManager.On("Count", mock.Anything, &providerModel.Instance{Endpoint: "http://localhost"}).Return(int64(1), nil)
	s.fackManager.On("Count", mock.Anything, mock.Anything).Return(int64(0), nil)
	s.fackManager.On("Delete", mock.Anything, int64(1)).Return(nil)
	s.fackManager.On("Delete", mock.Anything, int64(0)).Return(errors.New("not found"))
	s.fackManager.On("Get", mock.Anything, int64(1)).Return(&providerModel.Instance{
		ID:       1,
		Endpoint: "http://localhost",
	}, nil)
	s.fackManager.On("Get", mock.Anything, int64(0)).Return(nil, errors.New("not found"))
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
	s.Equal(ErrorUnhealthy, err)
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
	// err := s.controller.DeleteInstance(s.ctx, 0)
	// s.Error(err)

	err := s.controller.DeleteInstance(s.ctx, int64(1))
	s.NoError(err)
}

func (s *preheatSuite) TestUpdateInstance() {
	// TODO: test update more
	s.fackManager.On("Update", s.ctx, nil).Return(errors.New("no properties provided to update"))
	err := s.controller.UpdateInstance(s.ctx, nil)
	s.Error(err)

	err = s.controller.UpdateInstance(s.ctx, &providerModel.Instance{ID: 0})
	s.Error(err)

	s.fackManager.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err = s.controller.UpdateInstance(s.ctx, &providerModel.Instance{ID: 1}, "enabled")
	s.NoError(err)
}

func (s *preheatSuite) TestGetInstance() {
	inst, err := s.controller.GetInstance(s.ctx, 1)
	s.NoError(err)
	s.NotNil(inst)
}
