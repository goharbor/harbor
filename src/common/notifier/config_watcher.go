package notifier

import (
	"errors"
	"reflect"

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils"
)

//WatchConfigChanges is used to watch the configuration changes.
func WatchConfigChanges(cfg map[string]interface{}) error {
	if cfg == nil {
		return errors.New("Empty configurations")
	}

	//Currently only watch the scan all policy change.
	if v, ok := cfg[ScanAllPolicyTopic]; ok {
		if reflect.TypeOf(v).Kind() == reflect.Map {
			policyCfg := &models.ScanAllPolicy{}
			vMap := v.(map[string]interface{})
			//Reset filed name.
			if pv, yes := vMap["parameter"]; yes {
				vMap["Parm"] = pv
				delete(vMap, "parameter")
			}
			if err := utils.ConvertMapToStruct(policyCfg, vMap); err != nil {
				return err
			}

			policyNotification := ScanPolicyNotification{
				Type:      policyCfg.Type,
				DailyTime: 0,
			}

			if t, yes := policyCfg.Parm["daily_time"]; yes {
				if reflect.TypeOf(t).Kind() == reflect.Int {
					policyNotification.DailyTime = (int64)(t.(int))
				}
			}

			return Publish(ScanAllPolicyTopic, policyNotification)
		}
	}

	return nil
}
