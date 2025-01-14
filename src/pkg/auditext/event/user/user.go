package user

import (
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/auditext/event"
	pkgUser "github.com/goharbor/harbor/src/pkg/user"
)

func init() {
	var userResolver = &UserEventResolver{
		EventResolver: event.EventResolver{
			BaseURLPattern:      "/api/v2.0/users",
			ResourceType:        rbac.ResourceUser.String(),
			SucceedCodes:        []int{http.StatusCreated, http.StatusOK},
			SensitiveAttributes: []string{"password"},
			HasResourceName:     true,
			IDToNameFunc:        UserIDToName,
		},
	}
	commonevent.RegisterResolver(`/api/v2.0/users$`, userResolver)
	commonevent.RegisterResolver(`^/api/v2.0/users/\d+/password$`, userResolver)
	commonevent.RegisterResolver(`^/api/v2.0/users/\d+/sysadmin$`, userResolver)
	commonevent.RegisterResolver(`^/api/v2.0/users/\d+$`, userResolver)
}

type UserEventResolver struct {
	event.EventResolver
}

func UserIDToName(userID string) string {
	id, err := strconv.ParseInt(userID, 10, 32)
	if err != nil {
		log.Errorf("failed to parse userID: %v to int", userID)
		return ""
	}
	// use different context to so that the user is visible before the transaction is committed
	user, err := pkgUser.Mgr.Get(orm.Context(), int(id))
	if err != nil {
		log.Errorf("failed to parse userID: %v to int, err %v", userID, err)
		return ""
	}
	return user.Username
}
