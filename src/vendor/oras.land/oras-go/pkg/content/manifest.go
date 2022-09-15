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
	"sort"

	"github.com/opencontainers/go-digest"
	specs "github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	artifact "oras.land/oras-go/pkg/artifact"
)

// GenerateManifest generates a manifest. The manifest will include the provided config,
// and descs as layers. Raw bytes will be returned.
func GenerateManifest(config *ocispec.Descriptor, annotations map[string]string, descs ...ocispec.Descriptor) ([]byte, ocispec.Descriptor, error) {
	// Config - either it was set, or we have to set it
	if config == nil {
		_, configGen, err := GenerateConfig(nil)
		if err != nil {
			return nil, ocispec.Descriptor{}, err
		}
		config = &configGen
	}
	return pack(*config, annotations, descs)
}

// GenerateConfig generates a blank config with optional annotations.
func GenerateConfig(annotations map[string]string) ([]byte, ocispec.Descriptor, error) {
	configBytes := []byte("{}")
	dig := digest.FromBytes(configBytes)
	config := ocispec.Descriptor{
		MediaType:   artifact.UnknownConfigMediaType,
		Digest:      dig,
		Size:        int64(len(configBytes)),
		Annotations: annotations,
	}
	return configBytes, config, nil
}

// GenerateManifestAndConfig generates a config and then a manifest. Raw bytes will be returned.
func GenerateManifestAndConfig(manifestAnnotations map[string]string, configAnnotations map[string]string, descs ...ocispec.Descriptor) (manifest []byte, manifestDesc ocispec.Descriptor, config []byte, configDesc ocispec.Descriptor, err error) {
	config, configDesc, err = GenerateConfig(configAnnotations)
	if err != nil {
		return nil, ocispec.Descriptor{}, nil, ocispec.Descriptor{}, err
	}
	manifest, manifestDesc, err = GenerateManifest(&configDesc, manifestAnnotations, descs...)
	if err != nil {
		return nil, ocispec.Descriptor{}, nil, ocispec.Descriptor{}, err
	}
	return
}

// pack given a bunch of descriptors, create a manifest that references all of them
func pack(config ocispec.Descriptor, annotations map[string]string, descriptors []ocispec.Descriptor) ([]byte, ocispec.Descriptor, error) {
	if descriptors == nil {
		descriptors = []ocispec.Descriptor{} // make it an empty array to prevent potential server-side bugs
	}
	// sort descriptors alphanumerically by sha hash so it always is consistent
	sort.Slice(descriptors, func(i, j int) bool {
		return descriptors[i].Digest < descriptors[j].Digest
	})
	manifest := ocispec.Manifest{
		Versioned: specs.Versioned{
			SchemaVersion: 2, // historical value. does not pertain to OCI or docker version
		},
		Config:      config,
		Layers:      descriptors,
		Annotations: annotations,
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return nil, ocispec.Descriptor{}, err
	}
	manifestDescriptor := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    digest.FromBytes(manifestBytes),
		Size:      int64(len(manifestBytes)),
	}

	return manifestBytes, manifestDescriptor, nil
}
