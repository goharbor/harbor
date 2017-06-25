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
	"github.com/vmware/harbor/src/common/models"
	"strings"
)

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
