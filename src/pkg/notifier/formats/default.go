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

package formats

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

var (
	// defaultFormatter is the global single formatter for Default.
	defaultFormatter Formatter = &Default{}
)

func init() {
	// for forward compatibility, empty is also the default.
	registerFormats("", defaultFormatter)
	registerFormats(DefaultFormat, defaultFormatter)
}

const (
	// DefaultFormat is the type for default format.
	DefaultFormat = "Default"
)

// Default is the instance for default format(original format in previous versions).
type Default struct{}

// Format implements the interface Formatter.
/*
{
   "type":"PULL_ARTIFACT",
   "occur_at":1678082303,
   "operator":"admin",
   "event_data":{
      "resources":[
         {
            "digest":"sha256:dde8e930c7b6a490f728e66292bc9bce42efc9bbb5278bae40e4f30f6e00fe8c",
            "tag":"sha256:dde8e930c7b6a490f728e66292bc9bce42efc9bbb5278bae40e4f30f6e00fe8c",
            "resource_url":"harbor.dev/library/busybox@sha256:dde8e930c7b6a490f728e66292bc9bce42efc9bbb5278bae40e4f30f6e00fe8c"
         }
      ],
      "repository":{
         "date_created":1677053165,
         "name":"busybox",
         "namespace":"library",
         "repo_full_name":"library/busybox",
         "repo_type":"public"
      }
   }
}
*/
func (d *Default) Format(_ context.Context, he *model.HookEvent) (http.Header, []byte, error) {
	if he == nil {
		return nil, nil, errors.Errorf("HookEvent should not be nil")
	}

	payload, err := json.Marshal(he.Payload)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error to marshal payload")
	}

	header := http.Header{
		"Content-Type": []string{"application/json"},
	}
	return header, payload, nil
}
