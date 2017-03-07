package notary

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	notarytest "github.com/vmware/harbor/src/common/utils/notary/test"

	"path"
	"testing"
)

var endpoint = "10.117.4.142"

func TestMain(m *testing.M) {
	notaryServer := notarytest.NewNotaryServer(endpoint)
	defer notaryServer.Close()
	notaryEndpoint = notaryServer.URL
	notaryCachePath = "/tmp/notary"
	m.Run()
}

func TestGetTargets(t *testing.T) {
	targets, err := GetTargets("admin", path.Join(endpoint, "notary-demo/busybox"))
	assert.Nil(t, err, fmt.Sprintf("Unexpected error: %v", err))
	assert.Equal(t, 1, len(targets), "")
	assert.Equal(t, "1.0", targets[0].Tag, "")

	targets, err = GetTargets("admin", path.Join(endpoint, "notary-demo/notexist"))
	assert.Nil(t, err, fmt.Sprintf("Unexpected error: %v", err))
	assert.Equal(t, 0, len(targets), "Targets list should be empty for non exist repo.")
}
