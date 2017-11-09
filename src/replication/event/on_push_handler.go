package event

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/vmware/harbor/src/replication/core"
	"github.com/vmware/harbor/src/replication/models"
)

//OnPushHandler implements the notification handler interface to handle image on push event.
type OnPushHandler struct{}

//OnPushNotification contains the data required by this handler
type OnPushNotification struct {
	//The ID of the project where the being pushed images are located
	ProjectID int
}

//Handle implements the same method of notification handler interface
func (oph *OnPushHandler) Handle(value interface{}) error {
	if value == nil {
		return errors.New("OnPushHandler can not handle nil value")
	}

	vType := reflect.TypeOf(value)
	if vType.Kind() != reflect.Struct || vType.String() != "event.OnPushNotification" {
		return fmt.Errorf("Mismatch value type of OnPushHandler, expect %s but got %s", "event.OnPushNotification", vType.String())
	}

	notification := value.(OnDeletionNotification)
	//TODO:Call projectManager to get the projectID
	fmt.Println(notification.ProjectName)
	query := models.QueryParameter{
		ProjectID: 0,
	}

	policies := core.DefaultController.GetPolicies(query)
	if policies != nil && len(policies) > 0 {
		for _, p := range policies {
			if err := core.DefaultController.Replicate(p.ID); err != nil {
				//TODO:Log error
				fmt.Println(err.Error())
			}
		}
	}

	return nil
}

//IsStateful implements the same method of notification handler interface
func (oph *OnPushHandler) IsStateful() bool {
	//Statless
	return false
}
