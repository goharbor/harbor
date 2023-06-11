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

package bundle

import "github.com/sigstore/rekor/pkg/generated/models"

// RekorBundle holds metadata about recording a Signature's ephemeral key to
// a Rekor transparency log.
type RekorBundle struct {
	SignedEntryTimestamp []byte
	Payload              RekorPayload
}

type RekorPayload struct {
	Body           interface{} `json:"body"`
	IntegratedTime int64       `json:"integratedTime"`
	LogIndex       int64       `json:"logIndex"`
	LogID          string      `json:"logID"`
}

func EntryToBundle(entry *models.LogEntryAnon) *RekorBundle {
	if entry.Verification == nil {
		return nil
	}
	return &RekorBundle{
		SignedEntryTimestamp: entry.Verification.SignedEntryTimestamp,
		Payload: RekorPayload{
			Body:           entry.Body,
			IntegratedTime: *entry.IntegratedTime,
			LogIndex:       *entry.LogIndex,
			LogID:          *entry.LogID,
		},
	}
}
