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

package robot

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

func init() {
	jwt.MarshalSingleStringAsArray = false
}

// Claim implements the interface of jwt.Claims
type Claim struct {
	jwt.RegisteredClaims
	TokenID   int64           `json:"id"`
	ProjectID int64           `json:"pid"`
	Access    []*types.Policy `json:"access"`
}

// Valid valid the claims "tokenID, projectID and access".
func (rc Claim) Valid() error {
	if rc.TokenID < 0 {
		return errors.New("token id must an valid INT")
	}
	if rc.ProjectID < 0 {
		return errors.New("project id must an valid INT")
	}
	if rc.Access == nil {
		return errors.New("the access info cannot be nil")
	}
	var v = jwt.NewValidator(jwt.WithLeeway(common.JwtLeeway))

	if stdErr := v.Validate(rc.RegisteredClaims); stdErr != nil {
		return stdErr
	}
	return nil
}
