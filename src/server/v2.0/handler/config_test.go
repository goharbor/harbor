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

package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common/security"
	cfgModels "github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/configure"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	configtesting "github.com/goharbor/harbor/src/testing/controller/config"
	"github.com/goharbor/harbor/src/testing/mock"
)

func TestGetConfigurationsReturnsErrorResponder(t *testing.T) {
	secCtx := &securitytesting.Context{}
	mock.OnAnything(secCtx, "IsSolutionUser").Return(false)
	mock.OnAnything(secCtx, "IsAuthenticated").Return(true)
	mock.OnAnything(secCtx, "Can").Return(true)

	// auth_mode is a string in the response model, so a numeric value makes
	// toResponseModel fail
	ctl := &configtesting.Controller{}
	mock.OnAnything(ctl, "UserConfigs").Return(map[string]*cfgModels.Value{
		"auth_mode": {Val: 1, Editable: true},
	}, nil)

	api := &configAPI{controller: ctl}
	ctx := security.NewContext(context.Background(), secCtx)
	responder := api.GetConfigurations(ctx, configure.GetConfigurationsParams{})

	assert.NotNil(t, responder)
	_, ok := responder.(*configure.GetConfigurationsOK)
	assert.False(t, ok, "GetConfigurations answered OK after it failed to build the response model")
}
