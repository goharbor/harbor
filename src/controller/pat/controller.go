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

package pat

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/pat"
	"github.com/goharbor/harbor/src/pkg/pat/model"
	"github.com/goharbor/harbor/src/pkg/project/models"
)

var (
	// Ctl is a global variable for the default PAT controller implementation
	Ctl = NewController()
)

// Controller to handle the requests related with personal access tokens
type Controller interface {
	// Create a new personal access token and return (id, plaintextSecret, error)
	Create(ctx context.Context, pat *model.PersonalAccessToken) (int64, string, error)

	// Get a personal access token by ID
	Get(ctx context.Context, id int64) (*model.PersonalAccessToken, error)

	// Count returns the total count of personal access tokens according to the query
	Count(ctx context.Context, query *q.Query) (int64, error)

	// List personal access tokens
	List(ctx context.Context, query *q.Query) ([]*model.PersonalAccessToken, error)

	// Update updates a personal access token
	Update(ctx context.Context, pat *model.PersonalAccessToken, props ...string) error

	// Delete deletes a personal access token
	Delete(ctx context.Context, id int64) error

	// RefreshSecret refreshes the secret of a personal access token
	RefreshSecret(ctx context.Context, id int64, newSecret string) (string, error)
}

// controller implements the Controller interface
type controller struct {
	patMgr     pat.Manager
	projectCtl project.Controller
	userCtl    user.Controller
}

// NewController returns a new PAT controller
func NewController() Controller {
	return &controller{
		patMgr:     pat.NewManager(),
		projectCtl: project.Ctl,
		userCtl:    user.Ctl,
	}
}

// Create creates a new personal access token
func (c *controller) Create(ctx context.Context, pat *model.PersonalAccessToken) (int64, string, error) {
	// Calculate expires_at based on duration (-1 means never expires)
	var expiresAt int64 = -1
	if pat.ExpiresAt > 0 {
		expiresAt = pat.ExpiresAt
	}

	// Generate secret using robot account's CreateSec utility
	secret, plaintextSecret, salt, err := robot.CreateSec()
	if err != nil {
		return 0, "", err
	}

	// Prefix the plaintext secret with "hbr_pat_" for new tokens
	fullPlaintextSecret := fmt.Sprintf("hbr_pat_%s", plaintextSecret)

	// Compute scope: use user-supplied scope if provided, otherwise auto-compute
	var scope string
	if pat.Scope != "" {
		scope, err = c.intersectScope(ctx, pat.UserID, pat.Scope)
		if err != nil {
			return 0, "", errors.Wrapf(err, "failed to process scope for user %d", pat.UserID)
		}
	} else {
		scope, err = c.computeScope(ctx, pat.UserID)
		if err != nil {
			return 0, "", errors.Wrapf(err, "failed to compute scope for user %d", pat.UserID)
		}
	}

	patToCreate := &model.PersonalAccessToken{
		UserID:      pat.UserID,
		Name:        pat.Name,
		Description: pat.Description,
		Secret:      secret,
		Salt:        salt,
		ExpiresAt:   expiresAt,
		Disabled:    false,
		IsLegacy:    false,
		Scope:       scope,
	}

	id, err := c.patMgr.Create(ctx, patToCreate)
	if err != nil {
		return 0, "", err
	}

	log.Debugf("created personal access token %d for user %d", id, pat.UserID)
	return id, fullPlaintextSecret, nil
}

// computeScope generates the scope for a PAT based on the user's project permissions
func (c *controller) computeScope(ctx context.Context, userID int) (string, error) {
	// Get the user
	u, err := c.userCtl.Get(ctx, userID, nil)
	if err != nil {
		return "[]", err
	}

	// List all projects the user has access to (including public projects)
	query := q.New(q.KeyWords{"member": &models.MemberQuery{UserID: userID}})
	query.PageSize = 1000
	projects, err := c.projectCtl.List(ctx, query, project.Metadata(false))
	if err != nil {
		return "[]", err
	}

	// Also get public projects
	publicQuery := q.New(q.KeyWords{"public": true})
	publicQuery.PageSize = 1000
	publicProjects, err := c.projectCtl.List(ctx, publicQuery, project.Metadata(false))
	if err != nil {
		return "[]", err
	}

	// Combine and deduplicate projects
	projectMap := make(map[int64]*models.Project)
	for _, p := range projects {
		projectMap[p.ProjectID] = p
	}
	for _, p := range publicProjects {
		projectMap[p.ProjectID] = p
	}

	var projectScopes []model.ProjectScope

	// For each project, determine push/pull permissions
	for _, p := range projectMap {
		roles, err := c.projectCtl.ListRoles(ctx, p.ProjectID, u)
		if err != nil {
			continue
		}

		// Determine access level based on roles
		hasPush := false
		hasPull := true // All project members can pull

		for _, role := range roles {
			if role == common.RoleProjectAdmin || role == common.RoleMaintainer || role == common.RoleDeveloper {
				hasPush = true
				break
			}
		}

		access := []model.AccessLevel{}
		if hasPull {
			access = append(access, model.AccessLevel{
				Resource: rbac.ResourceRepository.String(),
				Actions:  []string{"pull"},
			})
		}
		if hasPush {
			access = append(access, model.AccessLevel{
				Resource: rbac.ResourceRepository.String(),
				Actions:  []string{"push"},
			})
		}

		if len(access) > 0 {
			projectScopes = append(projectScopes, model.ProjectScope{
				ProjectID:   p.ProjectID,
				ProjectName: p.Name,
				Access:      access,
			})
		}
	}

	scopeJSON, err := json.Marshal(projectScopes)
	if err != nil {
		return "[]", err
	}

	return string(scopeJSON), nil
}

// intersectScope parses the user-supplied scope and intersects it with the user's
// actual project permissions. The user can only narrow their scope, never broaden it.
func (c *controller) intersectScope(ctx context.Context, userID int, userScope string) (string, error) {
	var requestedScopes []model.ProjectScope
	if err := json.Unmarshal([]byte(userScope), &requestedScopes); err != nil {
		return "[]", errors.Wrap(err, "invalid scope JSON")
	}

	// Get the user's full effective scope
	fullScope, err := c.computeScope(ctx, userID)
	if err != nil {
		return "[]", err
	}

	var fullScopes []model.ProjectScope
	if err := json.Unmarshal([]byte(fullScope), &fullScopes); err != nil {
		return "[]", errors.Wrap(err, "failed to parse computed scope")
	}

	// Build a lookup map from project ID -> AccessLevel (from full scope)
	type actionSet map[string]struct{}
	fullActionsByProject := make(map[int64]map[string]actionSet)
	for _, ps := range fullScopes {
		actionsByResource := make(map[string]actionSet)
		for _, al := range ps.Access {
			actions := make(actionSet)
			for _, a := range al.Actions {
				actions[a] = struct{}{}
			}
			actionsByResource[al.Resource] = actions
		}
		fullActionsByProject[ps.ProjectID] = actionsByResource
	}

	// Intersect: for each requested project scope, clamp actions to what the user actually has
	var intersected []model.ProjectScope
	for _, req := range requestedScopes {
		fullResources, exists := fullActionsByProject[req.ProjectID]
		if !exists {
			continue
		}
		var access []model.AccessLevel
		for _, al := range req.Access {
			fullActions, ok := fullResources[al.Resource]
			if !ok {
				continue
			}
			var allowedActions []string
			for _, a := range al.Actions {
				if _, allowed := fullActions[a]; allowed {
					allowedActions = append(allowedActions, a)
				}
			}
			if len(allowedActions) > 0 {
				access = append(access, model.AccessLevel{
					Resource: al.Resource,
					Actions:  allowedActions,
				})
			}
		}
		if len(access) > 0 {
			intersected = append(intersected, model.ProjectScope{
				ProjectID:   req.ProjectID,
				ProjectName: req.ProjectName,
				Access:      access,
			})
		}
	}

	result, err := json.Marshal(intersected)
	if err != nil {
		return "[]", err
	}
	return string(result), nil
}

// Get returns a personal access token by ID
func (c *controller) Get(ctx context.Context, id int64) (*model.PersonalAccessToken, error) {
	return c.patMgr.Get(ctx, id)
}

// Count returns the count of personal access tokens
func (c *controller) Count(ctx context.Context, query *q.Query) (int64, error) {
	return c.patMgr.Count(ctx, query)
}

// List lists personal access tokens
func (c *controller) List(ctx context.Context, query *q.Query) ([]*model.PersonalAccessToken, error) {
	return c.patMgr.List(ctx, query)
}

// Update updates a personal access token
func (c *controller) Update(ctx context.Context, pat *model.PersonalAccessToken, props ...string) error {
	if len(props) == 0 {
		props = []string{"name", "description", "disabled"}
	}
	return c.patMgr.Update(ctx, pat, props...)
}

// Delete deletes a personal access token
func (c *controller) Delete(ctx context.Context, id int64) error {
	return c.patMgr.Delete(ctx, id)
}

// RefreshSecret refreshes the secret of a personal access token
func (c *controller) RefreshSecret(ctx context.Context, id int64, newSecret string) (string, error) {
	pat, err := c.patMgr.Get(ctx, id)
	if err != nil {
		return "", err
	}

	var plaintextSecret string
	var secret string

	// If newSecret is provided, use it; otherwise generate a new one
	if newSecret != "" {
		// Validate the provided secret
		if !robot.IsValidSec(newSecret) {
			return "", fmt.Errorf("invalid secret format: must be 8-128 characters with at least one uppercase, lowercase, and digit")
		}
		plaintextSecret = newSecret
		secret = utils.Encrypt(newSecret, pat.Salt, utils.SHA256)
	} else {
		// Generate a new secret
		var generatedSecret string
		var genSecret string
		var genErr error
		generatedSecret, genSecret, _, genErr = robot.CreateSec(pat.Salt)
		if genErr != nil {
			return "", genErr
		}
		plaintextSecret = genSecret
		secret = generatedSecret
	}

	// Update the PAT with the new secret
	pat.Secret = secret
	if err := c.patMgr.Update(ctx, pat, "secret"); err != nil {
		return "", err
	}

	fullPlaintextSecret := fmt.Sprintf("hbr_pat_%s", plaintextSecret)
	log.Debugf("refreshed secret for personal access token %d", id)
	return fullPlaintextSecret, nil
}
