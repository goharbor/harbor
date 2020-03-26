package chart

import (
	helm_chart "helm.sh/helm/v3/pkg/chart"
	"time"
)

// VersionDetails keeps the detailed data info of the chart version
type VersionDetails struct {
	Dependencies []*helm_chart.Dependency `json:"dependencies"`
	Values       map[string]interface{}   `json:"values"`
	Files        map[string]string        `json:"files"`
	Security     *SecurityReport          `json:"security"`
}

// SecurityReport keeps the info related with security
// e.g.: digital signature, vulnerability scanning etc.
type SecurityReport struct {
	Signature *DigitalSignature `json:"signature"`
}

// DigitalSignature used to indicate if the chart has been signed
type DigitalSignature struct {
	Signed     bool   `json:"signed"`
	Provenance string `json:"prov_file"`
}

// Info keeps the information of the chart
type Info struct {
	Name          string    `json:"name"`
	TotalVersions uint32    `json:"total_versions"`
	LatestVersion string    `json:"latest_version"`
	Created       time.Time `json:"created"`
	Updated       time.Time `json:"updated"`
	Icon          string    `json:"icon"`
	Home          string    `json:"home"`
	Deprecated    bool      `json:"deprecated"`
}
