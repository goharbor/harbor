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

package hub

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

var commitRe = regexp.MustCompile(`^[0-9a-f]{40}$`)

// IsCommit reports whether s is a full 40-hex Hub commit sha.
func IsCommit(s string) bool {
	return commitRe.MatchString(s)
}

// Snapshot is the Hub file listing of one commit.
type Snapshot struct {
	// ModelID is the canonical model ID, case preserved.
	ModelID      string
	Commit       string
	LastModified time.Time
	Licenses     []string
	Files        []File
}

// File is one file of a snapshot.
type File struct {
	Path string
	Size int64
	// BlobID is the git blob sha1 of the file content (for LFS files, of the pointer).
	BlobID string
	// SHA256 is the hex sha256 of the content for LFS/Xet files, empty for non-LFS files.
	SHA256 string
}

// IsLFS reports whether the Hub publishes the sha256 of the file content.
func (f File) IsLFS() bool {
	return f.SHA256 != ""
}

// Ref is a Hub branch or tag.
type Ref struct {
	Name   string
	Commit string
}

// JSON field names below are pinned: the revision API uses siblings[].rfilename,
// blobId and lfs.sha256, unlike the tree API (path, oid, lfs.oid).
type revisionResponse struct {
	ID           string          `json:"id"`
	SHA          string          `json:"sha"`
	LastModified time.Time       `json:"lastModified"`
	CardData     *cardData       `json:"cardData"`
	Siblings     []siblingEntity `json:"siblings"`
}

type cardData struct {
	License json.RawMessage `json:"license"`
}

type siblingEntity struct {
	RFilename string     `json:"rfilename"`
	BlobID    string     `json:"blobId"`
	Size      *int64     `json:"size"`
	LFS       *lfsEntity `json:"lfs"`
}

type lfsEntity struct {
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

type refsResponse struct {
	Branches []refEntity `json:"branches"`
	Tags     []refEntity `json:"tags"`
}

type refEntity struct {
	Name         string `json:"name"`
	TargetCommit string `json:"targetCommit"`
}

type modelEntity struct {
	ID string `json:"id"`
}

type accountEntity struct {
	Name string `json:"name"`
}

var sha256Re = regexp.MustCompile(`^[0-9a-f]{64}$`)

// DecodeSnapshot parses a response of GET /api/models/{id}/revision/{rev}?blobs=true.
func DecodeSnapshot(data []byte) (*Snapshot, error) {
	var r revisionResponse
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	return r.toSnapshot()
}

func (r *revisionResponse) toSnapshot() (*Snapshot, error) {
	if r.ID == "" {
		return nil, fmt.Errorf("revision response has no model id")
	}
	if !IsCommit(r.SHA) {
		return nil, fmt.Errorf("revision response for %s has invalid commit %q", r.ID, r.SHA)
	}
	s := &Snapshot{
		ModelID:      r.ID,
		Commit:       r.SHA,
		LastModified: r.LastModified.UTC(),
	}
	if r.CardData != nil {
		s.Licenses = parseLicenses(r.CardData.License)
	}
	for _, sib := range r.Siblings {
		if sib.Size == nil {
			return nil, fmt.Errorf("model %s: file %s has no size, was blobs=true dropped?", r.ID, sib.RFilename)
		}
		f := File{Path: sib.RFilename, Size: *sib.Size, BlobID: sib.BlobID}
		if sib.LFS != nil {
			if !sha256Re.MatchString(sib.LFS.SHA256) {
				return nil, fmt.Errorf("model %s: file %s has invalid lfs sha256 %q", r.ID, sib.RFilename, sib.LFS.SHA256)
			}
			f.SHA256 = sib.LFS.SHA256
			f.Size = sib.LFS.Size
		}
		s.Files = append(s.Files, f)
	}
	return s, nil
}

// parseLicenses accepts the model card license as a string or a list of strings.
// Any other shape is free-form card data and is ignored rather than failing the snapshot.
func parseLicenses(raw json.RawMessage) []string {
	var one string
	if err := json.Unmarshal(raw, &one); err == nil {
		if one == "" {
			return nil
		}
		return []string{one}
	}
	var many []string
	if err := json.Unmarshal(raw, &many); err != nil {
		return nil
	}
	var out []string
	for _, l := range many {
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}
