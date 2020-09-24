package lib

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/testing/apitests/api-testing/client"
	"github.com/goharbor/harbor/src/testing/apitests/api-testing/models"
)

// SystemUtil : For getting system info
type SystemUtil struct {
	rootURI       string
	hostname      string
	testingClient *client.APIClient
}

// NewSystemUtil : Constructor
func NewSystemUtil(rootURI, hostname string, httpClient *client.APIClient) *SystemUtil {
	if len(strings.TrimSpace(rootURI)) == 0 || httpClient == nil {
		return nil
	}

	return &SystemUtil{
		rootURI:       rootURI,
		hostname:      hostname,
		testingClient: httpClient,
	}
}

// GetSystemInfo : Get systeminfo
func (nsu *SystemUtil) GetSystemInfo() error {
	url := nsu.rootURI + "/api/systeminfo"
	data, err := nsu.testingClient.Get(url)
	if err != nil {
		return err
	}

	var info models.SystemInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return err
	}

	if info.RegistryURL != nsu.hostname {
		return fmt.Errorf("Invalid registry url in system info: expect %s got %s ", nsu.hostname, info.RegistryURL)
	}

	return nil
}
