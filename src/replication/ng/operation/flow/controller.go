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

package flow

// Flow defines the replication flow
type Flow interface {
	// returns the count of tasks which have been scheduled and the error
	Run(interface{}) (int, error)
}

// Controller is the controller that controls the replication flows
type Controller interface {
	Start(Flow) (int, error)
}

// NewController returns an instance of the default flow controller
func NewController() Controller {
	return &controller{}
}

type controller struct{}

func (c *controller) Start(flow Flow) (int, error) {
	return flow.Run(nil)
}
