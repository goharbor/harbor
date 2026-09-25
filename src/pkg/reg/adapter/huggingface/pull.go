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

package huggingface

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"sync"

	"github.com/docker/distribution"
	_ "github.com/docker/distribution/manifest/ocischema" // registers the OCI manifest unmarshaller
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sync/errgroup"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/synth"
)

const hashConcurrency = 8

// The ArtifactRegistry methods carry no context. Cancellation of a download happens when the
// caller closes the returned body.

// ManifestExist ...
func (a *adapter) ManifestExist(repository, reference string) (bool, *distribution.Descriptor, error) {
	art, err := a.resolve(context.Background(), repository, reference)
	if errors.IsNotFoundErr(err) {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}
	return true, &distribution.Descriptor{
		MediaType: v1.MediaTypeImageManifest,
		Digest:    art.Digest,
		Size:      int64(len(art.Manifest)),
	}, nil
}

// PullManifest accepts a tag, a commit or a manifest digest as reference.
func (a *adapter) PullManifest(repository, reference string, _ ...string) (distribution.Manifest, string, error) {
	art, err := a.resolve(context.Background(), repository, reference)
	if err != nil {
		return nil, "", err
	}
	manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, art.Manifest)
	if err != nil {
		return nil, "", err
	}
	return manifest, art.Digest.String(), nil
}

// BlobExist ...
func (a *adapter) BlobExist(repository, dgst string) (bool, error) {
	d, err := digest.Parse(dgst)
	if err != nil {
		return false, errors.BadRequestError(err)
	}
	_, err = a.find(context.Background(), repository, d)
	if errors.IsNotFoundErr(err) {
		return false, nil
	}
	return err == nil, err
}

// PullBlob ...
func (a *adapter) PullBlob(repository, dgst string) (int64, io.ReadCloser, error) {
	return a.openBlob(context.Background(), repository, dgst, 0, -1)
}

// PullBlobChunk ...
func (a *adapter) PullBlobChunk(repository, dgst string, _, start, end int64) (int64, io.ReadCloser, error) {
	if start < 0 || end < start {
		return 0, nil, errors.BadRequestError(nil).WithMessagef("invalid range %d-%d", start, end)
	}
	return a.openBlob(context.Background(), repository, dgst, start, end-start+1)
}

// ListTags ...
func (a *adapter) ListTags(repository string) ([]string, error) {
	ctx := context.Background()
	refs, err := a.hub.Refs(ctx, a.locator.modelID(ctx, repository))
	if err != nil {
		return nil, err
	}
	return newRefIndex(refs).tags(), nil
}

// ListReferrers ...
func (a *adapter) ListReferrers(_, _, _ string) (*v1.Index, map[string][]string, error) {
	return &v1.Index{
		Versioned: specs.Versioned{SchemaVersion: 2},
		MediaType: v1.MediaTypeImageIndex,
		Manifests: []v1.Descriptor{},
	}, map[string][]string{}, nil
}

// resolve returns the artifact of a tag, a commit or a manifest digest.
func (a *adapter) resolve(ctx context.Context, repository, reference string) (*synth.Artifact, error) {
	if d, err := digest.Parse(reference); err == nil {
		loc, err := a.find(ctx, repository, d)
		if err != nil {
			return nil, err
		}
		if loc.Kind != kindManifest {
			return nil, errors.NotFoundError(fmt.Errorf("%s is not a manifest of %s", d, repository))
		}
		art, _, err := a.artifact(ctx, repository, loc.Commit)
		if err != nil {
			return nil, err
		}
		if art.Digest != d {
			return nil, errors.NotFoundError(fmt.Errorf("manifest %s of %s is no longer produced", d, repository))
		}
		return art, nil
	}
	commit := reference
	if !hub.IsCommit(reference) {
		var err error
		if commit, err = a.commitOf(ctx, repository, reference); err != nil {
			return nil, err
		}
	}
	art, _, err := a.artifact(ctx, repository, commit)
	return art, err
}

func (a *adapter) commitOf(ctx context.Context, repository, tag string) (string, error) {
	var commit string
	err := cache.FetchOrSave(ctx, a.locator.cache, a.locator.tagKey(repository, tag), &commit, func() (any, error) {
		refs, err := a.hub.Refs(ctx, a.locator.modelID(ctx, repository))
		if err != nil {
			return nil, err
		}
		return newRefIndex(refs).commit(tag)
	}, tagTTL)
	return commit, err
}

// artifact synthesizes the artifact of a commit, indexes its digests and returns it with the model ID.
func (a *adapter) artifact(ctx context.Context, repository, commit string) (*synth.Artifact, string, error) {
	s := &hub.Snapshot{}
	err := cache.FetchOrSave(ctx, a.locator.cache, a.locator.snapshotKey(repository, commit), s, func() (any, error) {
		snapshot, err := a.hub.Snapshot(ctx, a.locator.modelID(ctx, repository), commit)
		if err != nil {
			return nil, err
		}
		a.locator.setModelID(ctx, repository, snapshot.ModelID)
		return snapshot, nil
	}, snapshotTTL)
	if err != nil {
		return nil, "", err
	}
	digests, err := a.contentDigests(ctx, s)
	if err != nil {
		return nil, "", err
	}
	art, err := synth.Build(s, digests)
	if err != nil {
		return nil, "", err
	}
	a.locator.index(ctx, repository, s.ModelID, s.Commit, art)
	return art, s.ModelID, nil
}

// contentDigests returns the sha256 of every included non-LFS file, downloading only the ones
// whose git blob id is not cached yet.
func (a *adapter) contentDigests(ctx context.Context, s *hub.Snapshot) (map[string]digest.Digest, error) {
	var mu sync.Mutex
	digests := map[string]digest.Digest{}
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(hashConcurrency)
	for _, f := range s.Files {
		if f.IsLFS() || !synth.Included(f.Path) {
			continue
		}
		g.Go(func() error {
			var d string
			err := cache.FetchOrSave(gctx, a.locator.cache, a.locator.blobKey(f.BlobID), &d, func() (any, error) {
				dg, err := a.hub.ContentDigest(gctx, s.ModelID, s.Commit, f)
				return dg.String(), err
			}, blobTTL)
			if err != nil {
				return err
			}
			mu.Lock()
			digests[f.Path] = digest.Digest(d)
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return digests, nil
}

// find locates a digest. On a cache miss it synthesizes the head of every ref of the
// repository, main first, which makes it correct after cache loss. A digest only reachable from a
// commit that is no longer a ref head is not found; the client re-requests by tag, which refills
// the index. Not found results are cached briefly.
func (a *adapter) find(ctx context.Context, repository string, d digest.Digest) (*location, error) {
	loc, err := a.locator.locate(ctx, repository, d)
	if err == nil {
		return loc, nil
	}
	if !errors.Is(err, cache.ErrNotFound) {
		return nil, err
	}
	notFound := errors.NotFoundError(fmt.Errorf("digest %s not found in any ref head of %s", d, repository))
	if a.locator.knownMissing(ctx, repository, d) {
		return nil, notFound
	}
	refs, err := a.hub.Refs(ctx, a.locator.modelID(ctx, repository))
	if err != nil {
		return nil, err
	}
	for _, commit := range headsMainFirst(refs) {
		art, modelID, err := a.artifact(ctx, repository, commit)
		if err != nil {
			if errors.IsRateLimitError(err) {
				return nil, err
			}
			log.Warningf("skip commit %s of %s while locating %s: %v", commit, repository, d, err)
			continue
		}
		if loc := locationIn(art, modelID, commit, d); loc != nil {
			a.locator.put(ctx, repository, d, loc)
			return loc, nil
		}
	}
	a.locator.markMissing(ctx, repository, d)
	return nil, notFound
}

// defaultBranch is the branch every Hub repository is created with.
const defaultBranch = "main"

// headsMainFirst returns the distinct commits of refs, the default branch first, then by sha.
func headsMainFirst(refs []hub.Ref) []string {
	seen := map[string]bool{}
	var main string
	var others []string
	for _, r := range refs {
		if r.Name == defaultBranch {
			main = r.Commit
		}
	}
	if main != "" {
		seen[main] = true
	}
	for _, r := range refs {
		if !seen[r.Commit] {
			seen[r.Commit] = true
			others = append(others, r.Commit)
		}
	}
	sort.Strings(others)
	if main == "" {
		return others
	}
	return append([]string{main}, others...)
}

func locationIn(art *synth.Artifact, modelID, commit string, d digest.Digest) *location {
	switch d {
	case art.Digest:
		return &location{ModelID: modelID, Commit: commit, Kind: kindManifest}
	case art.ConfigDigest:
		return &location{ModelID: modelID, Commit: commit, Kind: kindConfig}
	}
	for _, l := range art.Layers {
		if l.Digest == d {
			return &location{ModelID: modelID, Commit: commit, Kind: kindLayer, Path: l.Path, Size: l.Size}
		}
	}
	return nil
}

func (a *adapter) openBlob(ctx context.Context, repository, dgst string, offset, length int64) (int64, io.ReadCloser, error) {
	d, err := digest.Parse(dgst)
	if err != nil {
		return 0, nil, errors.BadRequestError(err)
	}
	loc, err := a.find(ctx, repository, d)
	if err != nil {
		return 0, nil, err
	}
	size, body, err := a.open(ctx, repository, loc, d, offset, length)
	if errors.IsNotFoundErr(err) {
		// The commit may be gone from the Hub (e.g. squashed history); relocate once.
		a.locator.forget(ctx, repository, d)
		if loc, err = a.find(ctx, repository, d); err != nil {
			return 0, nil, err
		}
		size, body, err = a.open(ctx, repository, loc, d, offset, length)
	}
	return size, body, err
}

func (a *adapter) open(ctx context.Context, repository string, loc *location, d digest.Digest, offset, length int64) (int64, io.ReadCloser, error) {
	if loc.Kind == kindLayer {
		if length < 0 {
			length = loc.Size - offset
		}
		if offset+length > loc.Size {
			return 0, nil, errors.BadRequestError(nil).WithMessagef("range %d+%d exceeds the size %d of %s", offset, length, loc.Size, d)
		}
		body, err := a.hub.Open(ctx, loc.ModelID, loc.Commit, loc.Path, offset, length)
		if err != nil {
			return 0, nil, err
		}
		return length, body, nil
	}
	art, _, err := a.artifact(ctx, repository, loc.Commit)
	if err != nil {
		return 0, nil, err
	}
	content := art.Manifest
	if loc.Kind == kindConfig {
		content = art.Config
	}
	if digest.FromBytes(content) != d {
		return 0, nil, errors.NotFoundError(fmt.Errorf("%s of %s is no longer produced", d, repository))
	}
	if length < 0 {
		length = int64(len(content)) - offset
	}
	if offset+length > int64(len(content)) {
		return 0, nil, errors.BadRequestError(nil).WithMessagef("range %d+%d exceeds the size %d of %s", offset, length, len(content), d)
	}
	return length, io.NopCloser(bytes.NewReader(content[offset : offset+length])), nil
}
