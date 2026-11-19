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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/goharbor/harbor/src/lib/errors"
)

const (
	// FilterTypeFile selects the files a source adapter packs into an artifact. Only sources that
	// list it in RegistryInfo.SupportedResourceFilters accept it.
	FilterTypeFile = "file"

	// RevisionTagPrefix starts the immutable tag of a source revision, see RevisionTag.
	RevisionTagPrefix = "sha-"
	revisionTagHexLen = 12

	maxFilePatterns = 64
)

var revisionTagRe = regexp.MustCompile(`^sha-[0-9a-f]{12}$`)

// RevisionTag returns the immutable tag of a source revision: "sha-" and the first 12 hex
// characters. It returns "" for a revision that does not start with 12 lowercase hex characters.
func RevisionTag(revision string) string {
	if len(revision) < revisionTagHexLen {
		return ""
	}
	tag := RevisionTagPrefix + revision[:revisionTagHexLen]
	if !revisionTagRe.MatchString(tag) {
		return ""
	}
	return tag
}

// IsRevisionTag reports whether tag has the form produced by RevisionTag.
func IsRevisionTag(tag string) bool {
	return revisionTagRe.MatchString(tag)
}

// FileFilter selects files of a source artifact by path. A path is selected when it matches
// any Include pattern (or Include is empty) and no Exclude pattern. Patterns use doublestar
// syntax against the full slash-separated path; a pattern without "/" matches the base name at
// any depth, so "*.json" selects "onnx/config.json" as well.
type FileFilter struct {
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
}

// ParseFileFilter reads the value of a file filter. The value is either an object
// {"include": [...], "exclude": [...]} or a string of patterns separated by commas, spaces or
// newlines where a leading "!" marks an exclude pattern; the string form is what the generic
// text input of the replication rule form produces. An empty value returns nil.
func ParseFileFilter(value any) (*FileFilter, error) {
	f := &FileFilter{}
	switch v := value.(type) {
	case nil:
		return nil, nil // nolint:nilnil // nil is the filter that selects every file
	case *FileFilter:
		if v == nil {
			return nil, nil // nolint:nilnil // nil is the filter that selects every file
		}
		*f = *v
	case FileFilter:
		*f = v
	case string:
		for _, p := range strings.FieldsFunc(v, func(r rune) bool {
			return r == ',' || r == ' ' || r == '\n' || r == '\t' || r == '\r'
		}) {
			if strings.HasPrefix(p, "!") {
				f.Exclude = append(f.Exclude, p[1:])
			} else {
				f.Include = append(f.Include, p)
			}
		}
	case map[string]any:
		data, err := json.Marshal(v)
		if err != nil {
			return nil, invalidFileFilter(err.Error())
		}
		dec := json.NewDecoder(strings.NewReader(string(data)))
		dec.DisallowUnknownFields()
		if err := dec.Decode(f); err != nil {
			return nil, invalidFileFilter(err.Error())
		}
	default:
		return nil, invalidFileFilter(fmt.Sprintf("unsupported value type %T", value))
	}
	if f.IsEmpty() {
		return nil, nil // nolint:nilnil // nil is the filter that selects every file
	}
	if err := f.Validate(); err != nil {
		return nil, err
	}
	return f, nil
}

// FileFilterOf returns the file filter among filters, or nil.
func FileFilterOf(filters []*Filter) (*FileFilter, error) {
	for _, f := range filters {
		if f != nil && f.Type == FilterTypeFile {
			return ParseFileFilter(f.Value)
		}
	}
	return nil, nil // nolint:nilnil // nil is the filter that selects every file
}

func invalidFileFilter(reason string) error {
	return errors.New(nil).WithCode(errors.BadRequestCode).WithMessagef("invalid file filter: %s", reason)
}

// IsEmpty reports whether the filter selects every file.
func (f *FileFilter) IsEmpty() bool {
	return f == nil || len(f.Include) == 0 && len(f.Exclude) == 0
}

// Validate checks the patterns.
func (f *FileFilter) Validate() error {
	if f == nil {
		return nil
	}
	if len(f.Include)+len(f.Exclude) > maxFilePatterns {
		return invalidFileFilter(fmt.Sprintf("more than %d patterns", maxFilePatterns))
	}
	for _, p := range append(append([]string{}, f.Include...), f.Exclude...) {
		if strings.TrimSpace(p) == "" {
			return invalidFileFilter("empty pattern")
		}
		if strings.HasPrefix(p, "/") {
			return invalidFileFilter(fmt.Sprintf("pattern %q must be relative", p))
		}
		if !doublestar.ValidatePattern(p) {
			return invalidFileFilter(fmt.Sprintf("malformed pattern %q", p))
		}
	}
	return nil
}

// Match reports whether the filter selects filePath. A nil filter selects everything.
func (f *FileFilter) Match(filePath string) bool {
	if f.IsEmpty() {
		return true
	}
	included := len(f.Include) == 0 || matchAnyPattern(f.Include, filePath)
	return included && !matchAnyPattern(f.Exclude, filePath)
}

func matchAnyPattern(patterns []string, filePath string) bool {
	for _, p := range patterns {
		target := filePath
		if !strings.Contains(p, "/") {
			target = path.Base(filePath)
		}
		if ok, err := doublestar.Match(p, target); err == nil && ok {
			return true
		}
	}
	return false
}

// Key returns a short stable identifier of the filter, independent of pattern order and
// duplicates, for cache keys. An empty filter returns "".
func (f *FileFilter) Key() string {
	if f.IsEmpty() {
		return ""
	}
	canonical := func(in []string) []string {
		seen := map[string]bool{}
		var out []string
		for _, p := range in {
			if !seen[p] {
				seen[p] = true
				out = append(out, p)
			}
		}
		sort.Strings(out)
		return out
	}
	data, _ := json.Marshal([2][]string{canonical(f.Include), canonical(f.Exclude)})
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:8])
}

// SourceOptions are policy settings that change what a source adapter produces for a copy.
// They travel with the source resource (Resource.SourceOptions) from the replication flow to
// the transfer, which hands them to the source adapter through adapter.SourceConfigurer.
type SourceOptions struct {
	// FileFilter limits the files packed into each artifact.
	FileFilter *FileFilter `json:"file_filter,omitempty"`
	// RevisionTag adds RevisionTag(revision) to the tags of every fetched artifact, so the copy
	// keeps an immutable tag next to the moving ones.
	RevisionTag bool `json:"revision_tag,omitempty"`
}

// IsEmpty reports whether the options change nothing.
func (o *SourceOptions) IsEmpty() bool {
	return o == nil || o.FileFilter.IsEmpty() && !o.RevisionTag
}

// File classes of ResolvedFile.Class, after the ModelPack model-spec layer media types.
const (
	FileClassWeight = "weight"
	FileClassConfig = "config"
	FileClassCode   = "code"
	FileClassDoc    = "doc"
)

// ResolveResult describes what a copy of one source reference would produce, see
// adapter.ArtifactResolver.
type ResolveResult struct {
	// Repository is the repository name the source adapter uses for the artifact, before any
	// destination namespace rewriting.
	Repository string `json:"repository"`
	// SourceID is the canonical identifier at the source, e.g. a Hugging Face model ID.
	SourceID  string `json:"source_id"`
	Reference string `json:"reference"`
	// Revision is the immutable source revision the reference resolves to, e.g. a git commit.
	Revision string `json:"revision"`
	// RevisionTag is the immutable tag a copy gets when SourceOptions.RevisionTag is set.
	RevisionTag string `json:"revision_tag,omitempty"`
	SourceURL   string `json:"source_url,omitempty"`
	// Tags are the moving tags of the source that point at Revision.
	Tags []string `json:"tags"`
	// Digest is the manifest digest a copy with the given filter produces; empty when no file matches.
	Digest       string     `json:"digest,omitempty"`
	Licenses     []string   `json:"licenses,omitempty"`
	Task         string     `json:"task,omitempty"`
	LastModified *time.Time `json:"last_modified,omitempty"`
	TotalFiles   int        `json:"total_files"`
	TotalSize    int64      `json:"total_size"`
	MatchedFiles int        `json:"matched_files"`
	MatchedSize  int64      `json:"matched_size"`
	// Files is sorted by path.
	Files []*ResolvedFile `json:"files"`
}

// ResolvedFile is one source file of a ResolveResult.
type ResolvedFile struct {
	Path  string `json:"path"`
	Size  int64  `json:"size"`
	Class string `json:"class"`
	// SHA256 is the hex sha256 of the content when known without downloading the file, or
	// when it was hashed to compute the digest.
	SHA256  string `json:"sha256,omitempty"`
	Matched bool   `json:"matched"`
}
