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
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

// ResolveName resolves name from descriptor
func ResolveName(desc ocispec.Descriptor) (string, bool) {
	name, ok := desc.Annotations[ocispec.AnnotationTitle]
	return name, ok
}

// tarDirectory walks the directory specified by path, and tar those files with a new
// path prefix.
func tarDirectory(root, prefix string, w io.Writer, stripTimes bool) error {
	tw := tar.NewWriter(w)
	defer tw.Close()
	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Rename path
		name, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		name = filepath.Join(prefix, name)
		name = filepath.ToSlash(name)

		// Generate header
		var link string
		mode := info.Mode()
		if mode&os.ModeSymlink != 0 {
			if link, err = os.Readlink(path); err != nil {
				return err
			}
		}
		header, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return errors.Wrap(err, path)
		}
		header.Name = name
		header.Uid = 0
		header.Gid = 0
		header.Uname = ""
		header.Gname = ""

		if stripTimes {
			header.ModTime = time.Time{}
			header.AccessTime = time.Time{}
			header.ChangeTime = time.Time{}
		}

		// Write file
		if err := tw.WriteHeader(header); err != nil {
			return errors.Wrap(err, "tar")
		}
		if mode.IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err := io.Copy(tw, file); err != nil {
				return errors.Wrap(err, path)
			}
		}

		return nil
	}); err != nil {
		return err
	}
	return nil
}

// extractTarDirectory extracts tar file to a directory specified by the `root`
// parameter. The file name prefix is ensured to be the string specified by the
// `prefix` parameter and is trimmed.
func extractTarDirectory(root, prefix string, r io.Reader) error {
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		// Name check
		name := header.Name
		path, err := ensureBasePath(root, prefix, name)
		if err != nil {
			return err
		}
		path = filepath.Join(root, path)

		// Link check
		switch header.Typeflag {
		case tar.TypeLink, tar.TypeSymlink:
			link := header.Linkname
			if !filepath.IsAbs(link) {
				link = filepath.Join(filepath.Dir(name), link)
			}
			if _, err := ensureBasePath(root, prefix, link); err != nil {
				return err
			}
		}

		// Create content
		switch header.Typeflag {
		case tar.TypeReg:
			err = writeFile(path, tr, header.FileInfo().Mode())
		case tar.TypeDir:
			err = os.MkdirAll(path, header.FileInfo().Mode())
		case tar.TypeLink:
			err = os.Link(header.Linkname, path)
		case tar.TypeSymlink:
			err = os.Symlink(header.Linkname, path)
		default:
			continue // Non-regular files are skipped
		}
		if err != nil {
			return err
		}

		// Change access time and modification time if possible (error ignored)
		os.Chtimes(path, header.AccessTime, header.ModTime)
	}
}

// ensureBasePath ensures the target path is in the base path,
// returning its relative path to the base path.
func ensureBasePath(root, base, target string) (string, error) {
	path, err := filepath.Rel(base, target)
	if err != nil {
		return "", err
	}
	cleanPath := filepath.ToSlash(filepath.Clean(path))
	if cleanPath == ".." || strings.HasPrefix(cleanPath, "../") {
		return "", fmt.Errorf("%q is outside of %q", target, base)
	}

	// No symbolic link allowed in the relative path
	dir := filepath.Dir(path)
	for dir != "." {
		if info, err := os.Lstat(filepath.Join(root, dir)); err != nil {
			if !os.IsNotExist(err) {
				return "", err
			}
		} else if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("no symbolic link allowed between %q and %q", base, target)
		}
		dir = filepath.Dir(dir)
	}

	return path, nil
}

func writeFile(path string, r io.Reader, perm os.FileMode) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, r)
	return err
}

func extractTarGzip(root, prefix, filename, checksum string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	zr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer zr.Close()
	var r io.Reader = zr
	var verifier digest.Verifier
	if checksum != "" {
		if digest, err := digest.Parse(checksum); err == nil {
			verifier = digest.Verifier()
			r = io.TeeReader(r, verifier)
		}
	}
	if err := extractTarDirectory(root, prefix, r); err != nil {
		return err
	}
	if verifier != nil && !verifier.Verified() {
		return errors.New("content digest mismatch")
	}
	return nil
}
