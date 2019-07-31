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

package systeminfo

import (
	"os"

	"github.com/goharbor/harbor/src/core/systeminfo/imagestorage"
	"github.com/goharbor/harbor/src/core/systeminfo/imagestorage/filesystem"
	"github.com/goharbor/harbor/src/core/systeminfo/imagestorage/swift"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/pkg/errors"
)

// Init image storage driver
func Init() error {
	if imagestorage.GlobalDriver != nil {
		return nil
	}

	switch os.Getenv("IMAGE_STORE_SWIFT_DRIVER") {
	default:
		path := os.Getenv("IMAGE_STORE_PATH")
		if len(path) == 0 {
			path = "/data"
		}
		imagestorage.GlobalDriver = filesystem.NewDriver(path)
		return nil
	case swift.DriverName:
		auth, err := openstack.AuthOptionsFromEnv()
		if err != nil {
			return errors.Wrap(err, "invalid Swift credentials")
		}
		imagestorage.GlobalDriver, err = swift.NewDriver(auth, os.Getenv("OS_REGION_NAME"), os.Getenv("OS_CONTAINER_NAME"))
		return errors.Wrap(err, "cannot create Swift driver")
	}
}
