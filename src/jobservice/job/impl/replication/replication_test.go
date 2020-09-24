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

package replication

import (
	"testing"

	"github.com/goharbor/harbor/src/jobservice/job/impl"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/transfer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseParam(t *testing.T) {
	params := map[string]interface{}{}
	// not exist param
	err := parseParam(params, "not_exist_param", nil)
	assert.NotNil(t, err)
	// the param is not string
	params["num"] = 1
	err = parseParam(params, "num", nil)
	assert.NotNil(t, err)
	// not a valid json struct
	type person struct {
		Name string
	}
	params["person"] = `"name": "tom"`
	p := &person{}
	err = parseParam(params, "person", p)
	assert.NotNil(t, err)
	// pass
	params["person"] = `{"name": "tom"}`
	err = parseParam(params, "person", p)
	assert.Nil(t, err)
	assert.Equal(t, "tom", p.Name)
}

func TestParseParams(t *testing.T) {
	params := map[string]interface{}{
		"src_resource": `{"type":"chart"}`,
		"dst_resource": `{"type":"chart"}`,
	}
	res, dst, err := parseParams(params)
	require.Nil(t, err)
	assert.Equal(t, "chart", string(res.Type))
	assert.Equal(t, "chart", string(dst.Type))
}

func TestMaxFails(t *testing.T) {
	rep := &Replication{}
	assert.Equal(t, uint(3), rep.MaxFails())
}

func TestShouldRetry(t *testing.T) {
	rep := &Replication{}
	assert.True(t, rep.ShouldRetry())
}

func TestValidate(t *testing.T) {
	rep := &Replication{}
	assert.Nil(t, rep.Validate(nil))
}

var transferred = false

var fakedTransferFactory = func(transfer.Logger, transfer.StopFunc) (transfer.Transfer, error) {
	return &fakedTransfer{}, nil
}

type fakedTransfer struct{}

func (f *fakedTransfer) Transfer(src *model.Resource, dst *model.Resource) error {
	transferred = true
	return nil
}

func TestRun(t *testing.T) {
	err := transfer.RegisterFactory("art", fakedTransferFactory)
	require.Nil(t, err)
	params := map[string]interface{}{
		"src_resource": `{"type":"art"}`,
		"dst_resource": `{}`,
	}
	rep := &Replication{}
	require.Nil(t, rep.Run(&impl.Context{}, params))
	assert.True(t, transferred)
}
