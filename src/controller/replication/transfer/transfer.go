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

package transfer

import (
	"errors"
	"fmt"

	"github.com/goharbor/harbor/src/pkg/reg/model"
)

var (
	registry = map[string]Factory{}
)

// Factory creates a specific Transfer. The "Logger" is used
// to log the processing messages and the "StopFunc"
// can be used to check whether the task has been stopped
// during the processing progress
type Factory func(Logger, StopFunc) (Transfer, error)

// Transfer defines an interface used to transfer the source
// resource to the destination
type Transfer interface {
	Transfer(src *model.Resource, dst *model.Resource, speed int32) error
}

// Logger defines an interface for logging
type Logger interface {
	// For debuging
	Debug(v ...interface{})
	// For debuging with format
	Debugf(format string, v ...interface{})
	// For logging info
	Info(v ...interface{})
	// For logging info with format
	Infof(format string, v ...interface{})
	// For warning
	Warning(v ...interface{})
	// For warning with format
	Warningf(format string, v ...interface{})
	// For logging error
	Error(v ...interface{})
	// For logging error with format
	Errorf(format string, v ...interface{})
}

// StopFunc is a function used to check whether the transfer
// process is stopped
type StopFunc func() bool

// RegisterFactory registers one transfer factory to the registry
func RegisterFactory(name string, factory Factory) error {
	if len(name) == 0 {
		return errors.New("empty name")
	}
	if factory == nil {
		return errors.New("empty resource transfer factory")
	}
	if _, exist := registry[name]; exist {
		return fmt.Errorf("resource transfer factory for %s already exists", name)
	}

	registry[name] = factory
	return nil
}

// GetFactory gets the transfer factory by the specified name
func GetFactory(name string) (Factory, error) {
	factory, exist := registry[name]
	if !exist {
		return nil, fmt.Errorf("transfer factory for %s not found", name)
	}

	return factory, nil
}
