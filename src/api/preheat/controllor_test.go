package preheat

import (
	"context"
	"errors"
	"testing"

	"github.com/goharbor/harbor/src/core/config"
	hmocks "github.com/goharbor/harbor/src/pkg/p2p/preheat/history/mocks"
	imocks "github.com/goharbor/harbor/src/pkg/p2p/preheat/instance/mocks"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type preheatSuite struct {
	suite.Suite
	controller    Controller
	instanceStore *imocks.Manager
	historyStore  *hmocks.Manager
}

func TestPreheatSuite(t *testing.T) {
	instanceStore := &imocks.Manager{}
	historyStore := &hmocks.Manager{}

	c, err := NewCoreController(context.Background())
	assert.NotNil(t, c)
	assert.NoError(t, err)

	// inject mock manager for test
	c.hManager = historyStore
	c.iManager = instanceStore
	suite.Run(t, &preheatSuite{
		controller:    c,
		instanceStore: instanceStore,
		historyStore:  historyStore,
	})
}

func TestNewCoreController(t *testing.T) {
	c, err := NewCoreController(context.Background())
	assert.NotNil(t, c)
	assert.NoError(t, err)
}

func (s *preheatSuite) SetupSuite() {
	config.Init()

	s.instanceStore.On("List", mock.Anything).Return(1, []*models.Metadata{
		{
			ID:       1,
			Provider: "dragonfly",
			Endpoint: "http://localhost",
			Status:   provider.DriverStatusHealthy,
			Enabled:  true,
		},
	}, nil)
	s.instanceStore.On("Save", mock.Anything).Return(int64(1), nil)
	s.instanceStore.On("Delete", int64(1)).Return(nil)
	s.instanceStore.On("Delete", int64(0)).Return(errors.New("not found"))
	s.instanceStore.On("Get", int64(1)).Return(&models.Metadata{
		ID:       1,
		Endpoint: "http://localhost",
	}, nil)
	s.instanceStore.On("Get", int64(0)).Return(nil, errors.New("not found"))
	s.instanceStore.On("Update", mock.Anything).Return(nil)

	s.historyStore.On("LoadHistories", mock.Anything).Return(1, []*models.HistoryRecord{
		{
			TaskID: "t1",
		},
	}, nil)
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

func (s *preheatSuite) TestListInstances() {
	total, instances, err := s.controller.ListInstances(nil)
	s.NoError(err)
	s.Equal(1, int(total))
	s.Equal(1, len(instances))
	s.Equal(int64(1), instances[0].ID)
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
	s.Equal(int64(1), id)

	id, err = s.controller.CreateInstance(&models.Metadata{
		Endpoint: "http://foo.bar",
		Provider: "kraken",
	})
	s.NoError(err)
	s.Equal(int64(1), id)
}

func (s *preheatSuite) TestDeleteInstance() {
	err := s.controller.DeleteInstance(0)
	s.Error(err)

	err = s.controller.DeleteInstance(1)
	s.NoError(err)
}

func (s *preheatSuite) TestUpdateInstance() {
	err := s.controller.UpdateInstance(0, nil)
	s.Error(err)

	err = s.controller.UpdateInstance(1, nil)
	s.Error(err)

	err = s.controller.UpdateInstance(0, map[string]interface{}{"enabled": false})
	s.Error(err)

	err = s.controller.UpdateInstance(1, map[string]interface{}{"enabled": false})
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
	s.instanceStore.On("List", mock.Anything).Return(1, []*models.Metadata{
		{
			ID:       1,
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

	s.instanceStore.On("List", mock.Anything).Return(1, []*models.Metadata{
		{
			ID:       1,
			Provider: "kraken",
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
	total, records, err := s.controller.LoadHistoryRecords(nil)
	s.NoError(err)
	s.Equal(1, int(total))
	s.Equal(1, len(records))
}

func (s *preheatSuite) TestGetInstance() {
	instance, err := s.controller.GetInstance(0)
	s.Error(err)
	s.Nil(instance)

	instance, err = s.controller.GetInstance(1)
	s.NoError(err)
	s.NotNil(instance)
}
