package adapter

import (
	"fmt"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/acceleration/model"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

var registry = map[string]Factory{}

// Factory creates a specific Adapter according to the params
type Factory interface {
	Create(service *model.AccelerationService) (Adapter, error)
}

// Adapter interface defines the capabilities of AccelerationService
type Adapter interface {
	// Convert ...
	Convert(art *artifact.Artifact, tag string) error
	// HealthCheck checks health status of registry
	HealthCheck() (string, error)
}

// RegisterFactory registers one adapter factory to the registry
func RegisterFactory(t string, factory Factory) error {
	if len(t) == 0 {
		return errors.New("invalid registry type")
	}
	if factory == nil {
		return errors.New("empty adapter factory")
	}

	if _, exist := registry[t]; exist {
		return fmt.Errorf("adapter factory for %s already exists", t)
	}
	registry[t] = factory
	return nil
}

// GetFactory gets the adapter factory by the specified name
func GetFactory(t string) (Factory, error) {
	factory, exist := registry[t]
	if !exist {
		return nil, fmt.Errorf("adapter factory for %s not found", t)
	}
	return factory, nil
}
