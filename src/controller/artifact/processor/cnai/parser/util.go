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
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
)

func untar(reader io.Reader) ([]byte, error) {
	tr := tar.NewReader(reader)
	var buf bytes.Buffer
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to read tar header: %w", err)
		}

		// skip the directory.
		if header.Typeflag == tar.TypeDir {
			continue
		}

		if _, err := io.Copy(&buf, tr); err != nil {
			return nil, fmt.Errorf("failed to copy content to buffer: %w", err)
		}
	}

	return buf.Bytes(), nil
}

// FileType represents the type of a file.
type FileType = string

const (
	TypeFile      FileType = "file"
	TypeDirectory FileType = "directory"
)

type FileNode struct {
	Name     string
	Type     FileType
	Size     int64
	Children map[string]*FileNode
	mu       sync.RWMutex
}

func NewFile(name string, size int64) *FileNode {
	return &FileNode{
		Name: name,
		Type: TypeFile,
		Size: size,
	}
}

func NewDirectory(name string) *FileNode {
	return &FileNode{
		Name:     name,
		Type:     TypeDirectory,
		Children: make(map[string]*FileNode),
	}
}

func (root *FileNode) AddChild(child *FileNode) error {
	root.mu.Lock()
	defer root.mu.Unlock()

	if root.Type != TypeDirectory {
		return fmt.Errorf("cannot add child to non-directory node")
	}

	root.Children[child.Name] = child
	return nil
}

func (root *FileNode) GetChild(name string) (*FileNode, bool) {
	root.mu.RLock()
	defer root.mu.RUnlock()

	child, ok := root.Children[name]
	return child, ok
}

func (root *FileNode) AddNode(path string, size int64, isDir bool) (*FileNode, error) {
	path = filepath.Clean(path)
	parts := strings.Split(path, string(filepath.Separator))

	current := root
	for i, part := range parts {
		if part == "" {
			continue
		}

		isLastPart := i == len(parts)-1
		child, exists := current.GetChild(part)
		if !exists {
			var newNode *FileNode
			if isLastPart {
				if isDir {
					newNode = NewDirectory(part)
				} else {
					newNode = NewFile(part, size)
				}
			} else {
				newNode = NewDirectory(part)
			}

			if err := current.AddChild(newNode); err != nil {
				return nil, err
			}

			current = newNode
		} else {
			child.mu.RLock()
			nodeType := child.Type
			child.mu.RUnlock()

			if isLastPart {
				if (isDir && nodeType != TypeDirectory) || (!isDir && nodeType != TypeFile) {
					return nil, fmt.Errorf("path conflicts: %s exists with different type", part)
				}
			}

			current = child
		}
	}

	return current, nil
}
