// Copyright 2015 clair authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
)

var (
	// ErrCouldNotExtract occurs when an extraction fails.
	ErrCouldNotExtract = errors.New("utils: could not extract the archive")

	// ErrExtractedFileTooBig occurs when a file to extract is too big.
	ErrExtractedFileTooBig = errors.New("utils: could not extract one or more files from the archive: file too big")

	readLen = 6 // max bytes to sniff

	gzipHeader  = []byte{0x1f, 0x8b}
	bzip2Header = []byte{0x42, 0x5a, 0x68}
	xzHeader    = []byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00}
)

// XzReader is an io.ReadCloser which decompresses xz compressed data.
type XzReader struct {
	io.ReadCloser
	cmd     *exec.Cmd
	closech chan error
}

// NewXzReader shells out to a command line xz executable (if
// available) to decompress the given io.Reader using the xz
// compression format and returns an *XzReader.
// It is the caller's responsibility to call Close on the XzReader when done.
func NewXzReader(r io.Reader) (*XzReader, error) {
	rpipe, wpipe := io.Pipe()
	ex, err := exec.LookPath("xz")
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(ex, "--decompress", "--stdout")

	closech := make(chan error)

	cmd.Stdin = r
	cmd.Stdout = wpipe

	go func() {
		err := cmd.Run()
		wpipe.CloseWithError(err)
		closech <- err
	}()

	return &XzReader{rpipe, cmd, closech}, nil
}

func (r *XzReader) Close() error {
	r.ReadCloser.Close()
	r.cmd.Process.Kill()
	return <-r.closech
}

// TarReadCloser embeds a *tar.Reader and the related io.Closer
// It is the caller's responsibility to call Close on TarReadCloser when
// done.
type TarReadCloser struct {
	*tar.Reader
	io.Closer
}

func (r *TarReadCloser) Close() error {
	return r.Closer.Close()
}

// SelectivelyExtractArchive extracts the specified files and folders
// from targz data read from the given reader and store them in a map indexed by file paths
func SelectivelyExtractArchive(r io.Reader, prefix string, toExtract []string, maxFileSize int64) (map[string][]byte, error) {
	data := make(map[string][]byte)

	// Create a tar or tar/tar-gzip/tar-bzip2/tar-xz reader
	tr, err := getTarReader(r)
	if err != nil {
		return data, ErrCouldNotExtract
	}
	defer tr.Close()

	// For each element in the archive
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return data, ErrCouldNotExtract
		}

		// Get element filename
		filename := hdr.Name
		filename = strings.TrimPrefix(filename, "./")
		if prefix != "" {
			filename = strings.TrimPrefix(filename, prefix)
		}

		// Determine if we should extract the element
		toBeExtracted := false
		for _, s := range toExtract {
			if strings.HasPrefix(filename, s) {
				toBeExtracted = true
				break
			}
		}

		if toBeExtracted {
			// File size limit
			if maxFileSize > 0 && hdr.Size > maxFileSize {
				return data, ErrExtractedFileTooBig
			}

			// Extract the element
			if hdr.Typeflag == tar.TypeSymlink || hdr.Typeflag == tar.TypeLink || hdr.Typeflag == tar.TypeReg {
				d, _ := ioutil.ReadAll(tr)
				data[filename] = d
			}
		}
	}

	return data, nil
}

// getTarReader returns a TarReaderCloser associated with the specified io.Reader.
//
// Gzip/Bzip2/XZ detection is done by using the magic numbers:
// Gzip: the first two bytes should be 0x1f and 0x8b. Defined in the RFC1952.
// Bzip2: the first three bytes should be 0x42, 0x5a and 0x68. No RFC.
// XZ: the first three bytes should be 0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00. No RFC.
func getTarReader(r io.Reader) (*TarReadCloser, error) {
	br := bufio.NewReader(r)
	header, err := br.Peek(readLen)
	if err == nil {
		switch {
		case bytes.HasPrefix(header, gzipHeader):
			gr, err := gzip.NewReader(br)
			if err != nil {
				return nil, err
			}
			return &TarReadCloser{tar.NewReader(gr), gr}, nil
		case bytes.HasPrefix(header, bzip2Header):
			bzip2r := ioutil.NopCloser(bzip2.NewReader(br))
			return &TarReadCloser{tar.NewReader(bzip2r), bzip2r}, nil
		case bytes.HasPrefix(header, xzHeader):
			xzr, err := NewXzReader(br)
			if err != nil {
				return nil, err
			}
			return &TarReadCloser{tar.NewReader(xzr), xzr}, nil
		}
	}

	dr := ioutil.NopCloser(br)
	return &TarReadCloser{tar.NewReader(dr), dr}, nil
}
