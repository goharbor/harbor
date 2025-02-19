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

package event

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"slices"

	ctlevent "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	"github.com/goharbor/harbor/src/controller/event/model"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

const (
	createOp          = "create"
	updateOp          = "update"
	deleteOp          = "delete"
	resourceIDPattern = `^%v/(\d+)$`
)

// ResolveIDToNameFunc is the function to resolve the resource name from resource id
type ResolveIDToNameFunc func(string) string

type Resolver struct {
	BaseURLPattern string
	ResourceType   string
	SucceedCodes   []int
	// SensitiveAttributes is the attributes that need to be redacted
	SensitiveAttributes []string
	// ShouldResolveName indicates if the resource name should be resolved before delete, if true, need to resolve the resource name before delete
	ShouldResolveName bool
	// IDToNameFunc is used to resolve the resource name from resource id
	IDToNameFunc ResolveIDToNameFunc
}

// PreCheck check if the event should be captured and resolve the resource name if needed, if need to resolve the resource name, return the resource name
func (e *Resolver) PreCheck(ctx context.Context, url string, method string) (capture bool, resourceName string) {
	capture = config.AuditLogEventEnabled(ctx, fmt.Sprintf("%v_%v", MethodToOperation(method), e.ResourceType))
	if !capture {
		return false, ""
	}
	// for delete operation on a resource has name, need to resolve the resource id to resource name before delete
	resName := ""
	if capture && method == http.MethodDelete && e.ShouldResolveName {
		re := regexp.MustCompile(fmt.Sprintf(resourceIDPattern, e.BaseURLPattern))
		m := re.FindStringSubmatch(url)
		if len(m) == 2 && e.IDToNameFunc != nil {
			resName = e.IDToNameFunc(m[1])
		}
	}
	return true, resName
}

// Resolve ...
func (e *Resolver) Resolve(ce *commonevent.Metadata, event *event.Event) error {
	if ce.RequestMethod != http.MethodPost && ce.RequestMethod != http.MethodDelete && ce.RequestMethod != http.MethodPut {
		return nil
	}
	evt := &model.CommonEvent{
		Operator:     ce.Username,
		ResourceType: e.ResourceType,
		OcurrAt:      time.Now(),
		IsSuccessful: true,
	}
	resourceName := ""
	operation := MethodToOperation(ce.RequestMethod)
	if len(operation) == 0 {
		return nil
	}
	evt.Operation = operation

	switch evt.Operation {
	case createOp:
		if len(ce.ResponseLocation) > 0 {
			// extract resource id from response location
			re := regexp.MustCompile(fmt.Sprintf(resourceIDPattern, e.BaseURLPattern))
			m := re.FindStringSubmatch(ce.ResponseLocation)
			if len(m) != 2 {
				return nil
			}
			evt.ResourceName = m[1]
			if e.IDToNameFunc != nil {
				resourceName = e.IDToNameFunc(m[1])
			}
		}
		if e.ShouldResolveName && resourceName != "" {
			evt.ResourceName = resourceName
		}

	case deleteOp:
		re := regexp.MustCompile(fmt.Sprintf(resourceIDPattern, e.BaseURLPattern))
		m := re.FindStringSubmatch(ce.RequestURL)
		if len(m) != 2 {
			return nil
		}
		evt.ResourceName = m[1]
		if e.ShouldResolveName && ce.ResourceName != "" {
			evt.ResourceName = ce.ResourceName
		}

	case updateOp:
		re := regexp.MustCompile(fmt.Sprintf(resourceIDPattern, e.BaseURLPattern))
		m := re.FindStringSubmatch(ce.RequestURL)
		if len(m) != 2 {
			return nil
		}
		evt.ResourceName = m[1]
		if e.IDToNameFunc != nil {
			resourceName = e.IDToNameFunc(m[1])
		}
		if e.ShouldResolveName && resourceName != "" {
			evt.ResourceName = resourceName
		}
	}

	evt.OperationDescription = fmt.Sprintf("%s %s with name: %s", evt.Operation, e.ResourceType, evt.ResourceName)

	if !slices.Contains(e.SucceedCodes, ce.ResponseCode) {
		evt.IsSuccessful = false
	}

	event.Topic = ctlevent.TopicCommonEvent
	event.Data = evt
	return nil
}

// MethodToOperation converts HTTP method to operation
func MethodToOperation(method string) string {
	switch method {
	case http.MethodPost:
		return createOp
	case http.MethodDelete:
		return deleteOp
	case http.MethodPut:
		return updateOp
	}
	return ""
}
