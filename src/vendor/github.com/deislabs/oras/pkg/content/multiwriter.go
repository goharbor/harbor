package content

import (
	"context"

	ctrcontent "github.com/containerd/containerd/content"
)

// MultiWriterIngester an ingester that can provide a single writer or multiple writers for a single
// descriptor. Useful when the target of a descriptor can have multiple items within it, e.g. a layer
// that is a tar file with multiple files, each of which should go to a different stream, some of which
// should not be handled at all.
type MultiWriterIngester interface {
	ctrcontent.Ingester
	Writers(ctx context.Context, opts ...ctrcontent.WriterOpt) (func(string) (ctrcontent.Writer, error), error)
}
