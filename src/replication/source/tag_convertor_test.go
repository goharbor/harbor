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

package source

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/replication"
	"github.com/vmware/harbor/src/replication/models"
)

func TestTagConvert(t *testing.T) {
	items := []models.FilterItem{
		models.FilterItem{
			Kind:  replication.FilterItemKindRepository,
			Value: "library/ubuntu",
		},
		models.FilterItem{
			Kind: replication.FilterItemKindProject,
		},
	}
	expected := []models.FilterItem{
		models.FilterItem{
			Kind:  replication.FilterItemKindTag,
			Value: "library/ubuntu:14.04",
		},
		models.FilterItem{
			Kind:  replication.FilterItemKindTag,
			Value: "library/ubuntu:16.04",
		},
		models.FilterItem{
			Kind: replication.FilterItemKindProject,
		},
	}

	convertor := NewTagConvertor(&fakeRegistryAdaptor{})
	assert.EqualValues(t, expected, convertor.Convert(items))
}
