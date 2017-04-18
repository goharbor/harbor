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

package filesystem

import (
	"syscall"

	storage "github.com/vmware/harbor/src/adminserver/systeminfo/imagestorage"
)

const (
	driverName = "filesystem"
)

type driver struct {
	path string
}

// NewDriver returns an instance of filesystem driver
func NewDriver(path string) storage.Driver {
	return &driver{
		path: path,
	}
}

// Name returns a human-readable name of the fielsystem driver
func (d *driver) Name() string {
	return driverName
}

// Cap returns the capacity of the filesystem storage
func (d *driver) Cap() (*storage.Capacity, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(d.path, &stat)
	if err != nil {
		return nil, err
	}

	return &storage.Capacity{
		Total: stat.Blocks * uint64(stat.Bsize),
		Free:  stat.Bavail * uint64(stat.Bsize),
	}, nil
}
