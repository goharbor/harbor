//
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
	"encoding/json"
)

const (
	KeyTypeFulcio   = "sigstore-oidc"
	KeySchemeFulcio = "https://fulcio.sigstore.dev"
)

var KeyAlgorithms = []string{"sha256", "sha512"}

type FulcioKeyVal struct {
	Identity string `json:"identity"`
	Issuer   string `json:"issuer,omitempty"`
}

func FulcioVerificationKey(email, issuer string) *Key {
	keyValBytes, _ := json.Marshal(FulcioKeyVal{Identity: email, Issuer: issuer})
	return &Key{
		Type:       KeyTypeFulcio,
		Scheme:     KeySchemeFulcio,
		Algorithms: KeyAlgorithms,
		Value:      keyValBytes,
	}
}

func GetFulcioKeyVal(key *Key) (*FulcioKeyVal, error) {
	fulcioKeyVal := &FulcioKeyVal{}
	err := json.Unmarshal(key.Value, fulcioKeyVal)
	return fulcioKeyVal, err
}
