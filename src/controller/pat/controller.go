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
	"fmt"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	pat "github.com/goharbor/harbor/src/pkg/pat"
	"github.com/goharbor/harbor/src/pkg/pat/model"
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
	patMgr pat.Manager
}

// NewController returns a new PAT controller
func NewController() Controller {
	return &controller{
		patMgr: pat.NewManager(),
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

	patToCreate := &model.PersonalAccessToken{
		UserID:      pat.UserID,
		Name:        pat.Name,
		Description: pat.Description,
		Secret:      secret,
		Salt:        salt,
		ExpiresAt:   expiresAt,
		Disabled:    false,
		IsLegacy:    false,
	}

	id, err := c.patMgr.Create(ctx, patToCreate)
	if err != nil {
		return 0, "", err
	}

	log.Debugf("created personal access token %d for user %d", id, pat.UserID)
	return id, fullPlaintextSecret, nil
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
