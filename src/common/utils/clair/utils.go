// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"

	"fmt"
	"strings"
)

//var client = NewClient()

// ParseClairSev parse the severity of clair to Harbor's Severity type if the string is not recognized the value will be set to unknown.
func ParseClairSev(clairSev string) models.Severity {
	sev := strings.ToLower(clairSev)
	switch sev {
	case "negligible":
		return models.SevNone
	case "low":
		return models.SevLow
	case "medium":
		return models.SevMedium
	case "high":
		return models.SevHigh
	default:
		return models.SevUnknown
	}
}

// UpdateScanOverview qeuries the vulnerability based on the layerName and update the record in img_scan_overview table based on digest.
func UpdateScanOverview(digest, layerName string, l ...*log.Logger) error {
	var logger *log.Logger
	if len(l) > 1 {
		return fmt.Errorf("More than one logger specified")
	} else if len(l) == 1 {
		logger = l[0]
	} else {
		logger = log.DefaultLogger()
	}
	client := NewClient(common.DefaultClairEndpoint, logger)
	res, err := client.GetResult(layerName)
	if err != nil {
		logger.Errorf("Failed to get result from Clair, error: %v", err)
		return err
	}
	vulnMap := make(map[models.Severity]int)
	features := res.Layer.Features
	totalComponents := len(features)
	logger.Infof("total features: %d", totalComponents)
	var temp models.Severity
	for _, f := range features {
		sev := models.SevNone
		for _, v := range f.Vulnerabilities {
			temp = ParseClairSev(v.Severity)
			if temp > sev {
				sev = temp
			}
		}
		logger.Infof("Feature: %s, Severity: %d", f.Name, sev)
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
	compOverview := &models.ComponentsOverview{
		Total:   totalComponents,
		Summary: compSummary,
	}
	return dao.UpdateImgScanOverview(digest, layerName, overallSev, compOverview)
}
