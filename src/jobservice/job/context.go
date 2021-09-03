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

package job

import (
	"context"

	"github.com/goharbor/harbor/src/jobservice/logger"
)

// Context is combination of BaseContext and other job specified resources.
// Context will be the real execution context for one job.
type Context interface {
	// Build the context based on the parent context
	//
	// A new job context will be generated based on the current context
	// for the provided job.
	//
	// Returns:
	// new Context based on the parent one
	// error if meet any problems
	Build(tracker Tracker) (Context, error)

	// Get property from the context
	//
	// prop string : key of the context property
	//
	// Returns:
	//  The data of the specified context property if have
	//  bool to indicate if the property existing
	Get(prop string) (interface{}, bool)

	// SystemContext returns the system context
	//
	// Returns:
	//  context.Context
	SystemContext() context.Context

	// Checkin is bridge func for reporting detailed status
	//
	// status string : detailed status
	//
	// Returns:
	//  error if meet any problems
	Checkin(status string) error

	// OPCommand return the control operational command like stop if have
	//
	// Returns:
	//  op command if have
	//  flag to indicate if have command
	OPCommand() (OPCommand, bool)

	// GetLogger returns the logger
	GetLogger() logger.Interface

	// Tracker of job.
	Tracker() Tracker
}

// ContextInitializer is a func to initialize the concrete job context
type ContextInitializer func(ctx context.Context) (Context, error)
