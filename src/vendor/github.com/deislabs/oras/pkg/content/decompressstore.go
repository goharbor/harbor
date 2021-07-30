package content

import (
	"context"
	"errors"
	"strings"

	ctrcontent "github.com/containerd/containerd/content"
)

// DecompressStore store to decompress content and extract from tar, if needed, wrapping
// another store. By default, a FileStore will simply take each artifact and write it to
// a file, as a MemoryStore will do into memory. If the artifact is gzipped or tarred,
// you might want to store the actual object inside tar or gzip. Wrap your Store
// with DecompressStore, and it will check the media-type and, if relevant,
// gunzip and/or untar.
//
// For example:
//
//        fileStore := NewFileStore(rootPath)
//        decompressStore := store.NewDecompressStore(fileStore, WithBlocksize(blocksize))
//
// The above example works if there is no tar, i.e. each artifact is just a single file, perhaps gzipped,
// or if there is only one file in each tar archive. In other words, when each content.Writer has only one target output stream.
// However, if you have multiple files in each tar archive, each archive of which is an artifact layer, then
// you need a way to select how to handle each file in the tar archive. In other words, when each content.Writer has more than one
// target output stream. In that case, use the following example:
//
//        multiStore := NewMultiStore(rootPath) // some store that can handle different filenames
//        decompressStore := store.NewDecompressStore(multiStore, WithBlocksize(blocksize), WithMultiWriterIngester())
//
type DecompressStore struct {
	ingester            ctrcontent.Ingester
	blocksize           int
	multiWriterIngester bool
}

func NewDecompressStore(ingester ctrcontent.Ingester, opts ...WriterOpt) DecompressStore {
	// we have to reprocess the opts to find the blocksize
	var wOpts WriterOpts
	for _, opt := range opts {
		if err := opt(&wOpts); err != nil {
			// TODO: we probably should handle errors here
			continue
		}
	}

	return DecompressStore{ingester, wOpts.Blocksize, wOpts.MultiWriterIngester}
}

// Writer get a writer
func (d DecompressStore) Writer(ctx context.Context, opts ...ctrcontent.WriterOpt) (ctrcontent.Writer, error) {
	// the logic is straightforward:
	// - if there is a desc in the opts, and the mediatype is tar or tar+gzip, then pass the correct decompress writer
	// - else, pass the regular writer
	var (
		writer        ctrcontent.Writer
		err           error
		multiIngester MultiWriterIngester
		ok            bool
	)

	// check to see if we are supposed to use a MultiWriterIngester
	if d.multiWriterIngester {
		multiIngester, ok = d.ingester.(MultiWriterIngester)
		if !ok {
			return nil, errors.New("configured to use multiwriter ingester, but ingester does not implement multiwriter")
		}
	}

	// we have to reprocess the opts to find the desc
	var wOpts ctrcontent.WriterOpts
	for _, opt := range opts {
		if err := opt(&wOpts); err != nil {
			return nil, err
		}
	}
	// figure out if compression and/or archive exists
	desc := wOpts.Desc
	// before we pass it down, we need to strip anything we are removing here
	// and possibly update the digest, since the store indexes things by digest
	hasGzip, hasTar, modifiedMediaType := checkCompression(desc.MediaType)
	wOpts.Desc.MediaType = modifiedMediaType
	opts = append(opts, ctrcontent.WithDescriptor(wOpts.Desc))
	// determine if we pass it blocksize, only if positive
	writerOpts := []WriterOpt{}
	if d.blocksize > 0 {
		writerOpts = append(writerOpts, WithBlocksize(d.blocksize))
	}

	writer, err = d.ingester.Writer(ctx, opts...)
	if err != nil {
		return nil, err
	}

	// do we need to wrap with an untar writer?
	if hasTar {
		// if not multiingester, get a regular writer
		if multiIngester == nil {
			writer = NewUntarWriter(writer, writerOpts...)
		} else {
			writers, err := multiIngester.Writers(ctx, opts...)
			if err != nil {
				return nil, err
			}
			writer = NewUntarWriterByName(writers, writerOpts...)
		}
	}
	if hasGzip {
		if writer == nil {
			writer, err = d.ingester.Writer(ctx, opts...)
			if err != nil {
				return nil, err
			}
		}
		writer = NewGunzipWriter(writer, writerOpts...)
	}
	return writer, nil
}

// checkCompression check if the mediatype uses gzip compression or tar.
// Returns if it has gzip and/or tar, as well as the base media type without
// those suffixes.
func checkCompression(mediaType string) (gzip, tar bool, mt string) {
	mt = mediaType
	gzipSuffix := "+gzip"
	gzipAltSuffix := ".gzip"
	tarSuffix := ".tar"
	switch {
	case strings.HasSuffix(mt, gzipSuffix):
		mt = mt[:len(mt)-len(gzipSuffix)]
		gzip = true
	case strings.HasSuffix(mt, gzipAltSuffix):
		mt = mt[:len(mt)-len(gzipAltSuffix)]
		gzip = true
	}

	if strings.HasSuffix(mt, tarSuffix) {
		mt = mt[:len(mt)-len(tarSuffix)]
		tar = true
	}
	return
}
