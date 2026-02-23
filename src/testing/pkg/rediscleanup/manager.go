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

package rediscleanup

import (
	"context"

	"github.com/goharbor/harbor/src/pkg/rediscleanup"
	"github.com/stretchr/testify/mock"
)

// Manager is a mock implementation of rediscleanup.Manager
type Manager struct {
	mock.Mock
}

// CleanupInvalidBlobSizeKeys is a mock implementation
func (m *Manager) CleanupInvalidBlobSizeKeys(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Ensure Manager implements rediscleanup.Manager interface
var _ rediscleanup.Manager = (*Manager)(nil)
