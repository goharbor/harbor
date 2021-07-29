package oras

import (
	"errors"
	"fmt"
)

// Common errors
var (
	ErrResolverUndefined = errors.New("resolver undefined")
)

// Path validation related errors
var (
	ErrDirtyPath               = errors.New("dirty path")
	ErrPathNotSlashSeparated   = errors.New("path not slash separated")
	ErrAbsolutePathDisallowed  = errors.New("absolute path disallowed")
	ErrPathTraversalDisallowed = errors.New("path traversal disallowed")
)

// ErrStopProcessing is used to stop processing an oras operation.
// This error only makes sense in sequential pulling operation.
var ErrStopProcessing = fmt.Errorf("stop processing")
