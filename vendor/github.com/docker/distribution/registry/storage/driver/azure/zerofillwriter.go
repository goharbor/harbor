package azure

import (
	"bytes"
	"io"
)

type blockBlobWriter interface {
	GetSize(container, blob string) (int64, error)
	WriteBlobAt(container, blob string, offset int64, chunk io.Reader) (int64, error)
}

// zeroFillWriter enables writing to an offset outside a block blob's size
// by offering the chunk to the underlying writer as a contiguous data with
// the gap in between filled with NUL (zero) bytes.
type zeroFillWriter struct {
	blockBlobWriter
}

func newZeroFillWriter(b blockBlobWriter) zeroFillWriter {
	w := zeroFillWriter{}
	w.blockBlobWriter = b
	return w
}

// Write writes the given chunk to the specified existing blob even though
// offset is out of blob's size. The gaps are filled with zeros. Returned
// written number count does not include zeros written.
func (z *zeroFillWriter) Write(container, blob string, offset int64, chunk io.Reader) (int64, error) {
	size, err := z.blockBlobWriter.GetSize(container, blob)
	if err != nil {
		return 0, err
	}

	var reader io.Reader
	var zeroPadding int64
	if offset <= size {
		reader = chunk
	} else {
		zeroPadding = offset - size
		offset = size // adjust offset to be the append index
		zeros := bytes.NewReader(make([]byte, zeroPadding))
		reader = io.MultiReader(zeros, chunk)
	}

	nn, err := z.blockBlobWriter.WriteBlobAt(container, blob, offset, reader)
	nn -= zeroPadding
	return nn, err
}
