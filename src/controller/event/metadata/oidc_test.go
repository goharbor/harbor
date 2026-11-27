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

package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	event2 "github.com/goharbor/harbor/src/controller/event"
	evtModel "github.com/goharbor/harbor/src/controller/event/model"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

func TestResolveOfOIDCOnboardEventMetadata(t *testing.T) {
	e := &event.Event{}
	require.NoError(t, (&OIDCOnboardEventMetadata{Username: "ci-user"}).Resolve(e))
	assert.Equal(t, event2.TopicCommonEvent, e.Topic)
	data, ok := e.Data.(*evtModel.CommonEvent)
	require.True(t, ok)
	assert.Equal(t, "ci-user", data.Operator)
	assert.Equal(t, "user", data.ResourceType)
	assert.Equal(t, "ci-user", data.ResourceName)
	assert.Equal(t, "login", data.Operation)
	assert.Equal(t, "login", data.OperationDescription)
	assert.True(t, data.IsSuccessful)
}
