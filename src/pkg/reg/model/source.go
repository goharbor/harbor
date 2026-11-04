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
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/goharbor/harbor/src/lib/errors"
)

const (
	// FilterTypeFile selects the files a source adapter packs into an artifact. Only sources that
	// list it in RegistryInfo.SupportedResourceFilters accept it.
	FilterTypeFile = "file"

	maxFilePatterns = 64
)

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
}

// IsEmpty reports whether the options change nothing.
func (o *SourceOptions) IsEmpty() bool {
	return o == nil || o.FileFilter.IsEmpty()
}
