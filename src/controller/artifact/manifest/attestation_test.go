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
	"fmt"
	"strings"
	"testing"

	"github.com/docker/distribution/manifest/schema2"
	digest "github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/testing/mock"
	tart "github.com/goharbor/harbor/src/testing/pkg/artifact"
	tregistry "github.com/goharbor/harbor/src/testing/pkg/registry"
)

func TestIsAttestationDescriptor(t *testing.T) {
	tests := []struct {
		name       string
		descriptor v1.Descriptor
		want       bool
	}{
		{
			name: "OCI manifest with attestation annotation",
			descriptor: v1.Descriptor{
				MediaType: v1.MediaTypeImageManifest,
				Annotations: map[string]string{
					referenceTypeAnnotation: attestationManifestType,
				},
			},
			want: true,
		},
		{
			name: "Docker schema2 manifest with attestation annotation",
			descriptor: v1.Descriptor{
				MediaType: schema2.MediaTypeManifest,
				Annotations: map[string]string{
					referenceTypeAnnotation: attestationManifestType,
				},
			},
			want: true,
		},
		{
			name: "OCI manifest without annotation",
			descriptor: v1.Descriptor{
				MediaType: v1.MediaTypeImageManifest,
			},
			want: false,
		},
		{
			name: "wrong annotation value",
			descriptor: v1.Descriptor{
				MediaType: v1.MediaTypeImageManifest,
				Annotations: map[string]string{
					referenceTypeAnnotation: "not-attestation",
				},
			},
			want: false,
		},
		{
			name: "OCI index with attestation annotation",
			descriptor: v1.Descriptor{
				MediaType: v1.MediaTypeImageIndex,
				Annotations: map[string]string{
					referenceTypeAnnotation: attestationManifestType,
				},
			},
			want: false,
		},
		{
			name:       "empty descriptor",
			descriptor: v1.Descriptor{},
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isAttestationDescriptor(tt.descriptor))
		})
	}
}

func TestPlatformChildren(t *testing.T) {
	amd64 := v1.Descriptor{
		Digest:   digest.FromString("amd64"),
		Platform: &v1.Platform{OS: "linux", Architecture: "amd64"},
	}
	arm64 := v1.Descriptor{
		Digest:   digest.FromString("arm64"),
		Platform: &v1.Platform{OS: "linux", Architecture: "arm64"},
	}
	attestation := v1.Descriptor{
		MediaType: v1.MediaTypeImageManifest,
		Digest:    digest.FromString("attestation"),
		Annotations: map[string]string{
			referenceTypeAnnotation: attestationManifestType,
		},
	}

	children := platformChildren([]v1.Descriptor{amd64, attestation, arm64})
	assert.Len(t, children, 2)
	assert.Equal(t, amd64.Digest, children[0].Digest)
	assert.Equal(t, arm64.Digest, children[1].Digest)

	// all attestations
	children = platformChildren([]v1.Descriptor{attestation})
	assert.Empty(t, children)

	// empty input
	children = platformChildren(nil)
	assert.Empty(t, children)
}

func TestSubjectDigests(t *testing.T) {
	t.Run("valid sha256 digest", func(t *testing.T) {
		subject := inTotoSubject{
			Name: "amd64",
			Digest: map[string]string{
				"sha256": "cad250bb95ea402adf4f687cc7d6747ecf0de875e6d6117f74437893964903df",
			},
		}
		digests := subjectDigests(subject)
		assert.Equal(t, []string{"sha256:cad250bb95ea402adf4f687cc7d6747ecf0de875e6d6117f74437893964903df"}, digests)
	})

	t.Run("empty encoded value is skipped", func(t *testing.T) {
		subject := inTotoSubject{
			Digest: map[string]string{"sha256": ""},
		}
		digests := subjectDigests(subject)
		assert.Empty(t, digests)
	})

	t.Run("invalid digest is skipped", func(t *testing.T) {
		subject := inTotoSubject{
			Digest: map[string]string{"sha256": "too-short"},
		}
		digests := subjectDigests(subject)
		assert.Empty(t, digests)
	})

	t.Run("empty digest map", func(t *testing.T) {
		subject := inTotoSubject{}
		digests := subjectDigests(subject)
		assert.Empty(t, digests)
	})
}

func TestPlatformMatchesName(t *testing.T) {
	tests := []struct {
		name     string
		platform *v1.Platform
		input    string
		want     bool
	}{
		{
			name:     "match architecture only",
			platform: &v1.Platform{OS: "linux", Architecture: "amd64"},
			input:    "amd64",
			want:     true,
		},
		{
			name:     "match os/architecture",
			platform: &v1.Platform{OS: "linux", Architecture: "amd64"},
			input:    "linux/amd64",
			want:     true,
		},
		{
			name:     "match os/architecture/variant",
			platform: &v1.Platform{OS: "linux", Architecture: "arm", Variant: "v7"},
			input:    "linux/arm/v7",
			want:     true,
		},
		{
			name:     "case insensitive",
			platform: &v1.Platform{OS: "linux", Architecture: "amd64"},
			input:    "AMD64",
			want:     true,
		},
		{
			name:     "no match",
			platform: &v1.Platform{OS: "linux", Architecture: "amd64"},
			input:    "arm64",
			want:     false,
		},
		{
			name:     "nil platform",
			platform: nil,
			input:    "amd64",
			want:     false,
		},
		{
			name:     "empty architecture",
			platform: &v1.Platform{OS: "linux"},
			input:    "amd64",
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, platformMatchesName(tt.platform, tt.input))
		})
	}
}

func TestDigestBySubjectName(t *testing.T) {
	amd64Digest := digest.FromString("amd64-content")
	arm64Digest := digest.FromString("arm64-content")

	siblings := []v1.Descriptor{
		{
			Digest:   amd64Digest,
			Platform: &v1.Platform{OS: "linux", Architecture: "amd64"},
		},
		{
			Digest:   arm64Digest,
			Platform: &v1.Platform{OS: "linux", Architecture: "arm64"},
		},
	}

	t.Run("match by architecture name", func(t *testing.T) {
		got := digestBySubjectName(siblings, "amd64")
		assert.Equal(t, amd64Digest.String(), got)
	})

	t.Run("match by os/arch name", func(t *testing.T) {
		got := digestBySubjectName(siblings, "linux/arm64")
		assert.Equal(t, arm64Digest.String(), got)
	})

	t.Run("empty name returns empty", func(t *testing.T) {
		got := digestBySubjectName(siblings, "")
		assert.Empty(t, got)
	})

	t.Run("whitespace-only name returns empty", func(t *testing.T) {
		got := digestBySubjectName(siblings, "   ")
		assert.Empty(t, got)
	})

	t.Run("no match returns empty", func(t *testing.T) {
		got := digestBySubjectName(siblings, "s390x")
		assert.Empty(t, got)
	})

	t.Run("ambiguous match returns empty", func(t *testing.T) {
		// Two siblings with same architecture
		dupes := []v1.Descriptor{
			{
				Digest:   digest.FromString("first"),
				Platform: &v1.Platform{OS: "linux", Architecture: "amd64"},
			},
			{
				Digest:   digest.FromString("second"),
				Platform: &v1.Platform{OS: "linux", Architecture: "amd64"},
			},
		}
		got := digestBySubjectName(dupes, "amd64")
		assert.Empty(t, got)
	})
}

func TestResolveSubjectFromStatement(t *testing.T) {
	amd64Encoded := "cad250bb95ea402adf4f687cc7d6747ecf0de875e6d6117f74437893964903df"
	arm64Encoded := "480b518ed0138eacf2d070de80cb8eb019fb0b3565e2598ed654a541c31061a0"
	amd64Ref := "sha256:" + amd64Encoded
	arm64Ref := "sha256:" + arm64Encoded

	children := []v1.Descriptor{
		{Digest: digest.Digest(amd64Ref), Platform: &v1.Platform{OS: "linux", Architecture: "amd64"}},
		{Digest: digest.Digest(arm64Ref), Platform: &v1.Platform{OS: "linux", Architecture: "arm64"}},
	}

	t.Run("subject digest matches a child", func(t *testing.T) {
		subjects := []inTotoSubject{{Name: "amd64", Digest: map[string]string{"sha256": amd64Encoded}}}
		assert.Equal(t, amd64Ref, resolveSubjectFromStatement(children, subjects))
	})

	t.Run("falls back to the subject name", func(t *testing.T) {
		subjects := []inTotoSubject{{Name: "linux/arm64", Digest: map[string]string{"sha256": "not-a-digest"}}}
		assert.Equal(t, arm64Ref, resolveSubjectFromStatement(children, subjects))
	})

	t.Run("no subjects resolves to nothing", func(t *testing.T) {
		assert.Empty(t, resolveSubjectFromStatement(children, nil))
	})

	t.Run("unknown subject resolves to nothing", func(t *testing.T) {
		subjects := []inTotoSubject{{Name: "s390x"}}
		assert.Empty(t, resolveSubjectFromStatement(children, subjects))
	})

	// A statement naming two different children does not identify a single
	// target, so nothing is attached rather than taking whichever came first.
	t.Run("two subjects matching different children is ambiguous", func(t *testing.T) {
		subjects := []inTotoSubject{
			{Name: "amd64", Digest: map[string]string{"sha256": amd64Encoded}},
			{Name: "arm64", Digest: map[string]string{"sha256": arm64Encoded}},
		}
		assert.Empty(t, resolveSubjectFromStatement(children, subjects))
	})

	t.Run("ambiguous names resolve to nothing", func(t *testing.T) {
		subjects := []inTotoSubject{{Name: "amd64"}, {Name: "arm64"}}
		assert.Empty(t, resolveSubjectFromStatement(children, subjects))
	})

	// The digest map iterates in unspecified order, so the same subject naming
	// two children must not resolve differently between runs.
	t.Run("ambiguous digests within one subject are order independent", func(t *testing.T) {
		sha512Encoded := strings.Repeat("ab", 64)
		withSha512 := append(children, v1.Descriptor{
			Digest:   digest.Digest("sha512:" + sha512Encoded),
			Platform: &v1.Platform{OS: "linux", Architecture: "s390x"},
		})
		subjects := []inTotoSubject{{Digest: map[string]string{"sha256": amd64Encoded, "sha512": sha512Encoded}}}
		for range 50 {
			assert.Empty(t, resolveSubjectFromStatement(withSha512, subjects))
		}
	})
}

// syntheticManifests builds index children of arbitrary width. Unresolvable
// attestations name a subject outside the index, which is what forces the
// payload lookup that the budget has to bound.
func syntheticManifests(platforms, attestations int, resolvable bool) []v1.Descriptor {
	digestOf := func(i int) digest.Digest { return digest.Digest(fmt.Sprintf("sha256:%064x", i)) }

	manifests := make([]v1.Descriptor, 0, platforms+attestations)
	for i := range platforms {
		manifests = append(manifests, v1.Descriptor{
			MediaType: v1.MediaTypeImageManifest,
			Size:      7143,
			Digest:    digestOf(i),
			Platform:  &v1.Platform{OS: "linux", Architecture: fmt.Sprintf("arch%d", i)},
		})
	}
	for i := range attestations {
		subject := digestOf(-1 - i)
		if resolvable {
			subject = digestOf(i % platforms)
		}
		manifests = append(manifests, v1.Descriptor{
			MediaType: v1.MediaTypeImageManifest,
			Size:      1024,
			Digest:    digestOf(platforms + i),
			Annotations: map[string]string{
				referenceTypeAnnotation:   attestationManifestType,
				referenceDigestAnnotation: subject.String(),
			},
			Platform: &v1.Platform{OS: "unknown", Architecture: "unknown"},
		})
	}
	return manifests
}

// An index whose attestations all name a subject outside it must not turn one
// push into a registry round trip per descriptor.
func TestClassifyBoundsSubjectLookups(t *testing.T) {
	for _, tc := range []struct {
		name            string
		platforms       int
		attestations    int
		expectedLookups int
	}{
		{"budget follows the child count", 1, 40, 1},
		{"budget stops at the ceiling", 64, 100, maxSubjectLookups},
	} {
		t.Run(tc.name, func(t *testing.T) {
			regCli := &tregistry.Client{}
			regCli.On("PullManifest", mock.Anything, mock.Anything).Return(nil, "", fmt.Errorf("boom"))

			manifests := syntheticManifests(tc.platforms, tc.attestations, false)
			references, candidates, err := NewInTotoAttestationClassifier(&tart.Manager{}, regCli).
				Classify(context.Background(), "library/test", manifests)

			require.NoError(t, err)
			regCli.AssertNumberOfCalls(t, "PullManifest", tc.expectedLookups)
			regCli.AssertNotCalled(t, "PullBlob", mock.Anything, mock.Anything)
			assert.Empty(t, candidates)
			assert.Len(t, references, tc.platforms+tc.attestations,
				"unresolved attestations stay children instead of being dropped")
		})
	}
}
