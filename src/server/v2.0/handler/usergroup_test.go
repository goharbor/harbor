package handler

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/usergroup"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/usergroup/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
)

// fakeUserGroupCtl only implements what the search handler uses and records the requested limit and offset
type fakeUserGroupCtl struct {
	usergroup.Controller
	groups []*model.UserGroup
	limit  int
	offset int
	calls  int
}

func (f *fakeUserGroupCtl) Count(_ context.Context, _ *q.Query) (int64, error) {
	return int64(len(f.groups)), nil
}

func (f *fakeUserGroupCtl) SearchByName(_ context.Context, _ string, limit, offset int) ([]*model.UserGroup, error) {
	f.limit, f.offset = limit, offset
	f.calls++
	return f.groups[min(offset, len(f.groups)):min(offset+limit, len(f.groups))], nil
}

type UserGroupTestSuite struct {
	htesting.Suite
	ctl *fakeUserGroupCtl
}

func (s *UserGroupTestSuite) SetupSuite() {
	s.ctl = &fakeUserGroupCtl{groups: []*model.UserGroup{
		{ID: 1, GroupName: "g1"}, {ID: 2, GroupName: "g2"},
		{ID: 3, GroupName: "g3"}, {ID: 4, GroupName: "g4"},
	}}
	s.Config = &restapi.Config{UsergroupAPI: &userGroupAPI{ctl: s.ctl}}
	s.Suite.SetupSuite()
	s.Security.On("IsAuthenticated").Return(true)
}

func (s *UserGroupTestSuite) TestSearchUserGroupsPagination() {
	cases := []struct {
		page     int64
		offset   int
		expected []string
		searched bool
	}{
		{1, 0, []string{"g1", "g2"}, true},
		{2, 2, []string{"g3", "g4"}, true},
		// a non-positive page number is the first page
		{0, 0, []string{"g1", "g2"}, true},
		{-1, 0, []string{"g1", "g2"}, true},
		{math.MinInt64, 0, []string{"g1", "g2"}, true},
		// pages past the end are answered without querying the database, even if the page number would overflow
		{3, 4, []string{}, false},
		{math.MaxInt64, 0, []string{}, false},
	}
	for _, c := range cases {
		s.ctl.calls = 0
		var result []*models.UserGroupSearchItem
		res, err := s.GetJSON(fmt.Sprintf("/usergroups/search?groupname=g&page=%d&page_size=2", c.page), &result)
		s.NoError(err)
		s.Equal(200, res.StatusCode)
		s.Equal("4", res.Header.Get("X-Total-Count"), "page %d", c.page)
		if c.searched {
			s.Equal(1, s.ctl.calls, "page %d", c.page)
			s.Equal(2, s.ctl.limit, "page %d", c.page)
			s.Equal(c.offset, s.ctl.offset, "page %d", c.page)
		} else {
			s.Equal(0, s.ctl.calls, "page %d", c.page)
		}
		names := []string{}
		for _, r := range result {
			names = append(names, r.GroupName)
		}
		s.Equal(c.expected, names, "page %d", c.page)
	}
}

func TestUserGroupTestSuite(t *testing.T) {
	suite.Run(t, &UserGroupTestSuite{})
}
