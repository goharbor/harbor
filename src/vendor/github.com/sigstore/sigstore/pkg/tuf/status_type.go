// Copyright 2022 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tuf

import (
	"fmt"
	"strings"
)

type StatusKind int

const (
	UnknownStatus StatusKind = iota
	Active
	Expired
)

var toStatusString = map[StatusKind]string{
	UnknownStatus: "Unknown",
	Active:        "Active",
	Expired:       "Expired",
}

func (s StatusKind) String() string {
	return toStatusString[s]
}

func (s StatusKind) MarshalText() ([]byte, error) {
	str := s.String()
	if len(str) == 0 {
		return nil, fmt.Errorf("error while marshalling, int(StatusKind)=%d not valid", int(s))
	}
	return []byte(s.String()), nil
}

func (s *StatusKind) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "unknown":
		*s = UnknownStatus
	case "active":
		*s = Active
	case "expired":
		*s = Expired
	default:
		return fmt.Errorf("error while unmarshalling, StatusKind=%v not valid", string(text))
	}
	return nil
}
