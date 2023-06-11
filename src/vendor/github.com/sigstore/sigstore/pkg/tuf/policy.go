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

// Contains root policy definitions.
// Eventually, this will move this to go-tuf definitions.

package tuf

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	cjson "github.com/secure-systems-lab/go-securesystemslib/cjson"
)

type Signed struct {
	Signed     json.RawMessage `json:"signed"`
	Signatures []Signature     `json:"signatures"`
}

type Signature struct {
	KeyID     string `json:"keyid"`
	Signature string `json:"sig"`
	Cert      string `json:"cert,omitempty"`
}

type Key struct {
	Type       string          `json:"keytype"`
	Scheme     string          `json:"scheme"`
	Algorithms []string        `json:"keyid_hash_algorithms,omitempty"`
	Value      json.RawMessage `json:"keyval"`

	id     string
	idOnce sync.Once
}

func (k *Key) ID() string {
	k.idOnce.Do(func() {
		data, _ := cjson.EncodeCanonical(k)
		digest := sha256.Sum256(data)
		k.id = hex.EncodeToString(digest[:])
	})
	return k.id
}

func (k *Key) ContainsID(id string) bool {
	return id == k.ID()
}

type Root struct {
	Type        string           `json:"_type"`
	SpecVersion string           `json:"spec_version"`
	Version     int              `json:"version"`
	Expires     time.Time        `json:"expires"`
	Keys        map[string]*Key  `json:"keys"`
	Roles       map[string]*Role `json:"roles"`
	Namespace   string           `json:"namespace"`

	ConsistentSnapshot bool `json:"consistent_snapshot"`
}

func NewRoot() *Root {
	return &Root{
		Type:        "root",
		SpecVersion: "1.0",
		Version:     1,
		// Default expires in 3 months
		Expires:            time.Now().AddDate(0, 3, 0).UTC().Round(time.Second),
		Keys:               make(map[string]*Key),
		Roles:              make(map[string]*Role),
		ConsistentSnapshot: true,
	}
}

func (r *Root) AddKey(key *Key) bool {
	changed := false
	if _, ok := r.Keys[key.ID()]; !ok {
		changed = true
		r.Keys[key.ID()] = key
	}

	return changed
}

type Role struct {
	KeyIDs    []string `json:"keyids"`
	Threshold int      `json:"threshold"`
}

func (r *Role) AddKeysWithThreshold(keys []*Key, threshold int) bool {
	roleIDs := make(map[string]struct{})
	for _, id := range r.KeyIDs {
		roleIDs[id] = struct{}{}
	}
	changed := false
	for _, key := range keys {
		if _, ok := roleIDs[key.ID()]; !ok {
			changed = true
			r.KeyIDs = append(r.KeyIDs, key.ID())
		}
	}
	r.Threshold = threshold
	return changed
}

func (r *Root) Marshal() (*Signed, error) {
	// Marshals the Root into a Signed type
	b, err := cjson.EncodeCanonical(r)
	if err != nil {
		return nil, err
	}
	return &Signed{Signed: b}, nil
}

func (r *Root) ValidKey(key *Key, role string) (string, error) {
	// Checks if id is a valid key for role by matching the identity and issuer if specified.
	// Returns the key ID or an error if invalid key.
	fulcioKeyVal, err := GetFulcioKeyVal(key)
	if err != nil {
		return "", fmt.Errorf("error parsing signer key: %w", err)
	}

	result := ""
	for keyid, rootKey := range r.Keys {
		fulcioRootKeyVal, err := GetFulcioKeyVal(rootKey)
		if err != nil {
			return "", fmt.Errorf("error parsing root key: %w", err)
		}
		if fulcioKeyVal.Identity == fulcioRootKeyVal.Identity {
			if fulcioRootKeyVal.Issuer == "" || fulcioRootKeyVal.Issuer == fulcioKeyVal.Issuer {
				result = keyid
				break
			}
		}
	}
	if result == "" {
		return "", errors.New("key not found in root keys")
	}

	rootRole, ok := r.Roles[role]
	if !ok {
		return "", errors.New("invalid role")
	}
	for _, id := range rootRole.KeyIDs {
		if id == result {
			return result, nil
		}
	}
	return "", errors.New("key not found in role")
}

func (s *Signed) JSONMarshal(prefix, indent string) ([]byte, error) {
	// Marshals Signed with prefix and indent.
	b, err := cjson.EncodeCanonical(s)
	if err != nil {
		return []byte{}, err
	}

	var out bytes.Buffer
	if err := json.Indent(&out, b, prefix, indent); err != nil {
		return []byte{}, err
	}

	return out.Bytes(), nil
}

func (s *Signed) AddOrUpdateSignature(key *Key, signature Signature) error {
	root := &Root{}
	if err := json.Unmarshal(s.Signed, root); err != nil {
		return fmt.Errorf("unmarshalling root policy: %w", err)
	}
	var err error
	signature.KeyID, err = root.ValidKey(key, "root")
	if err != nil {
		return errors.New("invalid root key")
	}
	signatures := []Signature{}
	for _, sig := range s.Signatures {
		if sig.KeyID != signature.KeyID {
			signatures = append(signatures, sig)
		}
	}
	signatures = append(signatures, signature)
	s.Signatures = signatures
	return nil
}
