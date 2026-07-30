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
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/docker/distribution/manifest/schema2"
	digest "github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg"
	accessorymodel "github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/registry"
)

// Attestations are linked to the images they describe by these annotations
// rather than by a distinct manifest media type.
// https://github.com/moby/buildkit/blob/master/docs/attestations/attestation-storage.md
const (
	referenceTypeAnnotation         = "vnd.docker.reference.type"
	referenceDigestAnnotation       = "vnd.docker.reference.digest"
	attestationManifestType         = "attestation-manifest"
	inTotoLayerMediaType            = "application/vnd.in-toto+json"
	maxStatementBytes         int64 = 4 << 20 // 4 MiB
)

func init() {
	RegisterChildClassifier(NewInTotoAttestationClassifier(pkg.ArtifactMgr, registry.Cli))
}

type inTotoStatement struct {
	Subject []inTotoSubject `json:"subject"`
}

type inTotoSubject struct {
	Name   string            `json:"name"`
	Digest map[string]string `json:"digest"`
}

// InTotoAttestationClassifier classifies in-toto attestation manifests as
// accessories of the image they attest. Named after the attestation format, not
// its producer: BuildKit, Buildah and Podman all emit this layout.
type InTotoAttestationClassifier struct {
	artMgr artifact.Manager
	regCli registry.Client
}

// NewInTotoAttestationClassifier returns a classifier for in-toto attestations.
func NewInTotoAttestationClassifier(artMgr artifact.Manager, regCli registry.Client) *InTotoAttestationClassifier {
	return &InTotoAttestationClassifier{
		artMgr: artMgr,
		regCli: regCli,
	}
}

func (c *InTotoAttestationClassifier) Classify(ctx context.Context, repository string, descriptor v1.Descriptor, siblings []v1.Descriptor) (*artifact.AccessoryCandidate, error) {
	if !isAttestationDescriptor(descriptor) {
		//nolint:nilnil // descriptor is not an attestation
		return nil, nil
	}

	subjects, err := c.loadSubjects(repository, descriptor.Digest.String())
	if err != nil {
		log.G(ctx).Debugf("could not load attestation subjects for %s@%s, falling back to annotation-based resolution: %v", repository, descriptor.Digest.String(), err)
	}

	targetDigest := resolveAttestationSubject(descriptor, siblings, subjects)
	if targetDigest == "" {
		//nolint:nilnil // no target artifact could be resolved
		return nil, nil
	}

	accessoryArt, err := c.artMgr.GetByDigest(ctx, repository, descriptor.Digest.String())
	if err != nil {
		return nil, err
	}
	targetArt, err := c.artMgr.GetByDigest(ctx, repository, targetDigest)
	if err != nil {
		return nil, err
	}

	return &artifact.AccessoryCandidate{
		ArtifactID:        accessoryArt.ID,
		SubArtifactID:     targetArt.ID,
		SubArtifactRepo:   repository,
		SubArtifactDigest: targetDigest,
		Digest:            accessoryArt.Digest,
		Size:              accessoryArt.Size,
		Type:              accessorymodel.TypeInTotoAttestation,
	}, nil
}

func (c *InTotoAttestationClassifier) loadSubjects(repository, digestRef string) ([]inTotoSubject, error) {
	manifest, _, err := c.regCli.PullManifest(repository, digestRef)
	if err != nil {
		return nil, err
	}

	mediaType, content, err := manifest.Payload()
	if err != nil {
		return nil, err
	}
	if mediaType != v1.MediaTypeImageManifest && mediaType != schema2.MediaTypeManifest {
		return nil, fmt.Errorf("unexpected attestation media type %s", mediaType)
	}

	imageManifest := &v1.Manifest{}
	if err := json.Unmarshal(content, imageManifest); err != nil {
		return nil, err
	}

	for _, layer := range imageManifest.Layers {
		if layer.MediaType != inTotoLayerMediaType {
			continue
		}
		// reject on the advertised size before spending a blob pull on it
		if layer.Size > maxStatementBytes {
			return nil, fmt.Errorf("in-toto payload too large: %d", layer.Size)
		}
		return c.readSubjects(repository, layer.Digest.String())
	}

	return nil, fmt.Errorf("no in-toto payload found in %s", digestRef)
}

func (c *InTotoAttestationClassifier) readSubjects(repository, digestRef string) ([]inTotoSubject, error) {
	_, blob, err := c.regCli.PullBlob(repository, digestRef)
	if err != nil {
		return nil, err
	}
	defer blob.Close()

	payload, err := io.ReadAll(io.LimitReader(blob, maxStatementBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(payload)) > maxStatementBytes {
		return nil, fmt.Errorf("in-toto payload exceeds %d bytes", maxStatementBytes)
	}

	statement := &inTotoStatement{}
	if err := json.Unmarshal(payload, statement); err != nil {
		return nil, err
	}
	return statement.Subject, nil
}

func isAttestationDescriptor(descriptor v1.Descriptor) bool {
	if descriptor.MediaType != v1.MediaTypeImageManifest && descriptor.MediaType != schema2.MediaTypeManifest {
		return false
	}
	return descriptor.Annotations[referenceTypeAnnotation] == attestationManifestType
}

func resolveAttestationSubject(descriptor v1.Descriptor, siblings []v1.Descriptor, subjects []inTotoSubject) string {
	children := platformChildren(siblings)
	if len(children) == 0 {
		return ""
	}

	if digestRef := descriptor.Annotations[referenceDigestAnnotation]; digestInIndex(children, digestRef) {
		return digestRef
	}

	for _, subject := range subjects {
		if digestRef := uniqueDigestInIndex(children, subjectDigests(subject)); digestRef != "" {
			return digestRef
		}
	}

	for _, subject := range subjects {
		if digestRef := digestBySubjectName(children, subject.Name); digestRef != "" {
			return digestRef
		}
	}

	return ""
}

func platformChildren(siblings []v1.Descriptor) []v1.Descriptor {
	children := make([]v1.Descriptor, 0, len(siblings))
	for _, sibling := range siblings {
		if isAttestationDescriptor(sibling) {
			continue
		}
		children = append(children, sibling)
	}
	return children
}

func digestInIndex(siblings []v1.Descriptor, digestRef string) bool {
	if digestRef == "" {
		return false
	}
	for _, sibling := range siblings {
		if sibling.Digest.String() == digestRef {
			return true
		}
	}
	return false
}

// A subject names one artifact under several algorithms, so matching more than
// one sibling means the payload disagrees with the index. Rejecting that also
// keeps the result independent of subjectDigests' map iteration order.
func uniqueDigestInIndex(siblings []v1.Descriptor, digestRefs []string) string {
	match := ""
	for _, digestRef := range digestRefs {
		if !digestInIndex(siblings, digestRef) {
			continue
		}
		if match != "" && match != digestRef {
			return ""
		}
		match = digestRef
	}
	return match
}

func subjectDigests(subject inTotoSubject) []string {
	digests := make([]string, 0, len(subject.Digest))
	for algorithm, encoded := range subject.Digest {
		if encoded == "" {
			continue
		}
		digestRef := digest.NewDigestFromEncoded(digest.Algorithm(algorithm), encoded)
		if digestRef.Validate() == nil {
			digests = append(digests, digestRef.String())
		}
	}
	return digests
}

func digestBySubjectName(siblings []v1.Descriptor, name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return ""
	}

	match := ""
	for _, sibling := range siblings {
		if !platformMatchesName(sibling.Platform, name) {
			continue
		}
		if match != "" {
			return ""
		}
		match = sibling.Digest.String()
	}
	return match
}

func platformMatchesName(platform *v1.Platform, name string) bool {
	if platform == nil {
		return false
	}

	candidates := []string{
		platform.Architecture,
	}
	if platform.OS != "" && platform.Architecture != "" {
		candidates = append(candidates, platform.OS+"/"+platform.Architecture)
	}
	if platform.OS != "" && platform.Architecture != "" && platform.Variant != "" {
		candidates = append(candidates, platform.OS+"/"+platform.Architecture+"/"+platform.Variant)
	}

	for _, candidate := range candidates {
		if candidate != "" && strings.EqualFold(candidate, name) {
			return true
		}
	}
	return false
}
