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
	"path/filepath"
	"strings"
	"testing"
)

func TestUntar(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErr  bool
		expected string
	}{
		{
			name:     "valid tar file with single file",
			content:  "test content",
			wantErr:  false,
			expected: "test content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)

			hdr := &tar.Header{
				Name: "test.txt",
				Mode: 0600,
				Size: int64(len(tt.content)),
			}
			if err := tw.WriteHeader(hdr); err != nil {
				t.Fatal(err)
			}
			if _, err := tw.Write([]byte(tt.content)); err != nil {
				t.Fatal(err)
			}
			tw.Close()

			result, err := untar(&buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("untar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(result) != tt.expected {
				t.Errorf("untar() = %v, want %v", string(result), tt.expected)
			}
		})
	}
}

func TestFileNode(t *testing.T) {
	t.Run("test file node operations", func(t *testing.T) {
		// Test creating root directory.
		root := NewDirectory("root")
		if root.Type != TypeDirectory {
			t.Errorf("Expected directory type, got %s", root.Type)
		}

		// Test creating file.
		file := NewFile("test.txt", 100)
		if file.Type != TypeFile {
			t.Errorf("Expected file type, got %s", file.Type)
		}

		// Test adding child to directory.
		err := root.AddChild(file)
		if err != nil {
			t.Errorf("Failed to add child: %v", err)
		}

		// Test getting child.
		child, exists := root.GetChild("test.txt")
		if !exists {
			t.Error("Expected child to exist")
		}
		if child.Name != "test.txt" {
			t.Errorf("Expected name test.txt, got %s", child.Name)
		}

		// Test adding child to file (should fail).
		err = file.AddChild(NewFile("invalid.txt", 50))
		if err == nil {
			t.Error("Expected error when adding child to file")
		}
	})
}

func TestAddNode(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		size    int64
		isDir   bool
		wantErr bool
		setupFn func(*FileNode)
	}{
		{
			name:    "add file",
			path:    "dir1/dir2/file.txt",
			size:    100,
			isDir:   false,
			wantErr: false,
		},
		{
			name:    "add directory",
			path:    "dir1/dir2/dir3",
			size:    0,
			isDir:   true,
			wantErr: false,
		},
		{
			name:    "add file with conflicting directory",
			path:    "dir1/dir2",
			size:    100,
			isDir:   false,
			wantErr: true,
			setupFn: func(node *FileNode) {
				node.AddNode("dir1/dir2", 0, true)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := NewDirectory("root")
			if tt.setupFn != nil {
				tt.setupFn(root)
			}

			_, err := root.AddNode(tt.path, tt.size, tt.isDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the path exists.
				current := root
				parts := filepath.Clean(tt.path)
				for part := range strings.SplitSeq(parts, string(filepath.Separator)) {
					if part == "" {
						continue
					}
					child, exists := current.GetChild(part)
					if !exists {
						t.Errorf("Expected path part %s to exist", part)
						return
					}
					current = child
				}
			}
		})
	}
}
