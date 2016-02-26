/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package models

import (
	"time"
)

// Notification holds all events.
type Notification struct {
	Events []Event
}

// Event holds the details of a event.
type Event struct {
	ID        string `json:"Id"`
	TimeStamp time.Time
	Action    string
	Target    *Target
	Request   *Request
	Actor     *Actor
}

// Target holds information about the target of a event.
type Target struct {
	MediaType  string
	Digest     string
	Repository string
	URL        string `json:"Url"`
}

// Actor holds information about actor.
type Actor struct {
	Name string
}

// Request holds information about a request.
type Request struct {
	ID        string `json:"Id"`
	Method    string
	UserAgent string
}
