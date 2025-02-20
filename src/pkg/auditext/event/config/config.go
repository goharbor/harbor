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

package config

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/goharbor/harbor/src/common/rbac"
	ctlevent "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	"github.com/goharbor/harbor/src/controller/event/model"
	"github.com/goharbor/harbor/src/lib/config"
	ext "github.com/goharbor/harbor/src/pkg/auditext/event"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

func init() {
	var configureEventResolver = &resolver{
		SensitiveAttributes: []string{"ldap_password", "oidc_client_secret"}, // all user config items with PasswordType defined in the metadatalist.go should be defined in SensitiveAttributes
	}
	commonevent.RegisterResolver(`/api/v2.0/configurations`, configureEventResolver)
}

const payloadSizeLimit = 450

// resolver used to resolve the configuration event
type resolver struct {
	SensitiveAttributes []string
}

func (c *resolver) Resolve(ce *commonevent.Metadata, evt *event.Event) error {
	e := &model.CommonEvent{}
	e.Operation = "update"
	e.Operator = ce.Username
	e.ResourceType = rbac.ResourceConfiguration.String()
	e.ResourceName = rbac.ResourceConfiguration.String()
	e.Payload = ext.Redact(ce.RequestPayload, c.SensitiveAttributes)
	e.OcurrAt = time.Now()
	if len(ce.RequestPayload) > payloadSizeLimit {
		ce.RequestPayload = fmt.Sprintf("%v...", ce.RequestPayload[:payloadSizeLimit])
	}
	e.OperationDescription = fmt.Sprintf("update configuration: %v", ce.RequestPayload)
	if ce.ResponseCode == http.StatusOK {
		e.IsSuccessful = true
	}
	evt.Topic = ctlevent.TopicCommonEvent
	evt.Data = e
	return nil
}

func (c *resolver) PreCheck(ctx context.Context, _ string, method string) (bool, string) {
	if method != http.MethodPut {
		return false, ""
	}
	return config.AuditLogEventEnabled(ctx, fmt.Sprintf("%v_%v", ext.MethodToOperation(method), rbac.ResourceConfiguration.String())), ""
}
