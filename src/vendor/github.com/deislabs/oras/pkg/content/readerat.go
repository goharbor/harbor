package content

import (
	"io"

	"github.com/containerd/containerd/content"
)

// ensure interface
var (
	_ content.ReaderAt = sizeReaderAt{}
)

type readAtCloser interface {
	io.ReaderAt
	io.Closer
}

type sizeReaderAt struct {
	readAtCloser
	size int64
}

func (ra sizeReaderAt) Size() int64 {
	return ra.size
}

type nopCloser struct {
	io.ReaderAt
}

func (nopCloser) Close() error {
	return nil
}
