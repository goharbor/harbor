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

type UsageKind int

const (
	UnknownUsage UsageKind = iota
	Fulcio
	Rekor
	CTFE
)

var toUsageString = map[UsageKind]string{
	UnknownUsage: "Unknown",
	Fulcio:       "Fulcio",
	Rekor:        "Rekor",
	CTFE:         "CTFE",
}

func (u UsageKind) String() string {
	return toUsageString[u]
}

func (u UsageKind) MarshalText() ([]byte, error) {
	str := u.String()
	if len(str) == 0 {
		return nil, fmt.Errorf("error while marshalling, int(UsageKind)=%d not valid", int(u))
	}
	return []byte(u.String()), nil
}

func (u *UsageKind) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "unknown":
		*u = UnknownUsage
	case "fulcio":
		*u = Fulcio
	case "rekor":
		*u = Rekor
	case "ctfe":
		*u = CTFE
	default:
		return fmt.Errorf("error while unmarshalling, UsageKind=%v not valid", string(text))
	}
	return nil
}
