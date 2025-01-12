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

package core

import "time"

type Artifact struct {
	Digest   string    `json:"digest"`
	Labels   []Label   `json:"labels"`
	Tags     []Tag     `json:"tags"`
	PullTime time.Time `json:"pull_time"`
	PushTime time.Time `json:"push_time"`
}

type Label struct {
	Name string `json:"name"`
}
type Tag struct {
	Name     string    `json:"name"`
	PullTime time.Time `json:"pull_time"`
	PushTime time.Time `json:"push_time"`
}
