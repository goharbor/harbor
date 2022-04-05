package systeminfo

import (
	"context"
	"github.com/goharbor/harbor/src/lib/config"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	htesting "github.com/goharbor/harbor/src/testing"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/version"
	"github.com/stretchr/testify/suite"
)

type sysInfoCtlTestSuite struct {
	htesting.Suite
	ctl Controller
}

func (s *sysInfoCtlTestSuite) SetupTest() {
	version.ReleaseVersion = "test"
	version.GitCommit = "fakeid"

	conf := map[string]interface{}{
		common.AUTHMode:                    "db_auth",
		common.SelfRegistration:            true,
		common.ExtEndpoint:                 "https://test.goharbor.io",
		common.ProjectCreationRestriction:  "everyone",
		common.RegistryStorageProviderName: "filesystem",
		common.ReadOnly:                    false,
		common.NotificationEnable:          false,
		common.WithChartMuseum:             false,
		common.WithNotary:                  true,
	}

	config.InitWithSettings(conf)
	s.ctl = Ctl
}

func (s *sysInfoCtlTestSuite) TestGetCert() {
	assert := s.Assert()
	testRootCertPath = "./notexist.crt"
	rc, err := s.ctl.GetCA(context.Background())
	assert.Nil(rc)
	assert.NotNil(err)
	assert.True(errors.IsNotFoundErr(err))
}

func (s *sysInfoCtlTestSuite) TestGetInfo() {
	assert := s.Assert()
	cases := []struct {
		withProtected bool
		expect        Data
	}{
		{
			withProtected: false,
			expect: Data{
				AuthMode:         "db_auth",
				HarborVersion:    "test-fakeid",
				SelfRegistration: true,
			},
		},
		{
			withProtected: true,
			expect: Data{
				AuthMode:         "db_auth",
				HarborVersion:    "test-fakeid",
				SelfRegistration: true,
				Protected: &protectedData{
					WithNotary:              true,
					RegistryURL:             "test.goharbor.io",
					ExtURL:                  "https://test.goharbor.io",
					ProjectCreationRestrict: "everyone",
					// CI pipeline has it
					HasCARoot:                   true,
					RegistryStorageProviderName: "filesystem",
					ReadOnly:                    false,
					WithChartMuseum:             false,
					NotificationEnable:          false,
				},
			},
		},
	}
	for _, tc := range cases {
		res, err := s.ctl.GetInfo(context.Background(), Options{
			WithProtectedInfo: tc.withProtected,
		})
		assert.Nil(err)
		exp := tc.expect
		if exp.Protected == nil {
			assert.Nil(res.Protected)
			assert.Equal(exp, *res)
		} else {
			// skip comparing exp.Protected.CurrentTime with res.Protected.CurrentTime
			exp.Protected.CurrentTime = res.Protected.CurrentTime
			assert.Equal(*exp.Protected, *res.Protected)
			exp.Protected = nil
			res.Protected = nil
			assert.Equal(exp, *res)
		}
	}
}

func TestControllerSuite(t *testing.T) {
	suite.Run(t, &sysInfoCtlTestSuite{})
}
