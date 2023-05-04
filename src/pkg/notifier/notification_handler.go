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

package notifier

import "context"

// NotificationHandler defines what operations a notification handler
// should have.
type NotificationHandler interface {
	// The name of the Handler
	Name() string

	// Handle the event when it coming.
	// value might be optional, it depends on usages.
	Handle(ctx context.Context, value interface{}) error

	// IsStateful returns whether the handler is stateful or not.
	// If handler is stateful, it will not be triggered in parallel.
	// Otherwise, the handler will be triggered concurrently if more
	// than one same handler are matched the topics.
	IsStateful() bool
}
