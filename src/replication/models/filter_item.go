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

package models

//FilterItem is the general data model represents the filtering resources which are used as input and output for the filters.
type FilterItem struct {

	//The kind of the filtering resources. Support 'project','repository' and 'tag' etc.
	Kind string `json:"kind"`

	//The key value of resource which can be used to filter out the resource matched with specified pattern.
	//E.g:
	//kind == 'project', value will be project name;
	//kind == 'repository', value will be repository name
	//kind == 'tag', value will be tag name.
	Value string `json:"value"`

	Operation string `json:"operation"`

	//Extension placeholder.
	//To append more additional information if required by the filter.
	Metadata map[string]interface{} `json:"metadata"`
}
