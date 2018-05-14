// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
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

package source

import (
	"testing"

	"github.com/vmware/harbor/src/common/utils/test"
	"github.com/vmware/harbor/src/replication"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/replication/models"
)

func TestInitOfLabelFilter(t *testing.T) {
	filter := NewLabelFilter(1)
	assert.Nil(t, filter.Init())
}

func TestGetConvertorOfLabelFilter(t *testing.T) {
	filter := NewLabelFilter(1)
	assert.Nil(t, filter.GetConvertor())
}

func TestDoFilterOfLabelFilter(t *testing.T) {
	test.InitDatabaseFromEnv()
	filter := NewLabelFilter(1)
	items := []models.FilterItem{
		models.FilterItem{
			Kind:  replication.FilterItemKindTag,
			Value: "library/hello-world:latest",
		},
	}
	result := filter.DoFilter(items)
	assert.Equal(t, 0, len(result))
}
