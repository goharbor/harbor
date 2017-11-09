package event

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/vmware/harbor/src/replication/core"
	"github.com/vmware/harbor/src/replication/models"
)

//OnDeletionHandler implements the notification handler interface to handle image on push event.
type OnDeletionHandler struct{}

//OnDeletionNotification contains the data required by this handler
type OnDeletionNotification struct {
	//The name of the project where the being pushed images are located
	ProjectName string
}

//Handle implements the same method of notification handler interface
func (oph *OnDeletionHandler) Handle(value interface{}) error {
	if value == nil {
		return errors.New("OnDeletionHandler can not handle nil value")
	}

	vType := reflect.TypeOf(value)
	if vType.Kind() != reflect.Struct || vType.String() != "event.OnDeletionNotification" {
		return fmt.Errorf("Mismatch value type of OnDeletionHandler, expect %s but got %s", "event.OnDeletionNotification", vType.String())
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
			//Error accumulated and then return?
			if err := core.DefaultController.Replicate(p.ID); err != nil {
				//TODO:Log error
				fmt.Println(err.Error())
			}
		}
	}

	return nil
}

//IsStateful implements the same method of notification handler interface
func (oph *OnDeletionHandler) IsStateful() bool {
	//Statless
	return false
}
