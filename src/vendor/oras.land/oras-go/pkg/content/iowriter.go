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
	"context"
	"io"
	"io/ioutil"

	"github.com/containerd/containerd/content"
	"github.com/opencontainers/go-digest"
)

// IoContentWriter writer that wraps an io.Writer, so the results can be streamed to
// an open io.Writer. For example, can be used to pull a layer and write it to a file, or device.
type IoContentWriter struct {
	writer   io.Writer
	digester digest.Digester
	size     int64
	hash     *digest.Digest
}

// NewIoContentWriter create a new IoContentWriter.
//
// By default, it calculates the hash when writing. If the option `skipHash` is true,
// it will skip doing the hash. Skipping the hash is intended to be used only
// if you are confident about the validity of the data being passed to the writer,
// and wish to save on the hashing time.
func NewIoContentWriter(writer io.Writer, opts ...WriterOpt) content.Writer {
	w := writer
	if w == nil {
		w = ioutil.Discard
	}
	// process opts for default
	wOpts := DefaultWriterOpts()
	for _, opt := range opts {
		if err := opt(&wOpts); err != nil {
			return nil
		}
	}
	ioc := &IoContentWriter{
		writer:   w,
		digester: digest.Canonical.Digester(),
		// we take the OutputHash, since the InputHash goes to the passthrough writer,
		// which then passes the processed output to us
		hash: wOpts.OutputHash,
	}
	return NewPassthroughWriter(ioc, func(r io.Reader, w io.Writer, done chan<- error) {
		// write out the data to the io writer
		var (
			err error
		)
		// we could use io.Copy, but calling it with the default blocksize is identical to
		// io.CopyBuffer. Otherwise, we would need some way to let the user flag "I want to use
		// io.Copy", when it should not matter to them
		b := make([]byte, wOpts.Blocksize, wOpts.Blocksize)
		_, err = io.CopyBuffer(w, r, b)
		done <- err
	}, opts...)
}

func (w *IoContentWriter) Write(p []byte) (n int, err error) {
	n, err = w.writer.Write(p)
	if err != nil {
		return 0, err
	}
	w.size += int64(n)
	if w.hash == nil {
		w.digester.Hash().Write(p[:n])
	}
	return
}

func (w *IoContentWriter) Close() error {
	return nil
}

// Digest may return empty digest or panics until committed.
func (w *IoContentWriter) Digest() digest.Digest {
	return w.digester.Digest()
}

// Commit commits the blob (but no roll-back is guaranteed on an error).
// size and expected can be zero-value when unknown.
// Commit always closes the writer, even on error.
// ErrAlreadyExists aborts the writer.
func (w *IoContentWriter) Commit(ctx context.Context, size int64, expected digest.Digest, opts ...content.Opt) error {
	return nil
}

// Status returns the current state of write
func (w *IoContentWriter) Status() (content.Status, error) {
	return content.Status{}, nil
}

// Truncate updates the size of the target blob
func (w *IoContentWriter) Truncate(size int64) error {
	return nil
}
