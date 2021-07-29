package content

import (
	"errors"

	"github.com/opencontainers/go-digest"
)

type WriterOpts struct {
	InputHash           *digest.Digest
	OutputHash          *digest.Digest
	Blocksize           int
	MultiWriterIngester bool
	IgnoreNoName        bool
}

type WriterOpt func(*WriterOpts) error

func DefaultWriterOpts() WriterOpts {
	return WriterOpts{
		InputHash:    nil,
		OutputHash:   nil,
		Blocksize:    DefaultBlocksize,
		IgnoreNoName: true,
	}
}

// WithInputHash provide the expected input hash to a writer. Writers
// may suppress their own calculation of a hash on the stream, taking this
// hash instead. If the Writer processes the data before passing it on to another
// Writer layer, this is the hash of the *input* stream.
//
// To have a blank hash, use WithInputHash(BlankHash).
func WithInputHash(hash digest.Digest) WriterOpt {
	return func(w *WriterOpts) error {
		w.InputHash = &hash
		return nil
	}
}

// WithOutputHash provide the expected output hash to a writer. Writers
// may suppress their own calculation of a hash on the stream, taking this
// hash instead. If the Writer processes the data before passing it on to another
// Writer layer, this is the hash of the *output* stream.
//
// To have a blank hash, use WithInputHash(BlankHash).
func WithOutputHash(hash digest.Digest) WriterOpt {
	return func(w *WriterOpts) error {
		w.OutputHash = &hash
		return nil
	}
}

// WithBlocksize set the blocksize used by the processor of data.
// The default is DefaultBlocksize, which is the same as that used by io.Copy.
// Includes a safety check to ensure the caller doesn't actively set it to <= 0.
func WithBlocksize(blocksize int) WriterOpt {
	return func(w *WriterOpts) error {
		if blocksize <= 0 {
			return errors.New("blocksize must be greater than or equal to 0")
		}
		w.Blocksize = blocksize
		return nil
	}
}

// WithMultiWriterIngester the passed ingester also implements MultiWriter
// and should be used as such. If this is set to true, but the ingester does not
// implement MultiWriter, calling Writer should return an error.
func WithMultiWriterIngester() WriterOpt {
	return func(w *WriterOpts) error {
		w.MultiWriterIngester = true
		return nil
	}
}

// WithErrorOnNoName some ingesters, when creating a Writer, do not return an error if
// the descriptor does not have a valid name on the descriptor. Passing WithErrorOnNoName
// tells the writer to return an error instead of passing the data to a nil writer.
func WithErrorOnNoName() WriterOpt {
	return func(w *WriterOpts) error {
		w.IgnoreNoName = false
		return nil
	}
}

// WithIgnoreNoName some ingesters, when creating a Writer, return an error if
// the descriptor does not have a valid name on the descriptor. Passing WithIgnoreNoName
// tells the writer not to return an error, but rather to pass the data to a nil writer.
//
// Deprecated: Use WithErrorOnNoName
func WithIgnoreNoName() WriterOpt {
	return func(w *WriterOpts) error {
		w.IgnoreNoName = true
		return nil
	}
}
