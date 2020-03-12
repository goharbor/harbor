package token

import (
	"os"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	robot_claim "github.com/goharbor/harbor/src/pkg/token/claims/robot"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	config.Init()

	result := m.Run()
	if result != 0 {
		os.Exit(result)
	}
}

func TestNew(t *testing.T) {
	rbacPolicy := &types.Policy{
		Resource: "/project/libray/repository",
		Action:   "pull",
	}
	policies := []*types.Policy{}
	policies = append(policies, rbacPolicy)

	tokenID := int64(123)
	projectID := int64(321)
	tokenExpiration := time.Duration(10) * 24 * time.Hour
	expiresAt := time.Now().UTC().Add(tokenExpiration).Unix()
	robot := robot_claim.Claim{
		TokenID:   tokenID,
		ProjectID: projectID,
		Access:    policies,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}
	token, err := New(DefaultTokenOptions(), robot)

	assert.Nil(t, err)
	assert.Equal(t, token.Header["alg"], "RS256")
	assert.Equal(t, token.Header["typ"], "JWT")

}

func TestRaw(t *testing.T) {
	rbacPolicy := &types.Policy{
		Resource: "/project/library/repository",
		Action:   "pull",
	}
	policies := []*types.Policy{}
	policies = append(policies, rbacPolicy)

	tokenID := int64(123)
	projectID := int64(321)

	tokenExpiration := time.Duration(10) * 24 * time.Hour
	expiresAt := time.Now().UTC().Add(tokenExpiration).Unix()
	robot := robot_claim.Claim{
		TokenID:   tokenID,
		ProjectID: projectID,
		Access:    policies,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}
	token, err := New(DefaultTokenOptions(), robot)
	assert.Nil(t, err)

	rawTk, err := token.Raw()
	assert.Nil(t, err)
	assert.NotNil(t, rawTk)
}

func TestParseWithClaims(t *testing.T) {
	rawTk := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MTIzLCJQcm9qZWN0SUQiOjAsIkFjY2VzcyI6W3siUmVzb3VyY2UiOiIvcHJvamVjdC9saWJyYXkvcmVwb3NpdG9yeSIsIkFjdGlvbiI6InB1bGwiLCJFZmZlY3QiOiIifV0sIlN0YW5kYXJkQ2xhaW1zIjp7ImV4cCI6MTU0ODE0MDIyOSwiaXNzIjoiaGFyYm9yLXRva2VuLWlzc3VlciJ9fQ.Jc3qSKN4SJVUzAvBvemVpRcSOZaHlu0Avqms04qzPm4ru9-r9IRIl3mnSkI6m9XkzLUeJ7Kiwyw63ghngnVKw_PupeclOGC6s3TK5Cfmo4h-lflecXjZWwyy-dtH_e7Us_ItS-R3nXDJtzSLEpsGHCcAj-1X2s93RB2qD8LNSylvYeDezVkTzqRzzfawPJheKKh9JTrz-3eUxCwQard9-xjlwvfUYULoHTn9npNAUq4-jqhipW4uE8HL-ym33AGF57la8U0RO11hmDM5K8-PiYknbqJ_oONeS3HBNym2pEFeGjtTv2co213wl4T5lemlg4SGolMBuJ03L7_beVZ0o-MKTkKDqDwJalb6_PM-7u3RbxC9IzJMiwZKIPnD3FvV10iPxUUQHaH8Jz5UZ2pFIhi_8BNnlBfT0JOPFVYATtLjHMczZelj2YvAeR1UHBzq3E0jPpjjwlqIFgaHCaN_KMwEvadTo_Fi2sEH4pNGP7M3yehU_72oLJQgF4paJarsmEoij6ZtPs6xekBz1fccVitq_8WNIz9aeCUdkUBRwI5QKw1RdW4ua-w74ld5MZStWJA8veyoLkEb_Q9eq2oAj5KWFjJbW5-ltiIfM8gxKflsrkWAidYGcEIYcuXr7UdqEKXxtPiWM0xb3B91ovYvO5402bn3f9-UGtlcestxNHA"
	rClaims := &robot_claim.Claim{}
	_, _ = Parse(DefaultTokenOptions(), rawTk, rClaims)
	assert.Equal(t, int64(123), rClaims.TokenID)
	assert.Equal(t, int64(0), rClaims.ProjectID)
	assert.Equal(t, "/project/libray/repository", rClaims.Access[0].Resource.String())
}
