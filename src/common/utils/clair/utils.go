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

package clair

import (
	"github.com/goharbor/harbor/src/common/models"
	"strings"
)

// var client = NewClient()

// ParseClairSev parse the severity of clair to Harbor's Severity type if the string is not recognized the value will be set to unknown.
func ParseClairSev(clairSev string) models.Severity {
	sev := strings.ToLower(clairSev)
	switch sev {
	case models.SeverityNone:
		return models.SevNone
	case models.SeverityLow:
		return models.SevLow
	case models.SeverityMedium:
		return models.SevMedium
	case models.SeverityHigh, models.SeverityCritical:
		return models.SevHigh
	default:
		return models.SevUnknown
	}
}

func transformVuln(clairVuln *models.ClairLayerEnvelope) (*models.ComponentsOverview, models.Severity) {
	vulnMap := make(map[models.Severity]int)
	features := clairVuln.Layer.Features
	totalComponents := len(features)
	var temp models.Severity
	for _, f := range features {
		sev := models.SevNone
		for _, v := range f.Vulnerabilities {
			temp = ParseClairSev(v.Severity)
			if temp > sev {
				sev = temp
			}
		}
		vulnMap[sev]++
	}
	overallSev := models.SevNone
	compSummary := []*models.ComponentsOverviewEntry{}
	for k, v := range vulnMap {
		if k > overallSev {
			overallSev = k
		}
		entry := &models.ComponentsOverviewEntry{
			Sev:   int(k),
			Count: v,
		}
		compSummary = append(compSummary, entry)
	}
	return &models.ComponentsOverview{
		Total:   totalComponents,
		Summary: compSummary,
	}, overallSev
}

// TransformVuln is for running scanning job in both job service V1 and V2.
func TransformVuln(clairVuln *models.ClairLayerEnvelope) (*models.ComponentsOverview, models.Severity) {
	return transformVuln(clairVuln)
}
