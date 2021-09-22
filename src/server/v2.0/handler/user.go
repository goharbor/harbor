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

package handler

import (
	"context"
	"fmt"
	commonmodels "github.com/goharbor/harbor/src/common/models"
	"regexp"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/rbac/system"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/user"
)

var userResource = system.NewNamespace().Resource(rbac.ResourceUser)

type usersAPI struct {
	BaseAPI
	ctl     user.Controller
	getAuth func(ctx context.Context) (string, error) // For testing
}

func newUsersAPI() *usersAPI {
	return &usersAPI{
		ctl:     user.Ctl,
		getAuth: config.AuthMode,
	}
}

func (u *usersAPI) SetCliSecret(ctx context.Context, params operation.SetCliSecretParams) middleware.Responder {
	uid := int(params.UserID)
	if err := u.requireForCLISecret(ctx, uid); err != nil {
		return u.SendError(ctx, err)
	}
	if err := requireValidSecret(params.Secret.Secret); err != nil {
		return u.SendError(ctx, err)
	}
	if err := u.ctl.SetCliSecret(ctx, uid, params.Secret.Secret); err != nil {
		log.G(ctx).Errorf("Failed to set CLI secret, error: %v", err)
		return u.SendError(ctx, err)
	}
	return operation.NewSetCliSecretOK()
}

func (u *usersAPI) CreateUser(ctx context.Context, params operation.CreateUserParams) middleware.Responder {
	if err := u.requireCreatable(ctx); err != nil {
		return u.SendError(ctx, err)
	}
	if err := requireValidSecret(params.UserReq.Password); err != nil {
		return u.SendError(ctx, err)
	}
	m := &commonmodels.User{
		Username: params.UserReq.Username,
		Realname: params.UserReq.Realname,
		Email:    params.UserReq.Email,
		Comment:  params.UserReq.Comment,
		Password: params.UserReq.Password,
	}
	if err := validateUserProfile(m); err != nil {
		return u.SendError(ctx, err)
	}
	uid, err := u.ctl.Create(ctx, m)
	if err != nil {
		log.G(ctx).Errorf("Failed to create user, error: %v", err)
		return u.SendError(ctx, err)
	}
	location := fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), uid)
	return operation.NewCreateUserCreated().WithLocation(location)

}

func (u *usersAPI) ListUsers(ctx context.Context, params operation.ListUsersParams) middleware.Responder {
	if err := u.RequireSystemAccess(ctx, rbac.ActionList, userResource); err != nil {
		return u.SendError(ctx, err)
	}
	query, err := u.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return u.SendError(ctx, err)
	}
	values := params.HTTPRequest.URL.Query()
	for _, k := range []string{"username", "email"} {
		if v := values.Get(k); v != "" {
			query.Keywords[k] = &q.FuzzyMatchValue{Value: v}
		}
	}
	total, err := u.ctl.Count(ctx, query)
	if err != nil {
		return u.SendError(ctx, err)
	}
	payload := make([]*models.UserResp, 0)
	if total > 0 {
		users, err := u.ctl.List(ctx, query)
		if err != nil {
			return u.SendError(ctx, err)
		}
		payload = make([]*models.UserResp, len(users))
		for i, u := range users {
			m := &model.User{
				User: u,
			}
			payload[i] = m.ToUserResp()
		}
	}
	return operation.NewListUsersOK().
		WithPayload(payload).
		WithLink(u.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithXTotalCount(total)
}

func (u *usersAPI) GetCurrentUserPermissions(ctx context.Context, params operation.GetCurrentUserPermissionsParams) middleware.Responder {
	if err := u.RequireAuthenticated(ctx); err != nil {
		return u.SendError(ctx, err)
	}
	scope := ""
	if params.Scope != nil {
		scope = *params.Scope
	}
	var policies []*types.Policy
	sctx, _ := security.FromContext(ctx)
	if ns, ok := types.NamespaceFromResource(rbac.Resource(scope)); ok {
		for _, policy := range ns.GetPolicies() {
			if sctx.Can(ctx, policy.Action, policy.Resource) {
				policies = append(policies, policy)
			}
		}
	}
	var res []*models.Permission
	relative := lib.BoolValue(params.Relative)
	for _, policy := range policies {
		var resource rbac.Resource
		// for resource `/project/1/repository` if `relative` is `true` then the resource in response will be `repository`
		if relative {
			relativeResource, err := policy.Resource.RelativeTo(rbac.Resource(scope))
			if err != nil {
				continue
			}
			resource = relativeResource
		} else {
			resource = policy.Resource
		}
		res = append(res, &models.Permission{
			Resource: resource.String(),
			Action:   policy.Action.String(),
		})
	}
	return operation.NewGetCurrentUserPermissionsOK().WithPayload(res)
}

func (u *usersAPI) DeleteUser(ctx context.Context, params operation.DeleteUserParams) middleware.Responder {
	uid := int(params.UserID)
	if err := u.requireDeletable(ctx, uid); err != nil {
		return u.SendError(ctx, err)
	}
	if err := u.ctl.Delete(ctx, uid); err != nil {
		log.G(ctx).Errorf("Failed to delete user %d, error: %v", uid, err)
		return u.SendError(ctx, err)
	}
	return operation.NewDeleteUserOK()
}

func (u *usersAPI) GetCurrentUserInfo(ctx context.Context, params operation.GetCurrentUserInfoParams) middleware.Responder {
	if err := u.RequireAuthenticated(ctx); err != nil {
		return u.SendError(ctx, err)
	}
	sctx, _ := security.FromContext(ctx)
	lsc, ok := sctx.(*local.SecurityContext)
	if !ok {
		return u.SendError(ctx, errors.PreconditionFailedError(nil).WithMessage("get current user not available for security context: %s", sctx.Name()))
	}
	resp, err := u.getUserByID(ctx, lsc.User().UserID)
	if err != nil {

		return u.SendError(ctx, err)
	}
	return operation.NewGetCurrentUserInfoOK().WithPayload(resp)
}

func (u *usersAPI) GetUser(ctx context.Context, params operation.GetUserParams) middleware.Responder {
	uid := int(params.UserID)
	if err := u.requireReadable(ctx, uid); err != nil {
		return u.SendError(ctx, err)
	}
	resp, err := u.getUserByID(ctx, uid)
	if err != nil {
		log.G(ctx).Errorf("Failed to get user info for ID %d, error: %v", uid, err)
		return u.SendError(ctx, err)
	}
	return operation.NewGetUserOK().WithPayload(resp)
}

func (u *usersAPI) getUserByID(ctx context.Context, id int) (*models.UserResp, error) {
	auth, err := u.getAuth(ctx)
	if err != nil {
		return nil, err
	}

	opt := &user.Option{
		WithOIDCInfo: auth == common.OIDCAuth && id > 1, // Super user is authenticated via DB
	}

	us, err := u.ctl.Get(ctx, id, opt)
	if err != nil {
		return nil, err
	}
	m := &model.User{
		User: us,
	}
	return m.ToUserResp(), nil
}

func (u *usersAPI) UpdateUserProfile(ctx context.Context, params operation.UpdateUserProfileParams) middleware.Responder {
	uid := int(params.UserID)
	if err := u.requireModifiable(ctx, uid); err != nil {
		return u.SendError(ctx, err)
	}
	m := &commonmodels.User{
		UserID:   uid,
		Realname: params.Profile.Realname,
		Email:    params.Profile.Email,
		Comment:  params.Profile.Comment,
	}
	if err := validateUserProfile(m); err != nil {
		return u.SendError(ctx, err)
	}
	if err := u.ctl.UpdateProfile(ctx, m); err != nil {
		log.G(ctx).Errorf("Failed to update user profile, error: %v", err)
		return u.SendError(ctx, err)
	}
	return operation.NewUpdateUserProfileOK()
}

func (u *usersAPI) SearchUsers(ctx context.Context, params operation.SearchUsersParams) middleware.Responder {
	if err := u.RequireAuthenticated(ctx); err != nil {
		return u.SendError(ctx, err)
	}
	query, err := u.BuildQuery(ctx, nil, nil, params.Page, params.PageSize)
	if err != nil {
		return u.SendError(ctx, err)
	}
	query.Keywords["username"] = &q.FuzzyMatchValue{Value: params.Username}
	total, err := u.ctl.Count(ctx, query)
	if err != nil {
		return u.SendError(ctx, err)
	}
	if total == 0 {
		return operation.NewSearchUsersOK().WithXTotalCount(0).WithPayload([]*models.UserSearchRespItem{})
	}
	l, err := u.ctl.List(ctx, query)
	if err != nil {
		return u.SendError(ctx, err)
	}
	var result []*models.UserSearchRespItem
	for _, us := range l {
		m := &model.User{User: us}
		result = append(result, m.ToSearchRespItem())
	}
	return operation.NewSearchUsersOK().
		WithXTotalCount(total).
		WithPayload(result).
		WithLink(u.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String())
}

func (u *usersAPI) UpdateUserPassword(ctx context.Context, params operation.UpdateUserPasswordParams) middleware.Responder {
	uid := int(params.UserID)
	if err := u.requireModifiable(ctx, uid); err != nil {
		return u.SendError(ctx, err)
	}
	sctx, _ := security.FromContext(ctx)
	if matchUserID(sctx, uid) {
		ok, err := u.ctl.VerifyPassword(ctx, sctx.GetUsername(), params.Password.OldPassword)
		if err != nil {
			log.G(ctx).Errorf("Failed to verify password for user: %s, error: %v", sctx.GetUsername(), err)
			return u.SendError(ctx, errors.UnknownError(nil).WithMessage("Failed to verify password"))
		}
		if !ok {
			return u.SendError(ctx, errors.ForbiddenError(nil).WithMessage("Current password is incorrect"))
		}
	}
	newPwd := params.Password.NewPassword
	if err := requireValidSecret(newPwd); err != nil {
		return u.SendError(ctx, err)
	}
	ok, err := u.ctl.VerifyPassword(ctx, sctx.GetUsername(), newPwd)
	if err != nil {
		log.G(ctx).Errorf("Failed to verify password for user: %s, error: %v", sctx.GetUsername(), err)
		return u.SendError(ctx, errors.UnknownError(nil).WithMessage("Failed to verify password"))
	}
	if ok {
		return u.SendError(ctx, errors.BadRequestError(nil).WithMessage("New password is identical to old password"))
	}
	err2 := u.ctl.UpdatePassword(ctx, uid, params.Password.NewPassword)
	if err2 != nil {
		log.G(ctx).Errorf("Failed to update password, error: %v", err)
		return u.SendError(ctx, err)
	}
	return operation.NewUpdateUserPasswordOK()
}

func (u *usersAPI) SetUserSysAdmin(ctx context.Context, params operation.SetUserSysAdminParams) middleware.Responder {
	id := int(params.UserID)
	if err := u.RequireSystemAccess(ctx, rbac.ActionUpdate, rbac.ResourceUser); err != nil {
		return u.SendError(ctx, err)
	}
	if err := u.ctl.SetSysAdmin(ctx, id, params.SysadminFlag.SysadminFlag); err != nil {
		return u.SendError(ctx, err)
	}
	return operation.NewSetUserSysAdminOK()
}

func (u *usersAPI) requireForCLISecret(ctx context.Context, id int) error {
	a, err := u.getAuth(ctx)
	if err != nil {
		log.G(ctx).Errorf("Failed to get authmode, error: %v", err)
		return err
	}
	if a != common.OIDCAuth {
		return errors.PreconditionFailedError(nil).WithMessage("unable to update CLI secret under authmode: %s", a)
	}
	sctx, ok := security.FromContext(ctx)
	if !ok || !sctx.IsAuthenticated() {
		return errors.UnauthorizedError(nil)
	}
	if !matchUserID(sctx, id) && !sctx.Can(ctx, rbac.ActionUpdate, userResource) {
		return errors.ForbiddenError(nil).WithMessage("Not authorized to update the CLI secret for user: %d", id)
	}
	return nil
}

func (u *usersAPI) requireCreatable(ctx context.Context) error {
	a, err := u.getAuth(ctx)
	if err != nil {
		log.G(ctx).Errorf("Failed to get authmode, error: %v", err)
		return err
	}
	if a != common.DBAuth {
		return errors.ForbiddenError(nil).WithMessage("creating local user is not allowed under auth mode: %s", a)
	}
	sr, err := config.SelfRegistration(ctx)
	if err != nil {
		log.G(ctx).Errorf("Failed to get self registration flag, error: %v", err)
		return err
	}
	accessErr := u.RequireSystemAccess(ctx, rbac.ActionCreate, rbac.ResourceUser)
	if !sr {
		return accessErr
	}
	if accessErr != nil && !lib.GetCarrySession(ctx) {
		return errors.ForbiddenError(nil).WithMessage("self-registration cannot be triggered via API")
	}
	return nil
}

func (u *usersAPI) requireReadable(ctx context.Context, id int) error {
	sctx, ok := security.FromContext(ctx)
	if !ok || !sctx.IsAuthenticated() {
		return errors.UnauthorizedError(nil)
	}
	if !matchUserID(sctx, id) && !sctx.Can(ctx, rbac.ActionRead, userResource) {
		return errors.ForbiddenError(nil).WithMessage("Not authorized to read user: %d", id)
	}
	return nil
}

func (u *usersAPI) requireDeletable(ctx context.Context, id int) error {
	sctx, ok := security.FromContext(ctx)
	if !ok || !sctx.IsAuthenticated() {
		return errors.UnauthorizedError(nil)
	}
	if !sctx.Can(ctx, rbac.ActionDelete, userResource) {
		return errors.ForbiddenError(nil).WithMessage("Not authorized to delete users")
	}
	if matchUserID(sctx, id) || id == 1 {
		return errors.ForbiddenError(nil).WithMessage("User with ID %d cannot be deleted", id)
	}
	return nil
}

func (u *usersAPI) requireModifiable(ctx context.Context, id int) error {
	a, err := u.getAuth(ctx)
	if err != nil {
		return err
	}
	sctx, ok := security.FromContext(ctx)
	if !ok || !sctx.IsAuthenticated() {
		return errors.UnauthorizedError(nil)
	}
	if !modifiable(ctx, a, id) {
		return errors.ForbiddenError(nil).WithMessage("User with ID %d can't be updated", id)
	}
	return nil
}

func modifiable(ctx context.Context, authMode string, id int) bool {
	sctx, _ := security.FromContext(ctx)
	if authMode == common.DBAuth {

		// In db auth, admin can update anyone's info, and regular user can update his own
		return sctx.Can(ctx, rbac.ActionUpdate, userResource) || matchUserID(sctx, id)
	}
	// In none db auth, only the local admin's password can be updated.
	return id == 1 && sctx.Can(ctx, rbac.ActionUpdate, userResource)
}

func matchUserID(sctx security.Context, id int) bool {
	if localSCtx, ok := sctx.(*local.SecurityContext); ok {
		return localSCtx.User().UserID == id
	}
	return false
}

func requireValidSecret(in string) error {
	hasLower := regexp.MustCompile(`[a-z]`)
	hasUpper := regexp.MustCompile(`[A-Z]`)
	hasNumber := regexp.MustCompile(`[0-9]`)
	if len(in) >= 8 && hasLower.MatchString(in) && hasUpper.MatchString(in) && hasNumber.MatchString(in) {
		return nil
	}
	return errors.BadRequestError(nil).WithMessage("the password or secret must be longer than 8 chars with at least 1 uppercase letter, 1 lowercase letter and 1 number")
}

func validateUserProfile(user *commonmodels.User) error {
	if len(user.Email) > 0 {
		if m, _ := regexp.MatchString(`^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`, user.Email); !m {
			return errors.BadRequestError(nil).WithMessage("email with illegal format")
		}
	} else {
		return errors.BadRequestError(nil).WithMessage("email can't be empty")
	}

	if utils.IsIllegalLength(user.Realname, 1, 255) {
		return errors.BadRequestError(nil).WithMessage("realname with illegal length")
	}

	if utils.IsContainIllegalChar(user.Realname, []string{",", "~", "#", "$", "%"}) {
		return errors.BadRequestError(nil).WithMessage("realname contains illegal characters")
	}

	if utils.IsIllegalLength(user.Comment, -1, 30) {
		return errors.BadRequestError(nil).WithMessage("comment with illegal length")
	}

	return nil
}
