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

package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	modelspec "github.com/CloudNativeAI/model-spec/specs-go/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/registry"
)

// NewFiles creates a new files parser.
func NewFiles(cli registry.Client) Parser {
	return &files{
		base: newBase(cli),
	}
}

// files is the parser for listing files in the model artifact.
type files struct {
	*base
}

type FileList struct {
	Name     string     `json:"name"`
	Type     string     `json:"type"`
	Size     int64      `json:"size,omitempty"`
	Children []FileList `json:"children,omitempty"`
}

// Parse parses the files list.
func (f *files) Parse(_ context.Context, _ *artifact.Artifact, manifest *ocispec.Manifest) (string, []byte, error) {
	if manifest == nil {
		return "", nil, fmt.Errorf("manifest cannot be nil")
	}

	rootNode, err := walkManifest(*manifest)
	if err != nil {
		return "", nil, fmt.Errorf("failed to walk manifest: %w", err)
	}

	fileLists := traverseFileNode(rootNode)
	data, err := json.Marshal(fileLists)
	if err != nil {
		return "", nil, err
	}

	return contentTypeJSON, data, nil
}

// walkManifest walks the manifest and returns the root file node.
func walkManifest(manifest ocispec.Manifest) (*FileNode, error) {
	root := NewDirectory("/")
	for _, layer := range manifest.Layers {
		if layer.Annotations != nil && layer.Annotations[modelspec.AnnotationFilepath] != "" {
			filepath := layer.Annotations[modelspec.AnnotationFilepath]
			// mark it to directory if the file path ends with "/".
			isDir := filepath[len(filepath)-1] == '/'
			_, err := root.AddNode(filepath, layer.Size, isDir)
			if err != nil {
				return nil, err
			}
		}
	}

	return root, nil
}

// traverseFileNode traverses the file node and returns the file list.
func traverseFileNode(node *FileNode) []FileList {
	if node == nil {
		return nil
	}

	var children []FileList
	for _, child := range node.Children {
		children = append(children, FileList{
			Name:     child.Name,
			Type:     child.Type,
			Size:     child.Size,
			Children: traverseFileNode(child),
		})
	}

	// sort the children by type (directories first) and then by name.
	sort.Slice(children, func(i, j int) bool {
		if children[i].Type != children[j].Type {
			return children[i].Type == TypeDirectory
		}

		return children[i].Name < children[j].Name
	})

	return children
}
