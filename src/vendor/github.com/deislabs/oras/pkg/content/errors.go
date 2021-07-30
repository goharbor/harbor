package content

import "errors"

// Common errors
var (
	ErrNotFound           = errors.New("not_found")
	ErrNoName             = errors.New("no_name")
	ErrUnsupportedSize    = errors.New("unsupported_size")
	ErrUnsupportedVersion = errors.New("unsupported_version")
)

// FileStore errors
var (
	ErrPathTraversalDisallowed = errors.New("path_traversal_disallowed")
	ErrOverwriteDisallowed     = errors.New("overwrite_disallowed")
)
