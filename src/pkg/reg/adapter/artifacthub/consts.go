package artifacthub

import (
	"errors"
	"fmt"
)

const (
	baseURL            = "https://artifacthub.io"
	getReplicationInfo = "/api/v1/harborReplication"
)

const (
	// HelmChart represents the kind of helm chart in artifact hub
	HelmChart = iota
	// FalcoRules represents the kind of falco rules in artifact hub
	FalcoRules
	// OPAPolicies represents the kind of OPA policies in artifact hub
	OPAPolicies
	// OLMOperators represents the kind of OLM operators in artifact hub
	OLMOperators
)

// ErrHTTPNotFound defines the return error when receiving 404 response code
var ErrHTTPNotFound = errors.New("not found")

func searchPackages(kind, offset, limit int, queryString string) string {
	if len(queryString) == 0 {
		return fmt.Sprintf("/api/v1/packages/search?kind=%d&limit=%d&offset=%d",
			kind, limit, offset)
	}
	return fmt.Sprintf("/api/v1/packages/search?kind=%d&limit=%d&offset=%d&ts_query_web=%s",
		kind, limit, offset, queryString)
}

func getHelmPackageDetail(fullName string) string {
	return fmt.Sprintf("/api/v1/packages/helm/%s", fullName)
}

func getHelmVersion(fullName, version string) string {
	return fmt.Sprintf("/api/v1/packages/helm/%s/%s", fullName, version)
}
