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

package operator

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/common/security"
	testsec "github.com/goharbor/harbor/src/testing/common/security"

	"github.com/stretchr/testify/assert"
)

func TestFromContext(t *testing.T) {
	{
		// no security context and operator context should return ""
		op := FromContext(context.Background())
		assert.Empty(t, op)
	}
	{
		// return operator from security context
		secCtx := &testsec.Context{}
		secCtx.On("GetUsername").Return("security-context-user")
		ctx := security.NewContext(context.Background(), secCtx)
		op := FromContext(ctx)
		assert.Equal(t, "security-context-user", op)
	}
	{
		// return operator from operator context
		ctx := context.WithValue(context.Background(), ContextKey{}, "operator-context-user")
		op := FromContext(ctx)
		assert.Equal(t, "operator-context-user", op)
	}
}
