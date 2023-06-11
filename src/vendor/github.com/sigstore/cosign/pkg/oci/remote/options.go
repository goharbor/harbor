//
// Copyright 2021 The Sigstore Authors.
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

package remote

import (
	"fmt"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

const (
	SignatureTagSuffix   = "sig"
	SBOMTagSuffix        = "sbom"
	AttestationTagSuffix = "att"
	CustomTagPrefix      = ""

	RepoOverrideEnvKey = "COSIGN_REPOSITORY"
)

// Option is a functional option for remote operations.
type Option func(*options)

type options struct {
	SignatureSuffix   string
	AttestationSuffix string
	SBOMSuffix        string
	TagPrefix         string
	TargetRepository  name.Repository
	ROpt              []remote.Option

	OriginalOptions []Option
}

var defaultOptions = []remote.Option{
	remote.WithAuthFromKeychain(authn.DefaultKeychain),
	// TODO(mattmoor): Incorporate user agent.
}

func makeOptions(target name.Repository, opts ...Option) *options {
	o := &options{
		SignatureSuffix:   SignatureTagSuffix,
		AttestationSuffix: AttestationTagSuffix,
		SBOMSuffix:        SBOMTagSuffix,
		TagPrefix:         CustomTagPrefix,
		TargetRepository:  target,
		ROpt:              defaultOptions,

		// Keep the original options around for things that want
		// to call something that takes options!
		OriginalOptions: opts,
	}

	for _, option := range opts {
		option(o)
	}

	return o
}

// WithPrefix is a functional option for overriding the default
// tag prefix.
func WithPrefix(prefix string) Option {
	return func(o *options) {
		o.TagPrefix = prefix
	}
}

// WithSignatureSuffix is a functional option for overriding the default
// signature tag suffix.
func WithSignatureSuffix(suffix string) Option {
	return func(o *options) {
		o.SignatureSuffix = suffix
	}
}

// WithAttestationSuffix is a functional option for overriding the default
// attestation tag suffix.
func WithAttestationSuffix(suffix string) Option {
	return func(o *options) {
		o.AttestationSuffix = suffix
	}
}

// WithSBOMSuffix is a functional option for overriding the default
// SBOM tag suffix.
func WithSBOMSuffix(suffix string) Option {
	return func(o *options) {
		o.SBOMSuffix = suffix
	}
}

// WithRemoteOptions is a functional option for overriding the default
// remote options passed to GGCR.
func WithRemoteOptions(opts ...remote.Option) Option {
	return func(o *options) {
		o.ROpt = opts
	}
}

// WithTargetRepository is a functional option for overriding the default
// target repository hosting the signature and attestation tags.
func WithTargetRepository(repo name.Repository) Option {
	return func(o *options) {
		o.TargetRepository = repo
	}
}

// GetEnvTargetRepository returns the Repository specified by
// `os.Getenv(RepoOverrideEnvKey)`, or the empty value if not set.
// Returns an error if the value is set but cannot be parsed.
func GetEnvTargetRepository() (name.Repository, error) {
	if ro := os.Getenv(RepoOverrideEnvKey); ro != "" {
		repo, err := name.NewRepository(ro)
		if err != nil {
			return name.Repository{}, fmt.Errorf("parsing $"+RepoOverrideEnvKey+": %w", err)
		}
		return repo, nil
	}
	return name.Repository{}, nil
}
