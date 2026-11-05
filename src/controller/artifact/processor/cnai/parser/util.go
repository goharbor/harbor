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

package parser // nolint:revive

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"

	"github.com/goharbor/harbor/src/lib/errors"
)

// maxTarMetadataOverhead is the extra budget of raw tar stream bytes allowed
// on top of the logical content limit, to account for tar headers (including
// PAX extended records and GNU sparse maps), block padding and end-of-archive
// markers of a legitimate archive.
const maxTarMetadataOverhead = 1 << 20 // 1MB

// limitedReader is like io.LimitReader, but returns a RequestEntityTooLarge
// error instead of io.EOF once more than remaining bytes have been consumed,
// so that hitting the limit is distinguishable from a legitimate end of
// stream.
type limitedReader struct {
	r         io.Reader
	remaining int64
}

func (l *limitedReader) Read(p []byte) (int, error) {
	if l.remaining <= 0 {
		return 0, errors.RequestEntityTooLargeError(errFileTooLarge)
	}
	if int64(len(p)) > l.remaining {
		p = p[:l.remaining]
	}
	n, err := l.r.Read(p)
	l.remaining -= int64(n)
	return n, err
}

// untar reads all file entries from the tar stream and returns their
// concatenated content. It enforces that the total number of bytes written
// across all entries does not exceed limit, so that a tar declaring a small
// blob size but expanding to a huge payload (e.g. GNU sparse files) cannot
// exhaust memory.
func untar(reader io.Reader, limit int64) ([]byte, error) {
	if limit < 0 {
		return nil, fmt.Errorf("invalid limit: %d", limit)
	}

	// Cap the raw bytes consumed from the underlying stream as well: the
	// manifest-declared blob size can lie, and tar.Reader.Next() materializes
	// metadata (PAX extended records, GNU sparse maps) in memory before the
	// per-entry content copy below ever sees it.
	tr := tar.NewReader(&limitedReader{r: reader, remaining: limit + maxTarMetadataOverhead})
	var buf bytes.Buffer
	remaining := limit
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

		// Copy at most remaining+1 bytes so we can detect when the cumulative
		// content exceeds the limit without trusting the declared header size.
		n, err := io.Copy(&buf, io.LimitReader(tr, remaining+1))
		if err != nil {
			return nil, fmt.Errorf("failed to copy content to buffer: %w", err)
		}

		remaining -= n
		if remaining < 0 {
			return nil, errors.RequestEntityTooLargeError(errFileTooLarge)
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
