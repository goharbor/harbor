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

package handler

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/controller/project"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

const (
	// ScheduleHourly : 'Hourly'
	ScheduleHourly = "Hourly"
	// ScheduleDaily : 'Daily'
	ScheduleDaily = "Daily"
	// ScheduleWeekly : 'Weekly'
	ScheduleWeekly = "Weekly"
	// ScheduleCustom : 'Custom'
	ScheduleCustom = "Custom"
	// ScheduleManual : 'Manual'
	ScheduleManual = "Manual"
	// ScheduleNone : 'None'
	ScheduleNone = "None"
)

func parseScanReportMimeTypes(header *string) []string {
	var mimeTypes []string

	if header != nil {
		for _, mimeType := range strings.Split(*header, ",") {
			mimeType = strings.TrimSpace(mimeType)
			switch mimeType {
			case v1.MimeTypeNativeReport, v1.MimeTypeGenericVulnerabilityReport:
				mimeTypes = append(mimeTypes, mimeType)
			}
		}
	}

	if len(mimeTypes) == 0 {
		mimeTypes = append(mimeTypes, v1.MimeTypeNativeReport)
	}

	return mimeTypes
}

func unescapePathParams(params interface{}, fieldNames ...string) error {
	val := reflect.ValueOf(params)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("params must be ptr")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("params must be struct")
	}

	for _, name := range fieldNames {
		field := val.FieldByName(name)
		if !field.IsValid() {
			log.Debugf("field %s not found in %s", name, val.Type().Name())
			continue
		}

		if !field.CanSet() {
			log.Debugf("field %s can not be changed in %s", name, val.Type().Name())
			continue
		}

		switch field.Type().Kind() {
		case reflect.String:
			v, err := url.PathUnescape(field.String())
			if err != nil {
				return err
			}
			field.SetString(v)
		default:
			log.Debugf("field %s can not be unescaped in %s", name, val.Type().Name())
		}
	}

	return nil
}

func parseProjectNameOrID(str string, isResourceName *bool) interface{} {
	if lib.BoolValue(isResourceName) {
		// always as projectName
		return str
	}

	v, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		// it's projectName
		return str
	}

	return v // projectID
}

func getProjectID(ctx context.Context, projectNameOrID interface{}) (int64, error) {
	projectName, ok := projectNameOrID.(string)
	if ok {
		p, err := project.Ctl.Get(ctx, projectName, project.Metadata(false))
		if err != nil {
			return 0, err
		}
		return p.ProjectID, nil
	}
	projectID, ok := projectNameOrID.(int64)
	if ok {
		return projectID, nil
	}
	return 0, errors.New("unknown project identifier type")
}
