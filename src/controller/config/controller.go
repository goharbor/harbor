//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package config

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/config/metadata"
	"github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
)

var (
	// Ctl Global instance of the config controller
	Ctl = NewController()
)

// Controller define operations related to configures
type Controller interface {
	// UserConfigs get the user scope configurations
	UserConfigs(ctx context.Context) (map[string]*models.Value, error)
	// UpdateUserConfigs update the user scope configurations
	UpdateUserConfigs(ctx context.Context, conf map[string]interface{}) error
	// GetAll get all configurations, used by internal, should include the system config items
	AllConfigs(ctx context.Context) (map[string]interface{}, error)
}

type controller struct {
}

// NewController ...
func NewController() Controller {
	return &controller{}
}

func (c *controller) UserConfigs(ctx context.Context) (map[string]*models.Value, error) {
	mgr := config.GetCfgManager(ctx)
	configs := mgr.GetUserCfgs(ctx)
	return ConvertForGet(ctx, configs, false)
}

func (c *controller) AllConfigs(ctx context.Context) (map[string]interface{}, error) {
	mgr := config.GetCfgManager(ctx)
	configs := mgr.GetAll(ctx)
	return configs, nil
}

func (c *controller) UpdateUserConfigs(ctx context.Context, conf map[string]interface{}) error {
	mgr := config.GetCfgManager(ctx)
	err := mgr.Load(ctx)
	if err != nil {
		return err
	}
	isSysErr, err := c.validateCfg(ctx, conf)
	if err != nil {
		if isSysErr {
			log.Errorf("failed to validate configurations: %v", err)
			return fmt.Errorf("failed to validate configuration")
		}
		return err
	}
	if err := mgr.UpdateConfig(ctx, conf); err != nil {
		log.Errorf("failed to upload configurations: %v", err)
		return fmt.Errorf("failed to validate configuration")
	}
	return nil
}

func (c *controller) validateCfg(ctx context.Context, cfgs map[string]interface{}) (bool, error) {
	mgr := config.GetCfgManager(ctx)
	flag, err := authModeCanBeModified(ctx)
	if err != nil {
		return true, err
	}
	if !flag {
		if failedKeys := c.checkUnmodifiable(ctx, cfgs, common.AUTHMode); len(failedKeys) > 0 {
			return false, errors.BadRequestError(nil).
				WithMessage(fmt.Sprintf("the keys %v can not be modified as new users have been inserted into database", failedKeys))
		}
	}
	err = mgr.ValidateCfg(ctx, cfgs)
	if err != nil {
		return false, errors.BadRequestError(err)
	}
	return false, nil
}

func (c *controller) checkUnmodifiable(ctx context.Context, cfgs map[string]interface{}, keys ...string) (failed []string) {
	mgr := config.GetCfgManager(ctx)
	if mgr == nil || cfgs == nil || keys == nil {
		return
	}
	for _, k := range keys {
		v := mgr.Get(ctx, k).GetString()
		if nv, ok := cfgs[k]; ok {
			if v != fmt.Sprintf("%v", nv) {
				failed = append(failed, k)
			}
		}
	}
	return
}

// ScanAllPolicy is represent the json request and object for scan all policy
// Only for migrating from the legacy schedule.
type ScanAllPolicy struct {
	Type  string                 `json:"type"`
	Param map[string]interface{} `json:"parameter,omitempty"`
}

// ConvertForGet - delete sensitive attrs and add editable field to every attr
func ConvertForGet(ctx context.Context, cfg map[string]interface{}, internal bool) (map[string]*models.Value, error) {
	result := map[string]*models.Value{}

	mList := metadata.Instance().GetAll()

	for _, item := range mList {
		val, exist := cfg[item.Name]
		// skip undefined items
		if !exist {
			continue
		}

		switch item.ItemType.(type) {
		case *metadata.PasswordType:
			// remove password for external api call
			if !internal {
				delete(cfg, item.Name)
				continue
			}
		case *metadata.MapType, *metadata.StringToStringMapType:
			// convert to string for map type
			valByte, err := json.Marshal(val)
			if err != nil {
				return nil, err
			}
			val = string(valByte)
		}
		result[item.Name] = &models.Value{
			Val:      val,
			Editable: true,
		}
	}

	// default value for ScanAllPolicy
	if _, ok := result[common.ScanAllPolicy]; !ok {
		cfg[common.ScanAllPolicy] = `{"type":"none"}`
	}

	// set value for auth_mode
	flag, err := authModeCanBeModified(ctx)
	if err != nil {
		return nil, err
	}
	result[common.AUTHMode].Editable = flag

	return result, nil
}

func authModeCanBeModified(ctx context.Context) (bool, error) {
	return dao.AuthModeCanBeModified(ctx)
}
