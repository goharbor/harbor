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

package scan

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/clair"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"reflect"
)

// VulnerabilityItem represents a vulnerability reported by scanner
type VulnerabilityItem struct {
	ID          string          `json:"id"`
	Severity    models.Severity `json:"severity"`
	Pkg         string          `json:"package"`
	Version     string          `json:"version"`
	Description string          `json:"description"`
	Link        string          `json:"link"`
	Fixed       string          `json:"fixedVersion,omitempty"`
}

// VulnerabilityList is a list of vulnerabilities, which should be scanner-agnostic
type VulnerabilityList []VulnerabilityItem

// ApplyWhitelist filters out the CVE defined in the whitelist in the parm.
// It returns the items that are filtered for the caller to track or log.
func (vl *VulnerabilityList) ApplyWhitelist(whitelist models.CVEWhitelist) VulnerabilityList {
	filtered := VulnerabilityList{}
	if whitelist.IsExpired() {
		log.Info("The input whitelist is expired, skip filtering")
		return filtered
	}
	s := whitelist.CVESet()
	r := (*vl)[:0]
	for _, v := range *vl {
		if _, ok := s[v.ID]; ok {
			log.Debugf("Filtered Vulnerability in whitelist, CVE ID: %s, severity: %s", v.ID, v.Severity)
			filtered = append(filtered, v)
		} else {
			r = append(r, v)
		}
	}
	val := reflect.ValueOf(vl)
	val.Elem().SetLen(len(r))
	return filtered
}

// Severity returns the highest severity of the vulnerabilities in the list
func (vl *VulnerabilityList) Severity() models.Severity {
	s := models.SevNone
	for _, v := range *vl {
		if v.Severity > s {
			s = v.Severity
		}
	}
	return s
}

// HasCVE returns whether the vulnerability list has the vulnerability with CVE ID in the parm
func (vl *VulnerabilityList) HasCVE(id string) bool {
	for _, v := range *vl {
		if v.ID == id {
			return true
		}
	}
	return false
}

// VulnListFromClairResult transforms the returned value of Clair API to a VulnerabilityList
func VulnListFromClairResult(layerWithVuln *models.ClairLayerEnvelope) VulnerabilityList {
	res := VulnerabilityList{}
	if layerWithVuln == nil {
		return res
	}
	l := layerWithVuln.Layer
	if l == nil {
		return res
	}
	features := l.Features
	if features == nil {
		return res
	}
	for _, f := range features {
		vulnerabilities := f.Vulnerabilities
		if vulnerabilities == nil {
			continue
		}
		for _, v := range vulnerabilities {
			vItem := VulnerabilityItem{
				ID:          v.Name,
				Pkg:         f.Name,
				Version:     f.Version,
				Severity:    clair.ParseClairSev(v.Severity),
				Fixed:       v.FixedBy,
				Link:        v.Link,
				Description: v.Description,
			}
			res = append(res, vItem)
		}
	}
	return res
}

// VulnListByDigest returns the VulnerabilityList based on the scan result of artifact with the digest in the parm
func VulnListByDigest(digest string) (VulnerabilityList, error) {
	var res VulnerabilityList
	overview, err := dao.GetImgScanOverview(digest)
	if err != nil {
		return res, err
	}
	if overview == nil || len(overview.DetailsKey) == 0 {
		return res, fmt.Errorf("unable to get the scan result for digest: %s, the artifact is not scanned", digest)
	}
	c := clair.NewClient(config.ClairEndpoint(), nil)
	clairRes, err := c.GetResult(overview.DetailsKey)
	if err != nil {
		return res, fmt.Errorf("failed to get scan result from Clair, error: %v", err)
	}
	return VulnListFromClairResult(clairRes), nil
}
