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

package target

import (
	"fmt"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/ui/config"
)

// Manager defines the methods that a target manager should implement
type Manager interface {
	GetTarget(int64) (*models.RepTarget, error)
}

// DefaultManager implement the Manager interface
type DefaultManager struct{}

// NewDefaultManager returns an instance of DefaultManger
func NewDefaultManager() *DefaultManager {
	return &DefaultManager{}
}

// GetTarget ...
func (d *DefaultManager) GetTarget(id int64) (*models.RepTarget, error) {
	target, err := dao.GetRepTarget(id)
	if err != nil {
		return nil, err
	}

	if target == nil {
		return nil, fmt.Errorf("target '%d' does not exist", id)
	}

	// decrypt the password
	if len(target.Password) > 0 {
		key, err := config.SecretKey()
		if err != nil {
			return nil, err
		}
		pwd, err := utils.ReversibleDecrypt(target.Password, key)
		if err != nil {
			return nil, err
		}
		target.Password = pwd
	}
	return target, nil
}
