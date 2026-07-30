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
	"strings"
	"testing"

	"github.com/docker/distribution/manifest/schema2"
	digest "github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
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

func TestResolveAttestationSubject(t *testing.T) {
	amd64Digest := "sha256:cad250bb95ea402adf4f687cc7d6747ecf0de875e6d6117f74437893964903df"
	arm64Digest := "sha256:480b518ed0138eacf2d070de80cb8eb019fb0b3565e2598ed654a541c31061a0"

	amd64Child := v1.Descriptor{
		Digest:   digest.Digest(amd64Digest),
		Platform: &v1.Platform{OS: "linux", Architecture: "amd64"},
	}
	arm64Child := v1.Descriptor{
		Digest:   digest.Digest(arm64Digest),
		Platform: &v1.Platform{OS: "linux", Architecture: "arm64"},
	}
	attestation := v1.Descriptor{
		MediaType: v1.MediaTypeImageManifest,
		Digest:    digest.FromString("attestation"),
		Annotations: map[string]string{
			referenceTypeAnnotation:   attestationManifestType,
			referenceDigestAnnotation: amd64Digest,
		},
	}
	siblings := []v1.Descriptor{amd64Child, arm64Child, attestation}

	t.Run("annotation digest matches platform child", func(t *testing.T) {
		got := resolveAttestationSubject(attestation, siblings, nil)
		assert.Equal(t, amd64Digest, got)
	})

	t.Run("annotation digest not in index falls back to subject digest", func(t *testing.T) {
		desc := v1.Descriptor{
			MediaType: v1.MediaTypeImageManifest,
			Annotations: map[string]string{
				referenceTypeAnnotation:   attestationManifestType,
				referenceDigestAnnotation: "sha256:0000000000000000000000000000000000000000000000000000000000000000",
			},
		}
		subjects := []inTotoSubject{
			{
				Name:   "amd64",
				Digest: map[string]string{"sha256": "cad250bb95ea402adf4f687cc7d6747ecf0de875e6d6117f74437893964903df"},
			},
		}
		got := resolveAttestationSubject(desc, siblings, subjects)
		assert.Equal(t, amd64Digest, got)
	})

	t.Run("falls back to subject name matching", func(t *testing.T) {
		desc := v1.Descriptor{
			MediaType: v1.MediaTypeImageManifest,
			Annotations: map[string]string{
				referenceTypeAnnotation:   attestationManifestType,
				referenceDigestAnnotation: "sha256:0000000000000000000000000000000000000000000000000000000000000000",
			},
		}
		subjects := []inTotoSubject{
			{
				Name:   "linux/arm64",
				Digest: map[string]string{"sha256": "not-a-valid-hex-that-matches"},
			},
		}
		got := resolveAttestationSubject(desc, siblings, subjects)
		assert.Equal(t, arm64Digest, got)
	})

	t.Run("no siblings returns empty", func(t *testing.T) {
		got := resolveAttestationSubject(attestation, []v1.Descriptor{attestation}, nil)
		assert.Empty(t, got)
	})

	t.Run("nil subjects with annotation not matching returns empty", func(t *testing.T) {
		desc := v1.Descriptor{
			MediaType: v1.MediaTypeImageManifest,
			Annotations: map[string]string{
				referenceTypeAnnotation:   attestationManifestType,
				referenceDigestAnnotation: "sha256:0000000000000000000000000000000000000000000000000000000000000000",
			},
		}
		got := resolveAttestationSubject(desc, siblings, nil)
		assert.Empty(t, got)
	})

	t.Run("no annotation digest falls back to subject", func(t *testing.T) {
		desc := v1.Descriptor{
			MediaType: v1.MediaTypeImageManifest,
			Annotations: map[string]string{
				referenceTypeAnnotation: attestationManifestType,
			},
		}
		subjects := []inTotoSubject{
			{
				Name:   "arm64",
				Digest: map[string]string{"sha256": "480b518ed0138eacf2d070de80cb8eb019fb0b3565e2598ed654a541c31061a0"},
			},
		}
		got := resolveAttestationSubject(desc, siblings, subjects)
		assert.Equal(t, arm64Digest, got)
	})
}

// Resolution must not depend on the digest map's iteration order.
func TestResolveAttestationSubjectAmbiguousDigests(t *testing.T) {
	amd64Encoded := "cad250bb95ea402adf4f687cc7d6747ecf0de875e6d6117f74437893964903df"
	sha256Ref := "sha256:" + amd64Encoded
	sha512Encoded := strings.Repeat("ab", 64)

	siblings := []v1.Descriptor{
		{
			Digest:   digest.Digest(sha256Ref),
			Platform: &v1.Platform{OS: "linux", Architecture: "amd64"},
		},
		{
			Digest:   digest.Digest("sha512:" + sha512Encoded),
			Platform: &v1.Platform{OS: "linux", Architecture: "arm64"},
		},
	}
	attestation := v1.Descriptor{
		MediaType:   v1.MediaTypeImageManifest,
		Annotations: map[string]string{referenceTypeAnnotation: attestationManifestType},
	}

	t.Run("digests matching different siblings resolve to nothing", func(t *testing.T) {
		// name is empty so the name fallback cannot mask the ambiguity
		subjects := []inTotoSubject{{
			Digest: map[string]string{"sha256": amd64Encoded, "sha512": sha512Encoded},
		}}
		for range 50 {
			assert.Empty(t, resolveAttestationSubject(attestation, append(siblings, attestation), subjects))
		}
	})

	t.Run("a single matching algorithm resolves deterministically", func(t *testing.T) {
		subjects := []inTotoSubject{{
			Digest: map[string]string{"sha256": amd64Encoded, "sha512": strings.Repeat("cd", 64)},
		}}
		for range 50 {
			assert.Equal(t, sha256Ref, resolveAttestationSubject(attestation, append(siblings, attestation), subjects))
		}
	})
}
