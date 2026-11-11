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
	"testing"

	"github.com/goharbor/harbor/src/lib/errors"
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

			result, err := untar(&buf, defaultFileSizeLimit)
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

func TestUntarExceedsLimit(t *testing.T) {
	// A single tar entry whose content exceeds the limit must be rejected
	// before it is fully materialized in memory. Note: Go's tar writer cannot
	// emit GNU sparse entries, so this test writes the full zero-filled
	// payload; on the read side untar behaves identically for a sparse entry
	// declaring the same logical size (the reader expands holes to zeros), so
	// this covers the sparse-bomb scenario as well.
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	const logicalSize = int64(defaultFileSizeLimit) * 4
	if err := tw.WriteHeader(&tar.Header{
		Name: "sparse.bin",
		Mode: 0600,
		Size: logicalSize,
	}); err != nil {
		t.Fatal(err)
	}
	// Write the declared number of bytes so the archive is well-formed; the
	// point is that untar must stop at the limit regardless of entry size.
	if _, err := io.CopyN(tw, zeroReader{}, logicalSize); err != nil {
		t.Fatal(err)
	}
	tw.Close()

	_, err := untar(&buf, defaultFileSizeLimit)
	if err == nil {
		t.Fatal("expected untar to reject content exceeding the size limit")
	}
	if !errors.IsErr(err, errors.RequestEntityTooLargeCode) {
		t.Errorf("expected RequestEntityTooLarge error, got %v", err)
	}
}

func TestUntarOversizedMetadataInNext(t *testing.T) {
	// tar.Reader.Next() materializes PAX extended records in memory before
	// untar's per-entry content copy runs, and those bytes never count against
	// the logical content limit. Many zero-size entries carrying large PAX
	// records keep the logical content at 0 bytes while the raw stream grows
	// unboundedly; the raw-stream limit must stop it.
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	chunk := strings.Repeat("x", 256*1024) // 256KB per record, under the reader's 1MB special-file cap
	for i := 0; int64(buf.Len()) <= int64(defaultFileSizeLimit)+maxTarMetadataOverhead; i++ {
		if err := tw.WriteHeader(&tar.Header{
			Name:       fmt.Sprintf("empty-%d.txt", i),
			Mode:       0600,
			Size:       0,
			Format:     tar.FormatPAX,
			PAXRecords: map[string]string{"HARBOR.padding": chunk},
		}); err != nil {
			t.Fatal(err)
		}
	}
	tw.Close()

	_, err := untar(&buf, defaultFileSizeLimit)
	if err == nil {
		t.Fatal("expected untar to reject oversized tar metadata")
	}
	if !errors.IsErr(err, errors.RequestEntityTooLargeCode) {
		t.Errorf("expected RequestEntityTooLarge error, got %v", err)
	}
}

func TestUntarNegativeLimit(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := tw.WriteHeader(&tar.Header{Name: "a.txt", Mode: 0600, Size: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("a")); err != nil {
		t.Fatal(err)
	}
	tw.Close()

	_, err := untar(&buf, -1)
	if err == nil {
		t.Fatal("expected untar to reject a negative limit")
	}
	if errors.IsErr(err, errors.RequestEntityTooLargeCode) {
		t.Errorf("negative limit should be a programmer error, not RequestEntityTooLarge, got %v", err)
	}
}

func TestUntarMultipleEntriesExceedLimit(t *testing.T) {
	// Many small entries individually under the limit must still be rejected
	// once their cumulative size crosses the limit.
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	chunk := bytes.Repeat([]byte("a"), 1024*1024) // 1MB each
	for i := 0; i < 8; i++ {                      // 8MB total > 4MB limit
		if err := tw.WriteHeader(&tar.Header{
			Name: fmt.Sprintf("file-%d.txt", i),
			Mode: 0600,
			Size: int64(len(chunk)),
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(chunk); err != nil {
			t.Fatal(err)
		}
	}
	tw.Close()

	_, err := untar(&buf, defaultFileSizeLimit)
	if err == nil {
		t.Fatal("expected untar to reject cumulative content exceeding the size limit")
	}
	if !errors.IsErr(err, errors.RequestEntityTooLargeCode) {
		t.Errorf("expected RequestEntityTooLarge error, got %v", err)
	}
}

// zeroReader is an infinite source of zero bytes.
type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
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
