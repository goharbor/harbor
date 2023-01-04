/*
Copyright The ORAS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package content

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/content/local"
	"github.com/containerd/containerd/remotes"
	"github.com/opencontainers/go-digest"
	specs "github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// OCI provides content from the file system with the OCI-Image layout.
// Reference: https://github.com/opencontainers/image-spec/blob/master/image-layout.md
type OCI struct {
	content.Store

	root    string
	index   *ocispec.Index
	nameMap map[string]ocispec.Descriptor
}

// NewOCI creates a new OCI store
func NewOCI(rootPath string) (*OCI, error) {
	fileStore, err := local.NewStore(rootPath)
	if err != nil {
		return nil, err
	}

	store := &OCI{
		Store: fileStore,
		root:  rootPath,
	}
	if err := store.validateOCILayoutFile(); err != nil {
		return nil, err
	}
	if err := store.LoadIndex(); err != nil {
		return nil, err
	}

	return store, nil
}

// LoadIndex reads the index.json from the file system
func (s *OCI) LoadIndex() error {
	path := filepath.Join(s.root, OCIImageIndexFile)
	indexFile, err := os.Open(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		s.index = &ocispec.Index{
			Versioned: specs.Versioned{
				SchemaVersion: 2, // historical value
			},
		}
		s.nameMap = make(map[string]ocispec.Descriptor)

		return nil
	}
	defer indexFile.Close()

	if err := json.NewDecoder(indexFile).Decode(&s.index); err != nil {
		return err
	}

	s.nameMap = make(map[string]ocispec.Descriptor)
	for _, desc := range s.index.Manifests {
		if name := desc.Annotations[ocispec.AnnotationRefName]; name != "" {
			s.nameMap[name] = desc
		}
	}

	return nil
}

// SaveIndex writes the index.json to the file system
func (s *OCI) SaveIndex() error {
	// first need to update the index
	var descs []ocispec.Descriptor
	for name, desc := range s.nameMap {
		if desc.Annotations == nil {
			desc.Annotations = map[string]string{}
		}
		desc.Annotations[ocispec.AnnotationRefName] = name
		descs = append(descs, desc)
	}
	s.index.Manifests = descs
	indexJSON, err := json.Marshal(s.index)
	if err != nil {
		return err
	}

	path := filepath.Join(s.root, OCIImageIndexFile)
	return ioutil.WriteFile(path, indexJSON, 0644)
}

func (s *OCI) Resolver() remotes.Resolver {
	return s
}

func (s *OCI) Resolve(ctx context.Context, ref string) (name string, desc ocispec.Descriptor, err error) {
	if err := s.LoadIndex(); err != nil {
		return "", ocispec.Descriptor{}, err
	}
	desc, ok := s.nameMap[ref]
	if !ok {
		return "", ocispec.Descriptor{}, fmt.Errorf("reference %s not in store", ref)
	}
	return ref, desc, nil
}

func (s *OCI) Fetcher(ctx context.Context, ref string) (remotes.Fetcher, error) {
	if err := s.LoadIndex(); err != nil {
		return nil, err
	}
	if _, ok := s.nameMap[ref]; !ok {
		return nil, fmt.Errorf("reference %s not in store", ref)
	}
	return s, nil
}

// Fetch get an io.ReadCloser for the specific content
func (s *OCI) Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	readerAt, err := s.Store.ReaderAt(ctx, desc)
	if err != nil {
		return nil, err
	}
	// just wrap the ReaderAt with a Reader
	return ioutil.NopCloser(&ReaderAtWrapper{readerAt: readerAt}), nil
}

// Pusher get a remotes.Pusher for the given ref
func (s *OCI) Pusher(ctx context.Context, ref string) (remotes.Pusher, error) {
	// separate the tag based ref from the hash
	var (
		baseRef, hash string
	)
	parts := strings.SplitN(ref, "@", 2)
	baseRef = parts[0]
	if len(parts) > 1 {
		hash = parts[1]
	}
	return &ociPusher{oci: s, ref: baseRef, digest: hash}, nil
}

// AddReference adds or updates an reference to index.
func (s *OCI) AddReference(name string, desc ocispec.Descriptor) {
	if desc.Annotations == nil {
		desc.Annotations = map[string]string{
			ocispec.AnnotationRefName: name,
		}
	} else {
		desc.Annotations[ocispec.AnnotationRefName] = name
	}

	if _, ok := s.nameMap[name]; ok {
		s.nameMap[name] = desc

		for i, ref := range s.index.Manifests {
			if name == ref.Annotations[ocispec.AnnotationRefName] {
				s.index.Manifests[i] = desc
				return
			}
		}

		// Process should not reach here.
		// Fallthrough to `Add` scenario and recover.
		s.index.Manifests = append(s.index.Manifests, desc)
		return
	}

	s.index.Manifests = append(s.index.Manifests, desc)
	s.nameMap[name] = desc
}

// DeleteReference deletes an reference from index.
func (s *OCI) DeleteReference(name string) {
	if _, ok := s.nameMap[name]; !ok {
		return
	}

	delete(s.nameMap, name)
	for i, desc := range s.index.Manifests {
		if name == desc.Annotations[ocispec.AnnotationRefName] {
			s.index.Manifests[i] = s.index.Manifests[len(s.index.Manifests)-1]
			s.index.Manifests = s.index.Manifests[:len(s.index.Manifests)-1]
			return
		}
	}
}

// ListReferences lists all references in index.
func (s *OCI) ListReferences() map[string]ocispec.Descriptor {
	return s.nameMap
}

// validateOCILayoutFile ensures the `oci-layout` file
func (s *OCI) validateOCILayoutFile() error {
	layoutFilePath := filepath.Join(s.root, ocispec.ImageLayoutFile)
	layoutFile, err := os.Open(layoutFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		layout := ocispec.ImageLayout{
			Version: ocispec.ImageLayoutVersion,
		}
		layoutJSON, err := json.Marshal(layout)
		if err != nil {
			return err
		}

		return ioutil.WriteFile(layoutFilePath, layoutJSON, 0644)
	}
	defer layoutFile.Close()

	var layout *ocispec.ImageLayout
	err = json.NewDecoder(layoutFile).Decode(&layout)
	if err != nil {
		return err
	}
	if layout.Version != ocispec.ImageLayoutVersion {
		return ErrUnsupportedVersion
	}

	return nil
}

// TODO: implement (needed to create a content.Store)
// TODO: do not return empty content.Info
// Abort completely cancels the ingest operation targeted by ref.
func (s *OCI) Info(ctx context.Context, dgst digest.Digest) (content.Info, error) {
	return content.Info{}, nil
}

// TODO: implement (needed to create a content.Store)
// Update updates mutable information related to content.
// If one or more fieldpaths are provided, only those
// fields will be updated.
// Mutable fields:
//  labels.*
func (s *OCI) Update(ctx context.Context, info content.Info, fieldpaths ...string) (content.Info, error) {
	return content.Info{}, errors.New("not yet implemented: Update (content.Store interface)")
}

// TODO: implement (needed to create a content.Store)
// Walk will call fn for each item in the content store which
// match the provided filters. If no filters are given all
// items will be walked.
func (s *OCI) Walk(ctx context.Context, fn content.WalkFunc, filters ...string) error {
	return errors.New("not yet implemented: Walk (content.Store interface)")
}

// Delete removes the content from the store.
func (s *OCI) Delete(ctx context.Context, dgst digest.Digest) error {
	return s.Store.Delete(ctx, dgst)
}

// TODO: implement (needed to create a content.Store)
func (s *OCI) Status(ctx context.Context, ref string) (content.Status, error) {
	// Status returns the status of the provided ref.
	return content.Status{}, errors.New("not yet implemented: Status (content.Store interface)")
}

// TODO: implement (needed to create a content.Store)
// ListStatuses returns the status of any active ingestions whose ref match the
// provided regular expression. If empty, all active ingestions will be
// returned.
func (s *OCI) ListStatuses(ctx context.Context, filters ...string) ([]content.Status, error) {
	return []content.Status{}, errors.New("not yet implemented: ListStatuses (content.Store interface)")
}

// TODO: implement (needed to create a content.Store)
// Abort completely cancels the ingest operation targeted by ref.
func (s *OCI) Abort(ctx context.Context, ref string) error {
	return errors.New("not yet implemented: Abort (content.Store interface)")
}

// ReaderAt provides contents
func (s *OCI) ReaderAt(ctx context.Context, desc ocispec.Descriptor) (content.ReaderAt, error) {
	return s.Store.ReaderAt(ctx, desc)
}

// ociPusher to push content for a single referencem can handle multiple descriptors.
// Needs to be able to recognize when a root manifest is being pushed and to create the tag
// for it.
type ociPusher struct {
	oci    *OCI
	ref    string
	digest string
}

// Push get a writer for a single Descriptor
func (p *ociPusher) Push(ctx context.Context, desc ocispec.Descriptor) (content.Writer, error) {
	// do we need to create a tag?
	switch desc.MediaType {
	case ocispec.MediaTypeImageManifest, ocispec.MediaTypeImageIndex:
		// if the hash of the content matches that which was provided as the hash for the root, mark it
		if p.digest != "" && p.digest == desc.Digest.String() {
			if err := p.oci.LoadIndex(); err != nil {
				return nil, err
			}
			p.oci.nameMap[p.ref] = desc
			if err := p.oci.SaveIndex(); err != nil {
				return nil, err
			}
		}
	}

	return p.oci.Store.Writer(ctx, content.WithDescriptor(desc), content.WithRef(p.ref))
}
