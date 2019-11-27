package authproxy

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/dao/group"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/authentication/v1beta1"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	dao.PrepareTestForPostgresSQL()
	result := m.Run()
	if result != 0 {
		os.Exit(result)
	}
}

func TestUserFromReviewStatus(t *testing.T) {
	type result struct {
		hasErr   bool
		username string
		groupLen int
	}
	cases := []struct {
		input  v1beta1.TokenReviewStatus
		expect result
	}{
		{
			input: v1beta1.TokenReviewStatus{
				Authenticated: false,
				Error:         "connection error",
			},
			expect: result{
				hasErr: true,
			},
		},
		{
			input: v1beta1.TokenReviewStatus{
				Authenticated: true,
				User: v1beta1.UserInfo{
					Username: "jack",
					UID:      "u-1",
				},
			},
			expect: result{
				hasErr:   false,
				username: "jack",
				groupLen: 0,
			},
		},
		{
			input: v1beta1.TokenReviewStatus{
				Authenticated: true,
				User: v1beta1.UserInfo{
					Username: "daniel",
					Groups:   []string{"group1", "group2"},
				},
				Error: "",
			},
			expect: result{
				hasErr:   false,
				username: "daniel",
				groupLen: 2,
			},
		},
	}
	for _, c := range cases {
		u, err := UserFromReviewStatus(c.input)
		if c.expect.hasErr == true {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, c.expect.username, u.Username)
			assert.Equal(t, c.expect.groupLen, len(u.GroupIDs))
		}
		if u != nil {
			for _, gid := range u.GroupIDs {
				t.Logf("Deleting group %d", gid)
				if err := group.DeleteUserGroup(gid); err != nil {
					panic(err)
				}
			}
		}
	}
}
