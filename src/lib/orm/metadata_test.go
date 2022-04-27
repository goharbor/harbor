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

package orm

import (
	"context"
	"testing"

	"github.com/beego/beego/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type foo struct {
	Field1 string `orm:"-"`
	Field2 string `orm:"column(customized_field2)" filter:"false"`
	Field3 string `sort:"false"`
	Field4 string `sort:"default:desc"`
}

func (f *foo) FilterByField5(context.Context, orm.QuerySeter, string, interface{}) orm.QuerySeter {
	return nil
}

func (f *foo) OtherFunc() {}

type bar struct {
	Field1 string
	Field2 string
}

func (b *bar) GetDefaultSorts() []*q.Sort {
	return []*q.Sort{
		{
			Key:  "Field1",
			DESC: true,
		},
		{
			Key:  "Field2",
			DESC: false,
		},
	}
}

func TestParseQueryObject(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	metadata := parseModel(&foo{})
	require.NotNil(metadata)
	require.Len(metadata.Keys, 8)

	key, exist := metadata.Keys["Field2"]
	require.True(exist)
	assert.Equal("Field2", key.Name)
	assert.False(key.Filterable)
	assert.True(key.Sortable)

	key, exist = metadata.Keys["customized_field2"]
	require.True(exist)
	assert.Equal("customized_field2", key.Name)
	assert.False(key.Filterable)
	assert.True(key.Sortable)

	key, exist = metadata.Keys["Field3"]
	require.True(exist)
	assert.Equal("Field3", key.Name)
	assert.True(key.Filterable)
	assert.False(key.Sortable)

	key, exist = metadata.Keys["field3"]
	require.True(exist)
	assert.Equal("field3", key.Name)
	assert.True(key.Filterable)
	assert.False(key.Sortable)

	key, exist = metadata.Keys["Field4"]
	require.True(exist)
	assert.Equal("Field4", key.Name)
	assert.True(key.Filterable)
	assert.True(key.Sortable)

	key, exist = metadata.Keys["field4"]
	require.True(exist)
	assert.Equal("field4", key.Name)
	assert.True(key.Filterable)
	assert.True(key.Sortable)

	key, exist = metadata.Keys["Field5"]
	require.True(exist)
	assert.Equal("Field5", key.Name)
	assert.True(key.Filterable)
	assert.False(key.Sortable)

	key, exist = metadata.Keys["field5"]
	require.True(exist)
	assert.Equal("field5", key.Name)
	assert.True(key.Filterable)
	assert.False(key.Sortable)

	require.Len(metadata.DefaultSorts, 1)
	assert.Equal("Field4", metadata.DefaultSorts[0].Key)
	assert.True(metadata.DefaultSorts[0].DESC)

	metadata = parseModel(&bar{})
	require.NotNil(metadata)
	require.Len(metadata.Keys, 4)
	require.Len(metadata.DefaultSorts, 2)
	assert.Equal("Field1", metadata.DefaultSorts[0].Key)
	assert.True(metadata.DefaultSorts[0].DESC)
	assert.Equal("Field2", metadata.DefaultSorts[1].Key)
	assert.False(metadata.DefaultSorts[1].DESC)
}

func Test_snakeCase(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"ProjectID", args{"ProjectID"}, "project_id"},
		{"project_id", args{"project_id"}, "project_id"},
		{"RepositoryName", args{"RepositoryName"}, "repository_name"},
		{"repository_name", args{"repository_name"}, "repository_name"},
		{"ProfileURL", args{"ProfileURL"}, "profile_url"},
		{"City", args{"City"}, "city"},
		{"Address1", args{"Address1"}, "address1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := snakeCase(tt.args.str); got != tt.want {
				t.Errorf("snakeCase() = %v, want %v", got, tt.want)
			}
		})
	}
}
