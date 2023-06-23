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
	"time"

	cdevents "github.com/cdevents/sdk-go/pkg/api"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	purl "github.com/package-url/packageurl-go"

	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

var (
	// cloudEventsFormatter is the global single formatter for CloudEvents.
	cdEventsFormatter Formatter = &CDEventsFormatter{}
)

func init() {
	registerFormats(CDEventsFormat, cdEventsFormatter)
}

const (
	// CDEventsFormat is the type for CDEvents format.
	CDEventsFormat = "CDEvents"
)

// cdEventFromHarborType
func cdEventFromHarborType(he *model.HookEvent) (cdevents.CDEvent, error) {
	// Only TopicPullArtifact is supported for now.
	// Other event types are not striclty a failure (they are missing by design),
	// but we still report them upstream as unknown types
	if he.EventType != event.TopicPushArtifact {
		return nil, errors.Errorf("unknown event type: %s", he.EventType)
	}
	if len(he.Payload.EventData.Resources) == 0 {
		return nil, errors.Errorf("no resources defined in the event data: %v", he.Payload.EventData)
	}
	cde, err := cdevents.NewArtifactPublishedEvent()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to produce cdevent for event type %s", he.EventType)
	}

	// Setting CDEvents context attributes
	cde.SetId(uuid.NewString())
	// Use the same source format as for cloud events.
	// The source is project specific, but it doesn't specify that it's an harbor registry.
	// In future we may want to include this information somewhere, so that consumers
	// may be able to parse the custom data if they want to.
	// CDEvents does not support yet a custom data schema of any kind.
	cde.SetSource(source(he.ProjectID, he.PolicyID))
	cde.SetTimestamp(time.Unix(he.Payload.OccurAt, 0))

	// Setting CDEvents subject attributes
	// We cannot handle multiple resources for CDEvents
	// Only use the first one for the subject, the rest go in customData
	resource := he.Payload.EventData.Resources[0]
	// Build the subject ID as a pURL of OCI type: https://github.com/package-url/purl-spec/blob/master/PURL-TYPES.rst#oci
	// According to the spec, the namespace is included in the repository_url only
	qualifiersMap := map[string]string{
		"repository_url": resource.ResourceURL,
		"tag":            resource.Tag,
	}
	cde.SetSubjectId(purl.NewPackageURL(purl.TypeOCI, "", he.Payload.EventData.Repository.Name, resource.Digest, purl.QualifiersFromMap(qualifiersMap), "").ToString())

	// Setting CDEvents customData
	cde.SetCustomData(cloudevents.ApplicationJSON, he.Payload.EventData)
	return cde, nil
}

// CDEventsFormatter is the instance for the CDEvents format.
type CDEventsFormatter struct{}

// Format implements the interface Formatter.
/*
{
   "specversion":"1.0",
   "requestid": "2eedfab8-61d3-4f3c-8ec3-8f82d1ec4c84",
   "id":"4b2f89a6-548d-4c12-9993-a1f5790b97d2",
   "source":"/projects/1/webhook/policies/3",
   "type":"harbor.artifact.pushed",
   "datacontenttype":"application/json",
   "time":"2023-03-06T06:08:43Z",
   "data":{
		"context": {
			"version": "0.3.0",
			"id": "2d47734c-83e2-41dc-bcbf-23bd2a3734a2",
			"source": "/projects/1/webhook/policies/3",
			"type": "dev.cdevents.artifact.published.0.1.1",
			"timestamp": "2023-03-06T00:08:43-06:00"
		},
		"subject": {
			"id": "pkg:oci/busybox@sha256:dde8e930c7b6a490f728e66292bc9bce42efc9bbb5278bae40e4f30f6e00fe8c?repository_url=harbor.dev%2Flibrary%2Fbusybox:v1.0@sha256:dde8e930c7b6a490f728e66292bc9bce42efc9bbb5278bae40e4f30f6e00fe8c&tag=v1.0",
			"source": "/projects/1/webhook/policies/3",
			"type": "artifact",
			"content": {}
		},
		"customData": {
			"resources": [
			{
				"digest": "sha256:dde8e930c7b6a490f728e66292bc9bce42efc9bbb5278bae40e4f30f6e00fe8c",
				"tag": "v1.0",
				"resource_url": "harbor.dev/library/busybox:v1.0@sha256:dde8e930c7b6a490f728e66292bc9bce42efc9bbb5278bae40e4f30f6e00fe8c"
			}
			],
			"repository": {
			"date_created": 1677053165,
			"name": "busybox",
			"namespace": "library",
			"repo_full_name": "library/busybox",
			"repo_type": "public"
			}
		},
		"customDataContentType": "application/json"
   },
   "operator":"robot$library+scanner-Trivy-51fe4548-bbe5-11ed-9217-0242ac14000d"
}
*/
func (cde *CDEventsFormatter) Format(ctx context.Context, he *model.HookEvent) (http.Header, []byte, error) {
	if he == nil {
		return nil, nil, errors.Errorf("HookEvent should not be nil")
	}

	cdEvent, err := cdEventFromHarborType(he)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error preparing the CDEvent")
	}
	event, err := cdevents.AsCloudEvent(cdEvent)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error rending the CDEvent as CloudEvent")
	}

	// Set Harbor extensions
	event.SetExtension(extRequestID, lib.GetXRequestID(ctx))
	event.SetExtension(extOperator, he.Payload.Operator)

	// Set the CloudEvent timestamp. CDEvent sets its timestamp to that available in
	// the HookEvent, but it does not enforce a timestamp in the CloudEvent
	event.SetTime(time.Unix(he.Payload.OccurAt, 0))

	// Set the ID at CloudEvents level to match that of the CDEvents
	// This ideally should be handled by the CDEvents SDK, see https://github.com/cdevents/sdk-go/issues/57
	event.SetID(cdEvent.GetId())

	data, err := json.Marshal(event)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error to marshal CloudEvents")
	}

	header := http.Header{
		"Content-Type": []string{cloudevents.ApplicationCloudEventsJSON},
	}
	return header, data, nil
}
