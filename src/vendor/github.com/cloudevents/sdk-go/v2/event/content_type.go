/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event

const (
	TextPlain                       = "text/plain"
	TextJSON                        = "text/json"
	ApplicationJSON                 = "application/json"
	ApplicationXML                  = "application/xml"
	ApplicationCloudEventsJSON      = "application/cloudevents+json"
	ApplicationCloudEventsBatchJSON = "application/cloudevents-batch+json"
)

// StringOfApplicationJSON returns a string pointer to "application/json"
func StringOfApplicationJSON() *string {
	a := ApplicationJSON
	return &a
}

// StringOfApplicationXML returns a string pointer to "application/xml"
func StringOfApplicationXML() *string {
	a := ApplicationXML
	return &a
}

// StringOfTextPlain returns a string pointer to "text/plain"
func StringOfTextPlain() *string {
	a := TextPlain
	return &a
}

// StringOfApplicationCloudEventsJSON  returns a string pointer to
// "application/cloudevents+json"
func StringOfApplicationCloudEventsJSON() *string {
	a := ApplicationCloudEventsJSON
	return &a
}

// StringOfApplicationCloudEventsBatchJSON returns a string pointer to
// "application/cloudevents-batch+json"
func StringOfApplicationCloudEventsBatchJSON() *string {
	a := ApplicationCloudEventsBatchJSON
	return &a
}
