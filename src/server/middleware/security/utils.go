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

package security

import (
	"net/http"
	"strings"

	commonsecret "github.com/goharbor/harbor/src/common/secret"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/jobservice/job"
)

func bearerToken(req *http.Request) string {
	if req == nil {
		return ""
	}
	h := req.Header.Get("Authorization")
	token := strings.Split(h, "Bearer")
	if len(token) < 2 {
		return ""
	}
	return strings.TrimSpace(token[1])
}

// FromJobservice detects whether this request is from jobservice.
func FromJobservice(req *http.Request) bool {
	sc, ok := security.FromContext(req.Context())
	if !ok {
		return false
	}
	// check whether the user is jobservice user
	return sc.GetUsername() == commonsecret.JobserviceUser
}

// FromJobRetention detects whether this request is from tag retention job.
func FromJobRetention(req *http.Request) bool {
	if req != nil && req.Header != nil {
		return req.Header.Get("VendorType") == job.RetentionVendorType
	}

	return false
}
