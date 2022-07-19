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
	"compress/gzip"
	"fmt"
	"io"

	"github.com/containerd/containerd/content"
)

// NewGunzipWriter wrap a writer with a gunzip, so that the stream is gunzipped
//
// By default, it calculates the hash when writing. If the option `skipHash` is true,
// it will skip doing the hash. Skipping the hash is intended to be used only
// if you are confident about the validity of the data being passed to the writer,
// and wish to save on the hashing time.
func NewGunzipWriter(writer content.Writer, opts ...WriterOpt) content.Writer {
	// process opts for default
	wOpts := DefaultWriterOpts()
	for _, opt := range opts {
		if err := opt(&wOpts); err != nil {
			return nil
		}
	}
	return NewPassthroughWriter(writer, func(r io.Reader, w io.Writer, done chan<- error) {
		gr, err := gzip.NewReader(r)
		if err != nil {
			done <- fmt.Errorf("error creating gzip reader: %v", err)
			return
		}
		// write out the uncompressed data
		b := make([]byte, wOpts.Blocksize, wOpts.Blocksize)
		for {
			var n int
			n, err = gr.Read(b)
			if err != nil && err != io.EOF {
				err = fmt.Errorf("GunzipWriter data read error: %v\n", err)
				break
			}
			l := n
			if n > len(b) {
				l = len(b)
			}
			if _, err2 := w.Write(b[:l]); err2 != nil {
				err = fmt.Errorf("GunzipWriter: error writing to underlying writer: %v", err2)
				break
			}
			if err == io.EOF {
				// clear the error
				err = nil
				break
			}
		}
		gr.Close()
		done <- err
	}, opts...)
}
