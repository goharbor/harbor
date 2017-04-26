// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package filter

import (
	"net/http"
	"testing"

	"github.com/astaxie/beego/context"
	"github.com/stretchr/testify/assert"
)

func TestSecurityFilter(t *testing.T) {
	// nil request
	ctx := &context.Context{
		Request: nil,
		Input:   context.NewInput(),
	}
	SecurityFilter(ctx)
	securityContext := ctx.Input.GetData(HarborSecurityContext)
	assert.Nil(t, securityContext)

	// the pattern of request does not need security check
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/static/index.html", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", req)
	}

	ctx = &context.Context{
		Request: req,
		Input:   context.NewInput(),
	}
	SecurityFilter(ctx)
	securityContext = ctx.Input.GetData(HarborSecurityContext)
	assert.Nil(t, securityContext)

	//TODO add a case to test normal process
}
