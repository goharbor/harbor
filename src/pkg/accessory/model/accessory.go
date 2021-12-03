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

package model

import (
	"fmt"
	"sync"
	"time"
)

const (
	// RefNone identifies base reference
	RefNone = "base"
	// RefSoft identifies soft reference
	RefSoft = "soft"
	// RefHard identifies hard reference
	RefHard = "hard"
)

type RefProvider interface {
	// Kind returns reference Kind.
	Kind() string
}

/*
Soft reference: The accessory is not tied to the subject manifest.
Hard reference: The accessory is tied to the subject manifest.

	Deletion
1. Soft Reference:  If the linkage is Soft Reference, when the subject artifact is removed, the linkage will be removed as well, the accessory artifact becomes an individual artifact.
2. Hard Reference:  If the linkage is Hard Reference, the accessory artifact will be removed together with the subject artifact.

	Garbage Collection
1. Soft Reference:  If the linkage is Soft Reference, Harbor treats the accessory as normal artifact and will not set it as the GC candidate.
2. Hard Reference:  If the linkage is Hard Reference, Harbor treats the accessory as an extra stuff of the subject artifact. It means, it being tied to the subject artifact and will be GCed whenever the subject artifact is marked and deleted.
*/
type RefIdentifier interface {
	// IsSoft indicates that the linkage of artifact and its accessory is soft reference.
	IsSoft() bool

	// IsHard indicates that the linkage of artifact and its accessory is hard reference.
	IsHard() bool
}

const (
	// TypeNone
	TypeNone = "base"
	// TypeCosignSignature ...
	TypeCosignSignature = "signature.cosign"
)

// AccessoryData ...
type AccessoryData struct {
	ID            int64
	ArtifactID    int64
	SubArtifactID int64
	Type          string
	Size          int64
	Digest        string
	CreatTime     time.Time
}

// Accessory: Independent, but linked to an existing subject artifact, which enabling the extendibility of an OCI artifact.
type Accessory interface {
	RefProvider
	RefIdentifier
	GetData() AccessoryData
}

// NewAccessoryFunc takes data to return a Accessory.
type NewAccessoryFunc func(data AccessoryData) Accessory

var (
	factories = map[string]NewAccessoryFunc{}
	lock      sync.RWMutex
)

// Register register accessory factory for type
func Register(typ string, factory NewAccessoryFunc) {
	lock.Lock()
	defer lock.Unlock()

	factories[typ] = factory
}

// New returns accessory instance
func New(typ string, data AccessoryData) (Accessory, error) {
	lock.Lock()
	defer lock.Unlock()

	factory, ok := factories[typ]
	if !ok {
		return nil, fmt.Errorf("accessory type %s not support", typ)
	}

	data.Type = typ
	return factory(data), nil
}
