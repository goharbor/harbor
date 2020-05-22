package registryctl

import (
	"github.com/goharbor/harbor/src/registryctl/api"
	"github.com/stretchr/testify/mock"
)

type Mockclient struct {
	mock.Mock
}

// Health ...
func (c *Mockclient) Health() error {
	return nil
}

// StartGC ...
func (c *Mockclient) StartGC() (*api.GCResult, error) {
	result := &api.GCResult{
		Status: true,
		Msg:    "this is a mock client",
	}
	return result, nil
}
