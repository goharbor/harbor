package ldap

import (
	"context"
	"fmt"

	goldap "github.com/go-ldap/ldap/v3"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/auth"
	cfgModels "github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/ldap/model"
)

var (
	// Mgr default quota manager
	Mgr = New()
)

// Manager is used for ldap management
type Manager interface {
	// Ping ldap test
	Ping(ctx context.Context, cfg cfgModels.LdapConf) (bool, error)
	SearchUser(ctx context.Context, sess *Session, username string) ([]model.User, error)
	ImportUser(ctx context.Context, sess *Session, ldapImportUsers []string) ([]model.FailedImportUser, error)
	SearchGroup(ctx context.Context, sess *Session, groupName, groupDN string) ([]model.Group, error)
}

// New returns a default implementation of Manager
func New() Manager {
	return &manager{}
}

type manager struct {
}

func (m *manager) Ping(ctx context.Context, cfg cfgModels.LdapConf) (bool, error) {
	return TestConfig(cfg)
}

func (m *manager) SearchUser(ctx context.Context, sess *Session, username string) ([]model.User, error) {
	users := make([]model.User, 0)
	if err := sess.Open(); err != nil {
		return users, err
	}
	defer sess.Close()

	ldapUsers, err := sess.SearchUser(username)
	if err != nil {
		return users, err
	}
	for _, u := range ldapUsers {
		ldapUser := model.User{
			Username:    u.Username,
			Realname:    u.Realname,
			GroupDNList: u.GroupDNList,
			Email:       u.Email,
		}
		users = append(users, ldapUser)
	}
	return users, nil
}

func (m *manager) ImportUser(ctx context.Context, sess *Session, ldapImportUsers []string) ([]model.FailedImportUser, error) {
	failedImportUser := make([]model.FailedImportUser, 0)
	if err := sess.Open(); err != nil {
		return failedImportUser, err
	}
	defer sess.Close()

	for _, tempUID := range ldapImportUsers {
		var u model.FailedImportUser
		u.UID = tempUID
		u.Error = ""

		if u.UID == "" {
			u.Error = "empty_uid"
			failedImportUser = append(failedImportUser, u)
			continue
		}

		if u.Error != "" {
			failedImportUser = append(failedImportUser, u)
			continue
		}

		ldapUsers, err := sess.SearchUser(u.UID)
		if err != nil {
			u.UID = tempUID
			u.Error = "failed_search_user"
			failedImportUser = append(failedImportUser, u)
			log.Errorf("Invalid LDAP search request for %s, error: %v", tempUID, err)
			continue
		}

		if ldapUsers == nil || len(ldapUsers) <= 0 {
			u.UID = tempUID
			u.Error = "unknown_user"
			failedImportUser = append(failedImportUser, u)
			continue
		}

		var user models.User

		user.Username = ldapUsers[0].Username
		user.Realname = ldapUsers[0].Realname
		user.Email = ldapUsers[0].Email
		err = auth.OnBoardUser(ctx, &user)

		if err != nil || user.UserID <= 0 {
			u.UID = tempUID
			u.Error = "failed to import user: " + u.UID
			failedImportUser = append(failedImportUser, u)
			log.Errorf("Can't import user %s, error: %s", tempUID, u.Error)
		}

	}

	return failedImportUser, nil
}

func (m *manager) SearchGroup(ctx context.Context, sess *Session, groupName, groupDN string) ([]model.Group, error) {
	err := sess.Open()
	if err != nil {
		return nil, err
	}
	defer sess.Close()

	ldapGroups := make([]model.Group, 0)

	// Search LDAP group by groupName or group DN
	if len(groupName) > 0 {
		ldapGroups, err = sess.SearchGroupByName(groupName)
		if err != nil {
			return nil, err
		}
	} else if len(groupDN) > 0 {
		if _, err := goldap.ParseDN(groupDN); err != nil {
			return nil, fmt.Errorf("invalid DN: %v", err)
		}
		ldapGroups, err = sess.SearchGroupByDN(groupDN)
		if err != nil {
			return nil, err
		}
	}

	return ldapGroups, nil
}
