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

package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	htesting "github.com/goharbor/harbor/src/testing"
)

var errRollback = errors.New("rollback the changes of the test")

type OnboardOIDCUserTestSuite struct {
	htesting.Suite
}

func newOIDCUser(username, subject string) *commonmodels.User {
	return &commonmodels.User{
		Username: username,
		Realname: username,
		OIDCUserMeta: &commonmodels.OIDCUser{
			SubIss: subject + "https://issuer.example.com",
			Secret: "secret",
			Token:  "token",
		},
	}
}

func (s *OnboardOIDCUserTestSuite) TestOnboard() {
	err := orm.WithTransaction(func(ctx context.Context) error {
		u := newOIDCUser("onboard-test", "subject")
		s.Require().NoError(Ctl.OnboardOIDCUser(ctx, u))
		s.NotZero(u.UserID)
		s.Equal(u.UserID, u.OIDCUserMeta.UserID)

		got, err := Ctl.GetBySubIss(ctx, "subject", "https://issuer.example.com")
		s.Require().NoError(err)
		s.Equal(u.UserID, got.UserID)
		return errRollback
	})(s.Context())
	s.ErrorIs(err, errRollback)
}

func (s *OnboardOIDCUserTestSuite) TestConflictKeepsTransactionUsable() {
	err := orm.WithTransaction(func(ctx context.Context) error {
		first := newOIDCUser("onboard-test", "subject")
		s.Require().NoError(Ctl.OnboardOIDCUser(ctx, first))

		// the same OIDC identity with another username, which conflicts on the OIDC meta after the user is created
		second := newOIDCUser("onboard-test-2", "subject")
		err := Ctl.OnboardOIDCUser(ctx, second)
		s.Equal(errors.ConflictCode, errors.ErrCode(err))

		// the transaction can still be used, and the user of the failed onboarding is rolled back
		_, err = Ctl.GetByName(ctx, "onboard-test-2")
		s.True(errors.IsNotFoundErr(err), "unexpected error: %v", err)
		got, err := Ctl.GetBySubIss(ctx, "subject", "https://issuer.example.com")
		s.Require().NoError(err)
		s.Equal(first.UserID, got.UserID)
		return errRollback
	})(s.Context())
	s.ErrorIs(err, errRollback)
}

func TestOnboardOIDCUserTestSuite(t *testing.T) {
	suite.Run(t, &OnboardOIDCUserTestSuite{})
}
