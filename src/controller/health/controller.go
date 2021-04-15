// Copyright 2019 Project Harbor Authors
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

package health

import (
	"context"
	"sort"
	"time"

	"github.com/docker/distribution/health"
)

var (
	timeout  = 60 * time.Second
	registry = map[string]health.Checker{}
	// Ctl is a global health controller
	Ctl = NewController()
)

// NewController returns a health controller instance
func NewController() Controller {
	return &controller{}
}

// Controller defines the health related operations
type Controller interface {
	GetHealth(ctx context.Context) *OverallHealthStatus
}

type controller struct{}

func (c *controller) GetHealth(ctx context.Context) *OverallHealthStatus {
	var isHealthy healthy = true
	components := []*ComponentHealthStatus{}
	ch := make(chan *ComponentHealthStatus, len(registry))
	for name, checker := range registry {
		go check(name, checker, timeout, ch)
	}
	for i := 0; i < len(registry); i++ {
		componentStatus := <-ch
		if len(componentStatus.Error) != 0 {
			isHealthy = false
		}
		components = append(components, componentStatus)
	}

	sort.Slice(components, func(i, j int) bool { return components[i].Name < components[j].Name })

	return &OverallHealthStatus{
		Status:     isHealthy.String(),
		Components: components,
	}
}

func check(name string, checker health.Checker,
	timeout time.Duration, c chan *ComponentHealthStatus) {
	statusChan := make(chan *ComponentHealthStatus)
	go func() {
		err := checker.Check()
		var healthy healthy = err == nil
		status := &ComponentHealthStatus{
			Name:   name,
			Status: healthy.String(),
		}
		if !healthy {
			status.Error = err.Error()
		}
		statusChan <- status
	}()

	select {
	case status := <-statusChan:
		c <- status
	case <-time.After(timeout):
		var healthy healthy = false
		c <- &ComponentHealthStatus{
			Name:   name,
			Status: healthy.String(),
			Error:  "failed to check the health status: timeout",
		}
	}
}
