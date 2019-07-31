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

package swift

import (
	"strconv"

	"github.com/goharbor/harbor/src/common/utils/log"
	storage "github.com/goharbor/harbor/src/core/systeminfo/imagestorage"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/containers"
	"github.com/pkg/errors"
)

const (
	// DriverName is the unique name representing the Swift driver
	DriverName = "swift"

	// quotaMetaHeader is the metadata name to handle quota in Swift containers
	// see: https://www.swiftstack.com/docs/admin/middleware/container_quotas.html#setting-the-limit-in-bytes
	quotaMetaHeader = "Quota-Bytes"
)

type driver struct {
	client    *gophercloud.ServiceClient
	container string
}

// NewDriver returns an instance of filesystem driver
func NewDriver(auth gophercloud.AuthOptions, region, container string) (storage.Driver, error) {
	provider, err := openstack.AuthenticatedClient(auth)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get provider client")
	}

	client, err := openstack.NewObjectStorageV1(provider, gophercloud.EndpointOpts{
		Region:       region,
		Availability: gophercloud.AvailabilityPublic,
	})
	if err != nil {
		return nil, errors.Wrap(err, "cannot get Swift client")
	}

	return &driver{
		client:    client,
		container: container,
	}, nil
}

// Name returns a human-readable name of the fielsystem driver
func (d *driver) Name() string {
	return DriverName
}

// Cap returns the capacity of the filesystem storage
func (d *driver) Cap() (*storage.Capacity, error) {
	response := containers.Get(d.client, d.container, containers.GetOpts{})
	m, err := response.ExtractMetadata()
	if err != nil {
		return nil, err
	}

	if quota, ok := m[quotaMetaHeader]; ok {
		quota, err := strconv.ParseUint(quota, 10, 64)
		if err != nil {
			log.Warningf("invalid quota: %v", err)
		} else {
			free := uint64(0)

			r, err := response.Extract()
			if err != nil {
				return nil, err
			}

			if r.BytesUsed < 0 {
				return nil, errors.Errorf("invalid negative bytes count: %d", r.BytesUsed)
			}

			used := uint64(r.BytesUsed)
			if quota > used {
				free = quota - used
			}

			return &storage.Capacity{
				Total: quota,
				Free:  free,
			}, nil
		}
	}
	return &storage.Capacity{
		Total: 0,
		Free:  0,
	}, nil
}
