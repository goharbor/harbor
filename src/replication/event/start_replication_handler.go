package event

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/vmware/harbor/src/replication/core"
)

//StartReplicationHandler implements the notification handler interface to handle start replication requests.
type StartReplicationHandler struct{}

//StartReplicationNotification contains data required by this handler
type StartReplicationNotification struct {
	//ID of the policy
	PolicyID int
}

//Handle implements the same method of notification handler interface
func (srh *StartReplicationHandler) Handle(value interface{}) error {
	if value == nil {
		return errors.New("StartReplicationHandler can not handle nil value")
	}

	vType := reflect.TypeOf(value)
	if vType.Kind() != reflect.Struct || vType.String() != "core.StartReplicationNotification" {
		return fmt.Errorf("Mismatch value type of StartReplicationHandler, expect %s but got %s", "core.StartReplicationNotification", vType.String())
	}

	notification := value.(StartReplicationNotification)
	if notification.PolicyID <= 0 {
		return errors.New("Invalid policy")
	}

	//Start replication
	//TODO:
	return core.DefaultController.Replicate(notification.PolicyID)
}

//IsStateful implements the same method of notification handler interface
func (srh *StartReplicationHandler) IsStateful() bool {
	//Stateless
	return false
}
