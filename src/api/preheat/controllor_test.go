package preheat

import (
	"context"
	"errors"
	"testing"

	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/history"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/instance"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/models"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
	"github.com/goharbor/harbor/src/testing/pkg/p2p/preheat/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type preheatSuite struct {
	suite.Suite
	controller    Controller
	instanceStore *dao.FakeInstanceStore
	historyStore  *dao.FakeHistoryStore
}

func TestPreheatSuite(t *testing.T) {
	instanceStore := &dao.FakeInstanceStore{}
	historyStore := &dao.FakeHistoryStore{}
	instance.StorageFactory = func() (instance.Storage, error) {
		return instanceStore, nil
	}
	history.StorageFactory = func() (history.Storage, error) {
		return historyStore, nil
	}
	c, err := NewCoreController(context.Background())
	assert.NotNil(t, c)
	assert.NoError(t, err)

	suite.Run(t, &preheatSuite{
		controller:    c,
		instanceStore: instanceStore,
		historyStore:  historyStore,
	})
}

func TestNewCoreController(t *testing.T) {
	instance.StorageFactory = func() (instance.Storage, error) {
		return nil, nil
	}
	history.StorageFactory = func() (history.Storage, error) {
		return nil, nil
	}
	c, err := NewCoreController(context.Background())
	assert.Nil(t, c)
	assert.Error(t, err)

	instance.StorageFactory = func() (instance.Storage, error) {
		return &dao.FakeInstanceStore{}, nil
	}
	history.StorageFactory = func() (history.Storage, error) {
		return &dao.FakeHistoryStore{}, nil
	}
	c, err = NewCoreController(context.Background())
	assert.NotNil(t, c)
	assert.NoError(t, err)
}

func (s *preheatSuite) SetupSuite() {
	config.Init()

	s.instanceStore.On("List", mock.Anything).Return([]*models.Metadata{
		{
			ID:       "i1",
			Provider: "dragonfly",
			Endpoint: "http://localhost",
			Status:   provider.DriverStatusHealthy,
			Enabled:  true,
		},
	}, nil)
	s.instanceStore.On("Save", mock.Anything).Return("i1", nil)
	s.instanceStore.On("Delete", "i1").Return(nil)
	s.instanceStore.On("Delete", "none").Return(errors.New("not found"))
	s.instanceStore.On("Get", "i1").Return(&models.Metadata{
		ID:       "i1",
		Endpoint: "http://localhost",
	}, nil)
	s.instanceStore.On("Get", "none").Return(nil, errors.New("not found"))
	s.instanceStore.On("Update", mock.Anything).Return(nil)

	s.historyStore.On("LoadHistories", mock.Anything).Return([]*models.HistoryRecord{
		{
			TaskID: "t1",
		},
	}, nil)
}

func (s *preheatSuite) TestGetAvailableProviders() {
	providers, err := s.controller.GetAvailableProviders()
	s.Equal(1, len(providers))
	s.Equal("dragonfly", providers[0].ID)
	s.NoError(err)

}

func (s *preheatSuite) TestListInstances() {
	instances, err := s.controller.ListInstances(nil)
	s.NoError(err)
	s.Equal(1, len(instances))
	s.Equal("i1", instances[0].ID)
}

func (s *preheatSuite) TestCreateInstance() {
	// Case: nil instance, expect error.
	id, err := s.controller.CreateInstance(nil)
	s.Empty(id)
	s.Error(err)

	// Case: instance with already existed endpoint, expect conflict.
	id, err = s.controller.CreateInstance(&models.Metadata{
		Endpoint: "http://localhost",
	})
	s.Equal(ErrorConflict, err)
	s.Empty(id)

	// Case: instance with invalid provider, expect error.
	id, err = s.controller.CreateInstance(&models.Metadata{
		Endpoint: "http://foo.bar",
		Provider: "none",
	})
	s.Error(err)
	s.Empty(id)

	// Case: instance with valid provider, expect ok.
	id, err = s.controller.CreateInstance(&models.Metadata{
		Endpoint: "http://foo.bar",
		Provider: "dragonfly",
	})
	s.NoError(err)
	s.Equal("i1", id)
}

func (s *preheatSuite) TestDeleteInstance() {
	err := s.controller.DeleteInstance("")
	s.Error(err)

	err = s.controller.DeleteInstance("none")
	s.Error(err)

	err = s.controller.DeleteInstance("i1")
	s.NoError(err)
}

func (s *preheatSuite) TestUpdateInstance() {
	err := s.controller.UpdateInstance("", nil)
	s.Error(err)

	err = s.controller.UpdateInstance("i1", nil)
	s.Error(err)

	err = s.controller.UpdateInstance("none", map[string]interface{}{"enabled": false})
	s.Error(err)

	err = s.controller.UpdateInstance("i1", map[string]interface{}{"enabled": false})
	s.NoError(err)
}

func (s *preheatSuite) TestPreheatImages() {
	// Case: no images provided
	result, err := s.controller.PreheatImages()
	s.Nil(result)
	s.Error(err)

	// Case: invalid images provided
	result, err = s.controller.PreheatImages("")
	s.Nil(result)
	s.Error(err)
	result, err = s.controller.PreheatImages("invalid")
	s.Nil(result)
	s.Error(err)
	result, err = s.controller.PreheatImages("library/alpine")
	s.Nil(result)
	s.Error(err)

	// Case: valid images provided, healthy instances available.
	s.instanceStore.On("List", mock.Anything).Return([]*models.Metadata{
		{
			ID:       "i1",
			Provider: "dragonfly",
			Endpoint: "http://localhost",
			Status:   provider.DriverStatusHealthy,
			Enabled:  true,
		},
	}, nil)
	result, err = s.controller.PreheatImages("library/alpine:latest")
	s.NotNil(result)
	s.NoError(err)
	s.Equal(1, len(result))
}

func (s *preheatSuite) TestLoadHistoryRecords() {
	records, err := s.controller.LoadHistoryRecords(nil)
	s.NoError(err)
	s.Equal(1, len(records))
}

func (s *preheatSuite) TestGetInstance() {
	instance, err := s.controller.GetInstance("none")
	s.Error(err)
	s.Nil(instance)

	instance, err = s.controller.GetInstance("i1")
	s.NoError(err)
	s.NotNil(instance)
}
