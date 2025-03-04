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

package cnai

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"path/filepath"
	"strings"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	ps "github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

const (
	ArtifactTypeCNAI = "MODEL"
	MediaType        = "application/vnd.cnai.model.config.v1+json"

	FilepathAnnotation = "org.cnai.model.filepath"

	DocMediaType = "application/vnd.cnai.model.doc.v1.tar"

	AdditionTypeReadme  = "README.MD"
	AdditionTypeLicense = "LICENSE"
	AdditionTypeFiles   = "FILES"
)

const (
	ReadmeFilename       = AdditionTypeReadme
	ReadmeShortFilename  = "README"
	LicenseFilename      = "LICENSE.txt"
	LicenseShortFilename = "LICENSE"
)

// Processor handles CNAI model artifacts processing operations, extending the base ManifestProcessor
type Processor struct {
	*base.ManifestProcessor
}

// FileInfo represents metadata about a file or directory in the artifact file tree
type FileInfo struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Size     int64       `json:"size,omitempty"`
	Children []*FileInfo `json:"children,omitempty"`
}

func init() {
	pc := &Processor{}

	pc.ManifestProcessor = base.NewManifestProcessor()
	if err := ps.Register(pc, MediaType); err != nil {
		log.Errorf("failed to register processor for media type %s: %v", MediaType, err)

		return
	}
}

// isSupportedAddition checks if the specified addition type is supported for CNAI artifacts
func isSupportedAddition(addition string) error {
	if addition == AdditionTypeReadme {
		return nil
	}

	if addition == AdditionTypeLicense {
		return nil
	}

	if addition == AdditionTypeFiles {
		return nil
	}

	return errors.New(nil).WithCode(errors.BadRequestCode).
		WithMessagef("addition %s isn't supported for %s", addition, ArtifactTypeCNAI)
}

// extractFileFromCompressedLayer extracts a specific file from compressed layer data by name
func extractFileFromCompressedLayer(b []byte, name string) ([]byte, error) {
	r := bytes.NewReader(b)
	tarReader := tar.NewReader(r)

	// iterate through all files in the tar archive
	for {
		// get the next header, which contains file metadata
		header, err := tarReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, err
		}

		// skip entries that are not regular files (directories, symlinks, etc.)
		if header.Typeflag != tar.TypeReg {
			continue
		}

		// check if current file matches (case-insensitive) the requested name
		if strings.EqualFold(filepath.Base(header.Name), name) ||
			strings.EqualFold(header.Name, name) {
			var buffer bytes.Buffer
			if _, err := io.Copy(&buffer, tarReader); err != nil {
				return nil, err
			}

			return buffer.Bytes(), nil
		}
	}

	return nil, errors.New(nil).WithCode(errors.NotFoundCode).
		WithMessagef("file %q not found in compressed layer", name)
}

// extractLayerContent pulls a blob from the registry and extracts file content based on the filepath annotation
func extractLayerContent(p *Processor, repositoryName, layerDigest, filepathAnnotation string) ([]byte, error) {
	// pull the blob from the registry using the repository name and layer digest
	_, blob, err := p.RegCli.PullBlob(repositoryName, layerDigest)
	if err != nil {
		return nil, err
	}
	defer blob.Close()

	// read the entire blob content into memory
	tarContent, err := io.ReadAll(blob)
	if err != nil {
		return nil, err
	}

	// extract the specific file from the compressed layer using the filepath annotation
	fileContent, err := extractFileFromCompressedLayer(
		tarContent, filepathAnnotation,
	)
	if err != nil {
		return nil, err
	}

	return fileContent, nil
}

// isReadMeLayer checks if layer contains a ReadMe file
func isReadMeLayer(layer v1.Descriptor) bool {
	if layer.MediaType != DocMediaType {
		return false
	}

	val, ok := layer.Annotations[FilepathAnnotation]
	if !ok {
		return false
	}

	if strings.EqualFold(val, ReadmeShortFilename) {
		return true
	}

	if strings.EqualFold(val, ReadmeFilename) {
		return true
	}

	return false
}

// isLicenseLayer checks if layer contains a License file
func isLicenseLayer(layer v1.Descriptor) bool {
	if layer.MediaType != DocMediaType {
		return false
	}

	val, ok := layer.Annotations[FilepathAnnotation]
	if !ok {
		return false
	}

	if strings.EqualFold(val, LicenseShortFilename) {
		return true
	}

	if strings.EqualFold(val, LicenseFilename) {
		return true
	}

	return false
}

// processPath processes file path components to build a tree structure of FileInfo nodes
func processPath(parts []string, size int64, tree map[string]*FileInfo, rootEntries *[]*FileInfo) {
	var (
		currentPath string    // tracks the full path
		parent      *FileInfo // reference to the parent node
	)

	// isFile determines if the current component is a file (last part of the path)
	isFile := func(i int, parts []string) bool {
		return i == len(parts)-1
	}

	// buildPath constructs the full path by appending the current part
	buildPath := func(currentPath, part string) string {
		if currentPath != "" {
			return currentPath + "/" + part
		}

		return part
	}

	// createNode instantiates a new FileInfo node with appropriate type and properties
	createNode := func(part string, size int64, isFile bool) *FileInfo {
		nodeType := "directory"
		if isFile {
			nodeType = "file"
		}

		node := &FileInfo{
			Name: part,
			Type: nodeType,
		}

		if isFile {
			node.Size = size
		} else {
			node.Children = []*FileInfo{}
		}

		return node
	}

	// existsInRoot checks if a node with the given name already exists at the root level
	existsInRoot := func(rootEntries *[]*FileInfo, name string) bool {
		for _, root := range *rootEntries {
			if root.Name == name {
				return true
			}
		}

		return false
	}

	for i, part := range parts {
		// build the full path for the current component
		currentPath = buildPath(currentPath, part)

		// skip if we already processed this path
		if node, exists := tree[currentPath]; exists {
			parent = node

			continue
		}

		// create new node based on whether it's a file or directory
		node := createNode(part, size, isFile(i, parts))

		// add to lookup map for future reference
		tree[currentPath] = node

		// link to parent or add to root entries
		if parent != nil {
			parent.Children = append(parent.Children, node)
		} else if !existsInRoot(rootEntries, node.Name) {
			*rootEntries = append(*rootEntries, node)
		}

		// current node becomes parent for next iteration
		parent = node
	}
}

// buildFileTree constructs a hierarchical file tree structure from layer descriptors
func buildFileTree(layers []v1.Descriptor) (rootEntries []*FileInfo) {
	// create a map to track all nodes in the tree by their path
	tree := make(map[string]*FileInfo)

	// helper function to split a path into standardized components
	splitPath := func(in string) []string {
		var parts []string

		for _, part := range strings.Split(filepath.ToSlash(in), "/") {
			if part != "" {
				parts = append(parts, part)
			}
		}

		return parts
	}

	// iterate through each layer to build the file tree
	for _, layer := range layers {
		// get the filepath annotation from the layer
		val, ok := layer.Annotations[FilepathAnnotation]
		if !ok || val == "" {
			continue
		}

		// process the path components to build the tree structure
		processPath(splitPath(val), layer.Size, tree, &rootEntries)
	}

	return rootEntries
}

// extractContent retrieves content from a layer and returns it as an Addition with specified content type
func (p *Processor) extractContent(repoName, layerDigest, filepath, contentType string) (*ps.Addition, error) {
	fileContent, err := extractLayerContent(
		p, repoName, layerDigest,
		filepath,
	)
	if err != nil {
		return nil, err
	}

	return &ps.Addition{
		Content:     fileContent,
		ContentType: contentType,
	}, nil
}

// getManifest retrieves and parses the OCI manifest for the artifact
func (p *Processor) getManifest(repoName string, digest string) (*v1.Manifest, error) {
	m, _, err := p.RegCli.PullManifest(repoName, digest)
	if err != nil {
		return nil, err
	}

	_, payload, err := m.Payload()
	if err != nil {
		return nil, err
	}

	manifest := &v1.Manifest{}
	if err := json.Unmarshal(payload, manifest); err != nil {
		return nil, err
	}

	return manifest, nil
}

// AbstractAddition extracts and returns the specified addition content for a CNAI artifact
func (p *Processor) AbstractAddition(_ context.Context, artifact *artifact.Artifact, addition string) (*ps.Addition, error) {
	if err := isSupportedAddition(addition); err != nil {
		return nil, err
	}

	manifest, err := p.getManifest(artifact.RepositoryName, artifact.Digest)
	if err != nil {
		return nil, err
	}

	// handle file tree addition
	if addition == AdditionTypeFiles {
		fileTree := buildFileTree(manifest.Layers)

		fileListContent, err := json.MarshalIndent(fileTree, "", "  ")
		if err != nil {
			return nil, err
		}

		return &ps.Addition{
			Content:     fileListContent,
			ContentType: "application/json; charset=utf-8",
		}, nil
	}

	// iterate through layers to find requested content (readme or license)
	for _, layer := range manifest.Layers {
		switch addition {
		case AdditionTypeReadme:
			if isReadMeLayer(layer) {
				return p.extractContent(
					artifact.RepositoryName,
					layer.Digest.String(),
					layer.Annotations[FilepathAnnotation],
					"text/markdown; charset=utf-8",
				)
			}
		case AdditionTypeLicense:
			if isLicenseLayer(layer) {
				return p.extractContent(
					artifact.RepositoryName,
					layer.Digest.String(),
					layer.Annotations[FilepathAnnotation],
					"text/plain; charset=utf-8",
				)
			}
		}
	}

	// requested addition wasn't found in any layer
	return nil, nil
}

// GetArtifactType returns the type of the CNAI artifact
func (p *Processor) GetArtifactType(_ context.Context, _ *artifact.Artifact) string {
	return ArtifactTypeCNAI
}

// ListAdditionTypes returns a list of all supported addition types for CNAI artifacts
func (p *Processor) ListAdditionTypes(_ context.Context, _ *artifact.Artifact) []string {
	return []string{
		AdditionTypeReadme,
		AdditionTypeLicense,
		AdditionTypeFiles,
	}
}
