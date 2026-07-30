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

package manifest

import (
	"context"
	"errors"
	"testing"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/artifact"
)

type stubClassifier struct {
	candidate *artifact.AccessoryCandidate
	err       error
	calls     *int
}

func (s stubClassifier) Classify(_ context.Context, _ string, _ v1.Descriptor, _ []v1.Descriptor) (*artifact.AccessoryCandidate, error) {
	if s.calls != nil {
		*s.calls++
	}
	return s.candidate, s.err
}

func withClassifiers(t *testing.T, classifiers ...ChildClassifier) {
	t.Helper()
	original := ChildClassifiers
	ChildClassifiers = classifiers
	t.Cleanup(func() { ChildClassifiers = original })
}

// An empty chain is silent data loss, not a visible failure, so pin the bootstrap.
func TestDefaultChainIsRegistered(t *testing.T) {
	require.NotEmpty(t, ChildClassifiers, "the package init() must register at least one classifier")

	var attestation *InTotoAttestationClassifier
	for _, classifier := range ChildClassifiers {
		if c, ok := classifier.(*InTotoAttestationClassifier); ok {
			attestation = c
			break
		}
	}
	require.NotNil(t, attestation, "the in-toto attestation classifier must be registered by default")

	// init() captures these globals, so it relies on their packages initialising first
	assert.NotNil(t, attestation.artMgr, "pkg.ArtifactMgr must be initialised before this package's init()")
	assert.NotNil(t, attestation.regCli, "registry.Cli must be initialised before this package's init()")
}

func TestClassifyChildNoClassifierClaims(t *testing.T) {
	calls := 0
	withClassifiers(t, stubClassifier{calls: &calls}, stubClassifier{calls: &calls})

	candidate, err := ClassifyChild(context.Background(), "library/hello", v1.Descriptor{}, nil)
	require.NoError(t, err)
	assert.Nil(t, candidate)
	assert.Equal(t, 2, calls, "every classifier is consulted when none claims the descriptor")
}

func TestClassifyChildFirstMatchWins(t *testing.T) {
	second := 0
	first := &artifact.AccessoryCandidate{SubArtifactDigest: "sha256:first"}
	withClassifiers(t,
		stubClassifier{candidate: first},
		stubClassifier{candidate: &artifact.AccessoryCandidate{SubArtifactDigest: "sha256:second"}, calls: &second},
	)

	candidate, err := ClassifyChild(context.Background(), "library/hello", v1.Descriptor{}, nil)
	require.NoError(t, err)
	require.NotNil(t, candidate)
	assert.Equal(t, "sha256:first", candidate.SubArtifactDigest)
	assert.Zero(t, second, "classifiers after the first match are not consulted")
}

func TestClassifyChildPropagatesError(t *testing.T) {
	notReached := 0
	wantErr := errors.New("boom")
	withClassifiers(t,
		stubClassifier{err: wantErr},
		stubClassifier{calls: &notReached},
	)

	candidate, err := ClassifyChild(context.Background(), "library/hello", v1.Descriptor{}, nil)
	assert.ErrorIs(t, err, wantErr)
	assert.Nil(t, candidate)
	assert.Zero(t, notReached, "the chain stops at the first error")
}

func TestClassifyChildEmptyChain(t *testing.T) {
	withClassifiers(t)

	candidate, err := ClassifyChild(context.Background(), "library/hello", v1.Descriptor{}, nil)
	require.NoError(t, err)
	assert.Nil(t, candidate)
}

func TestRegisterChildClassifier(t *testing.T) {
	withClassifiers(t)

	RegisterChildClassifier(stubClassifier{})
	RegisterChildClassifier(stubClassifier{})
	assert.Len(t, ChildClassifiers, 2)
}

func TestRegisterChildClassifierDropsNil(t *testing.T) {
	withClassifiers(t)

	RegisterChildClassifier(nil)
	assert.Empty(t, ChildClassifiers, "a nil classifier would panic on every index push")

	candidate, err := ClassifyChild(context.Background(), "library/hello", v1.Descriptor{}, nil)
	require.NoError(t, err)
	assert.Nil(t, candidate)
}
