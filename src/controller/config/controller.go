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

package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/config/metadata"
	"github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/audit"
	"github.com/goharbor/harbor/src/pkg/user"
)

const (
	configOverwriteJSON = "CONFIG_OVERWRITE_JSON"
)

var (
	// Ctl Global instance of the config controller
	Ctl            = NewController()
	readOnlyForAll = false
)

// Controller define operations related to configures
type Controller interface {
	// UserConfigs get the user scope configurations
	UserConfigs(ctx context.Context) (map[string]*models.Value, error)
	// UpdateUserConfigs update the user scope configurations
	UpdateUserConfigs(ctx context.Context, conf map[string]interface{}) error
	// AllConfigs get all configurations, used by internal, should include the system config items
	AllConfigs(ctx context.Context) (map[string]interface{}, error)
	// ConvertForGet - delete sensitive attrs and add editable field to every attr
	ConvertForGet(ctx context.Context, cfg map[string]interface{}, internal bool) (map[string]*models.Value, error)
	// OverwriteConfig overwrite config in the database and set all configure read only when CONFIG_OVERWRITE_JSON is provided
	OverwriteConfig(ctx context.Context) error
}

type controller struct {
	userManager user.Manager
}

// NewController ...
func NewController() Controller {
	return &controller{userManager: user.Mgr}
}

func (c *controller) UserConfigs(ctx context.Context) (map[string]*models.Value, error) {
	mgr := config.GetCfgManager(ctx)
	configs := mgr.GetUserCfgs(ctx)
	return c.ConvertForGet(ctx, configs, false)
}

func (c *controller) AllConfigs(ctx context.Context) (map[string]interface{}, error) {
	mgr := config.GetCfgManager(ctx)
	configs := mgr.GetAll(ctx)
	return configs, nil
}

func (c *controller) UpdateUserConfigs(ctx context.Context, conf map[string]interface{}) error {
	if readOnlyForAll {
		return errors.ForbiddenError(nil).WithMessage("current config is init by env variable: CONFIG_OVERWRITE_JSON, it cannot be updated")
	}
	mgr := config.GetCfgManager(ctx)
	err := mgr.Load(ctx)
	if err != nil {
		return err
	}
	err = c.validateCfg(ctx, conf)
	if err != nil {
		return err
	}
	if err := mgr.UpdateConfig(ctx, conf); err != nil {
		log.Errorf("failed to upload configurations: %v", err)
		return fmt.Errorf("failed to validate configuration")
	}
	// update the audit logger to point to the new endpoint
	return c.updateLogEndpoint(ctx, conf)
}

func (c *controller) updateLogEndpoint(ctx context.Context, cfgs map[string]interface{}) error {
	// check if the audit log forward endpoint updated
	if _, ok := cfgs[common.AuditLogForwardEndpoint]; ok {
		auditEP := config.AuditLogForwardEndpoint(ctx)
		if len(auditEP) == 0 {
			return nil
		}
		if !audit.CheckEndpointActive(auditEP) {
			return errors.BadRequestError(fmt.Errorf("could not connect to the audit endpoint: %v", auditEP))
		}
		audit.LogMgr.Init(ctx, auditEP)
	}
	return nil
}

func (c *controller) validateCfg(ctx context.Context, cfgs map[string]interface{}) error {
	mgr := config.GetCfgManager(ctx)

	// check if auth can be modified
	if nv, ok := cfgs[common.AUTHMode]; ok {
		if nv.(string) != mgr.Get(ctx, common.AUTHMode).GetString() {
			canBeModified, err := c.authModeCanBeModified(ctx)
			if err != nil {
				return err
			}
			if !canBeModified {
				return errors.BadRequestError(nil).
					WithMessage("the auth mode cannot be modified as new users have been inserted into database")
			}
		}
	}

	err := mgr.ValidateCfg(ctx, cfgs)
	if err != nil {
		return errors.BadRequestError(err)
	}

	// verify the skip audit log related cfgs
	if err = verifySkipAuditLogCfg(ctx, cfgs, mgr); err != nil {
		return err
	}
	// verify the value length related cfgs
	if err = verifyValueLengthCfg(ctx, cfgs); err != nil {
		return err
	}

	return nil
}

func verifySkipAuditLogCfg(ctx context.Context, cfgs map[string]interface{}, mgr config.Manager) error {
	updated := false
	endPoint := mgr.Get(ctx, common.AuditLogForwardEndpoint).GetString()
	skipAuditDB := mgr.Get(ctx, common.SkipAuditLogDatabase).GetBool()

	if skip, exist := cfgs[common.SkipAuditLogDatabase]; exist {
		skipAuditDB = skip.(bool)
		updated = true
	}
	if endpoint, exist := cfgs[common.AuditLogForwardEndpoint]; exist {
		endPoint = endpoint.(string)
		updated = true
	}

	if updated {
		if skipAuditDB && len(endPoint) == 0 {
			return errors.BadRequestError(errors.New("audit log forward endpoint should be configured before enable skip audit log in database"))
		}
	}
	return nil
}

// verifyValueLengthCfg verifies the cfgs which need to check the value max length to align with frontend.
func verifyValueLengthCfg(_ context.Context, cfgs map[string]interface{}) error {
	maxValue := maxValueLimitedByLength(common.UIMaxLengthLimitedOfNumber)
	validateCfgs := []string{
		common.TokenExpiration,
		common.RobotTokenDuration,
		common.SessionTimeout,
	}

	for _, c := range validateCfgs {
		if v, exist := cfgs[c]; exist {
			// the cfgs is unmarshal from json string, the number type will be float64
			if vf, ok := v.(float64); ok {
				if vf <= 0 {
					return errors.BadRequestError(nil).WithMessagef("the %s value must be positive", c)
				}

				if int64(vf) > maxValue {
					return errors.BadRequestError(nil).WithMessagef("the %s value is over the limit value: %d", c, maxValue)
				}
			}
		}
	}

	return nil
}

// maxValueLimitedByLength returns the max value can be equaled limited by the fixed length.
func maxValueLimitedByLength(length int) int64 {
	// return -1 if length is negative
	if length <= 0 {
		return -1
	}

	// the sum value
	var value int64
	// the times for multiple, should *10 for every time
	times := 1
	for i := 0; i < length; i++ {
		value = value + int64(9*times)
		times = times * 10
	}

	return value
}

// ScanAllPolicy is represent the json request and object for scan all policy
// Only for migrating from the legacy schedule.
type ScanAllPolicy struct {
	Type  string                 `json:"type"`
	Param map[string]interface{} `json:"parameter,omitempty"`
}

func (c *controller) ConvertForGet(ctx context.Context, cfg map[string]interface{}, internal bool) (map[string]*models.Value, error) {
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
			Editable: !readOnlyForAll,
		}
	}

	// default value for ScanAllPolicy
	if _, ok := result[common.ScanAllPolicy]; !ok {
		cfg[common.ScanAllPolicy] = `{"type":"none"}`
	}

	// set value for auth_mode
	canBeModified, err := c.authModeCanBeModified(ctx)
	if err != nil {
		return nil, err
	}
	result[common.AUTHMode].Editable = canBeModified && !readOnlyForAll

	return result, nil
}

func (c *controller) OverwriteConfig(ctx context.Context) error {
	cfgMap := map[string]interface{}{}
	if v, ok := os.LookupEnv(configOverwriteJSON); ok {
		err := json.Unmarshal([]byte(v), &cfgMap)
		if err != nil {
			return err
		}
		err = c.UpdateUserConfigs(ctx, cfgMap)
		if err != nil {
			return err
		}
		readOnlyForAll = true
	}
	return nil
}

func (c *controller) authModeCanBeModified(ctx context.Context) (bool, error) {
	cnt, err := c.userManager.Count(ctx, &q.Query{})
	if err != nil {
		return false, err
	}
	return cnt == 0, nil
}
