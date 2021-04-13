package signature

import (
	"github.com/goharbor/harbor/src/common"
	testutils "github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/lib/config"
	_ "github.com/goharbor/harbor/src/pkg/config/db"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	"github.com/goharbor/harbor/src/pkg/signature/notary/test"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	testutils.InitDatabaseFromEnv()
	// B/C the notary requires private key for signing token, b
	// before running locally, please make sure the env var is set as follow:
	// export TOKEN_PRIVATE_KEY_PATH="/harbor/tests/private_key.pem"
	endpoint := "jt-dev.local.goharbor.io"
	// notary-demo/busybox:1.0 is signed, more details in the notary/test pkg
	notaryServer := test.NewNotaryServer(endpoint)
	defer notaryServer.Close()
	conf := map[string]interface{}{
		common.WithNotary:  "true",
		common.NotaryURL:   notaryServer.URL,
		common.ExtEndpoint: "https://" + endpoint,
	}
	config.InitWithSettings(conf)
	result := m.Run()
	if result != 0 {
		os.Exit(result)
	}
}

func TestGetCheckerByRepo(t *testing.T) {
	type in struct {
		repo   string
		tag    string
		digest string
	}
	type res struct {
		tagSigned bool
		artSigned bool
	}
	m := GetManager()
	cases := []struct {
		input  in
		expect res
	}{
		{
			input: in{
				repo:   "library/busybox",
				tag:    "latest-signed",
				digest: "sha256:dd97a3fe6d721c5cf03abac0f50e2848dc583f7c4e41bf39102ceb42edfd1808",
			},
			expect: res{
				tagSigned: true,
				artSigned: true,
			},
		},
		{
			input: in{
				repo:   "library/busybox",
				tag:    "latest-signed",
				digest: "sha256:1359608115b94599e5641638bac5aef1ddfaa79bb96057ebf41ebc8d33acf8a8",
			},
			expect: res{
				tagSigned: false,
				artSigned: false,
			},
		},
		{
			input: in{
				repo:   "library/busybox",
				tag:    "2.0",
				digest: "sha256:dd97a3fe6d721c5cf03abac0f50e2848dc583f7c4e41bf39102ceb42edfd1808",
			},
			expect: res{
				tagSigned: false,
				artSigned: true,
			},
		},
		{
			input: in{
				repo:   "non-exist",
				tag:    "1.0",
				digest: "sha256:1359608115b94599e5641638bac5aef1ddfaa79bb96057ebf41ebc8d33acf8a7",
			},
			expect: res{
				tagSigned: false,
				artSigned: false,
			},
		},
	}
	for _, c := range cases {
		checker, err := m.GetCheckerByRepo(context.Background(), c.input.repo)
		assert.Nil(t, err)
		assert.Equal(t, c.expect.tagSigned, checker.IsTagSigned(c.input.tag, c.input.digest),
			"Unexpected tagSigned value for input: %#v", c.input)
		assert.Equal(t, c.expect.artSigned, checker.IsArtifactSigned(c.input.digest), "Unexpected artSigned value for input: %#v", c.input)
	}
}
