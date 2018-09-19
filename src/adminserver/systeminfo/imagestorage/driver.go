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

package imagestorage

// GlobalDriver is a global image storage driver
var GlobalDriver Driver

// Capacity holds information about capacity of image storage
type Capacity struct {
	// total size(byte)
	Total uint64 `json:"total"`
	// available size(byte)
	Free uint64 `json:"free"`
}

// Driver defines methods that an image storage driver must implement
type Driver interface {
	// Name returns a human-readable name of the driver
	Name() string
	// Cap returns the capacity of the image storage
	Cap() (*Capacity, error)
}
