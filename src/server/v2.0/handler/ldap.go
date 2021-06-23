package handler

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/ldap"
	cfgModels "github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/errors"
	ldapModel "github.com/goharbor/harbor/src/pkg/ldap/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/ldap"
)

type ldapAPI struct {
	BaseAPI
	ctl ldap.Controller
}

func newLdapAPI() *ldapAPI {
	return &ldapAPI{ctl: ldap.Ctl}
}

func (l *ldapAPI) PingLdap(ctx context.Context, params operation.PingLdapParams) middleware.Responder {
	if err := l.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceConfiguration); err != nil {
		return l.SendError(ctx, err)
	}
	basicCfg := cfgModels.LdapConf{
		URL:            params.Ldapconf.LdapURL,
		BaseDn:         params.Ldapconf.LdapBaseDn,
		SearchDn:       params.Ldapconf.LdapSearchDn,
		Filter:         params.Ldapconf.LdapFilter,
		SearchPassword: params.Ldapconf.LdapSearchPassword,
		UID:            params.Ldapconf.LdapUID,
		Scope:          int(params.Ldapconf.LdapScope),
		VerifyCert:     params.Ldapconf.LdapVerifyCert,
	}
	payload := &models.LdapPingResult{}
	suc, err := l.ctl.Ping(ctx, basicCfg)
	payload.Success = suc
	if err != nil {
		payload.Message = fmt.Sprintf("error: %v", err)
	}
	return operation.NewPingLdapOK().WithPayload(payload)
}

func (l *ldapAPI) SearchLdapUser(ctx context.Context, params operation.SearchLdapUserParams) middleware.Responder {
	if err := l.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceLdapUser); err != nil {
		return l.SendError(ctx, err)
	}
	var username string
	if params.Username != nil {
		username = *params.Username
	}
	ldapUsers, err := l.ctl.SearchUser(ctx, username)
	if err != nil {
		return l.SendError(ctx, err)
	}
	return operation.NewSearchLdapUserOK().WithPayload(toLdapUsersResp(ldapUsers))
}

func toLdapUsersResp(users []ldapModel.User) []*models.LdapUser {
	result := make([]*models.LdapUser, 0)
	for _, u := range users {
		user := &models.LdapUser{
			Email:    u.Email,
			Realname: u.Realname,
			Username: u.Username,
		}
		result = append(result, user)
	}
	return result
}

func (l *ldapAPI) ImportLdapUser(ctx context.Context, params operation.ImportLdapUserParams) middleware.Responder {
	if err := l.RequireSystemAccess(ctx, rbac.ActionCreate, rbac.ResourceLdapUser); err != nil {
		return l.SendError(ctx, err)
	}
	failedList, err := l.ctl.ImportUser(ctx, params.UIDList.LdapUIDList)
	if err != nil {
		return l.SendError(ctx, err)
	}
	if len(failedList) == 0 {
		return operation.NewImportLdapUserOK()
	}
	return operation.NewImportLdapUserNotFound().WithPayload(toFailedListResp(failedList))
}

func toFailedListResp(users []ldapModel.FailedImportUser) []*models.LdapFailedImportUser {
	result := make([]*models.LdapFailedImportUser, 0)
	for _, u := range users {
		failed := &models.LdapFailedImportUser{
			UID:   u.UID,
			Error: u.Error,
		}
		result = append(result, failed)
	}
	return result
}

func (l *ldapAPI) SearchLdapGroup(ctx context.Context, params operation.SearchLdapGroupParams) middleware.Responder {
	if err := l.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceLdapUser); err != nil {
		return l.SendError(ctx, err)
	}
	var groupName, groupDN string
	if params.Groupname != nil && len(*params.Groupname) > 0 {
		groupName = *params.Groupname

	}
	if params.Groupdn != nil {
		groupDN = *params.Groupdn
	}
	ug, err := l.ctl.SearchGroup(ctx, groupName, groupDN)
	if err != nil {
		return l.SendError(ctx, err)
	}
	if len(ug) == 0 {
		return l.SendError(ctx, errors.NotFoundError(fmt.Errorf("group name:%v, group DN:%v", groupName, groupDN)))
	}
	return operation.NewSearchLdapGroupOK().WithPayload(toUserGroupResp(ug))
}

func toUserGroupResp(ugs []ldapModel.Group) []*models.UserGroup {
	result := make([]*models.UserGroup, 0)
	for _, g := range ugs {
		ug := &models.UserGroup{
			GroupName:   g.Name,
			LdapGroupDn: g.Dn,
		}
		result = append(result, ug)
	}
	return result
}
