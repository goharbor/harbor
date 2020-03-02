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

package blob

// Option option for `Get` and `Exist` method of `Controller`
type Option func(*Options)

// Options options used by `Get` method of `Controller`
type Options struct {
	ArtifactDigest string // blob associated with the artifact
	ProjectID      int64  // blob associated with the project
}

// IsAssociatedWithArtifact set ArtifactDigest for the Options
func IsAssociatedWithArtifact(artifactDigest string) Option {
	return func(opts *Options) {
		opts.ArtifactDigest = artifactDigest
	}
}

// IsAssociatedWithProject set ProjectID for the Options
func IsAssociatedWithProject(projectID int64) Option {
	return func(opts *Options) {
		opts.ProjectID = projectID
	}
}
