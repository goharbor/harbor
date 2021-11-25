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
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/content/local"
	specs "github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// OCIStore provides content from the file system with the OCI-Image layout.
// Reference: https://github.com/opencontainers/image-spec/blob/master/image-layout.md
type OCIStore struct {
	content.Store

	root    string
	index   *ocispec.Index
	nameMap map[string]ocispec.Descriptor
}

// NewOCIStore creates a new OCI store
func NewOCIStore(rootPath string) (*OCIStore, error) {
	fileStore, err := local.NewStore(rootPath)
	if err != nil {
		return nil, err
	}

	store := &OCIStore{
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
func (s *OCIStore) LoadIndex() error {
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
func (s *OCIStore) SaveIndex() error {
	indexJSON, err := json.Marshal(s.index)
	if err != nil {
		return err
	}

	path := filepath.Join(s.root, OCIImageIndexFile)
	return ioutil.WriteFile(path, indexJSON, 0644)
}

// AddReference adds or updates an reference to index.
func (s *OCIStore) AddReference(name string, desc ocispec.Descriptor) {
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
func (s *OCIStore) DeleteReference(name string) {
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
func (s *OCIStore) ListReferences() map[string]ocispec.Descriptor {
	return s.nameMap
}

// validateOCILayoutFile ensures the `oci-layout` file
func (s *OCIStore) validateOCILayoutFile() error {
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
