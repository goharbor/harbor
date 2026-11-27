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

package usergroup

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/ldap"
	"github.com/goharbor/harbor/src/pkg/usergroup/model"
)

func TestCreateLDAPGroupSearchErrors(t *testing.T) {
	connErr := errors.New("LDAP Result Code 200 \"Network Error\": connection refused")

	tests := []struct {
		name        string
		searchErr   error
		wantBadReq  bool
		wantMessage string
		wantErr     error
	}{
		{name: "ldap not found", searchErr: ldap.ErrNotFound, wantBadReq: true, wantMessage: "LDAP Group DN is not found"},
		{name: "harbor not found", searchErr: errors.NotFoundError(nil), wantBadReq: true, wantMessage: "LDAP Group DN is not found"},
		{name: "no group and no error", searchErr: nil, wantBadReq: true, wantMessage: "LDAP Group DN is not found"},
		{name: "ldap DN syntax", searchErr: ldap.ErrDNSyntax, wantBadReq: true, wantMessage: "invalid DN syntax"},
		{name: "auth invalid DN", searchErr: auth.ErrInvalidLDAPGroupDN, wantBadReq: true, wantMessage: "invalid DN syntax"},
		{name: "connection failure", searchErr: connErr, wantErr: connErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &controller{
				searchGroup: func(_ context.Context, _ string) (*model.UserGroup, error) {
					return nil, tt.searchErr
				},
			}

			_, err := c.Create(context.Background(), model.UserGroup{
				GroupType:   common.LDAPGroupType,
				LdapGroupDN: "cn=harbor_users,ou=groups,dc=example,dc=com",
			})

			require.Error(t, err)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.False(t, errors.IsErr(err, errors.BadRequestCode))
				return
			}
			assert.Equal(t, tt.wantBadReq, errors.IsErr(err, errors.BadRequestCode))
			assert.Contains(t, err.Error(), tt.wantMessage)
		})
	}
}
