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
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/scan/report"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

func boolValue(v *bool) bool {
	if v != nil {
		return *v
	}

	return false
}

func resolveVulnerabilitiesAddition(ctx context.Context, artifact *artifact.Artifact) (*processor.Addition, error) {
	reports, err := scan.DefaultController.GetReport(ctx, artifact, []string{v1.MimeTypeNativeReport})
	if err != nil {
		return nil, err
	}

	vulnerabilities := make(map[string]interface{})
	for _, rp := range reports {
		// Resolve scan report data only when it is ready
		if len(rp.Report) == 0 {
			continue
		}

		vrp, err := report.ResolveData(rp.MimeType, []byte(rp.Report), report.WithArtifactDigest(rp.Digest))
		if err != nil {
			return nil, err
		}

		if v, ok := vulnerabilities[rp.MimeType]; ok {
			r, err := report.Merge(rp.MimeType, v, vrp)
			if err != nil {
				return nil, err
			}
			vulnerabilities[rp.MimeType] = r
		} else {
			vulnerabilities[rp.MimeType] = vrp
		}
	}

	content, _ := json.Marshal(vulnerabilities)

	return &processor.Addition{
		Content:     content,
		ContentType: "application/json",
	}, nil
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
