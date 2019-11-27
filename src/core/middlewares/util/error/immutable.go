package error

import (
	"fmt"
)

// ErrImmutable ...
type ErrImmutable struct {
	repo string
	tag  string
}

// Error ...
func (ei ErrImmutable) Error() string {
	return fmt.Sprintf("Failed to process request due to '%s:%s' configured as immutable.", ei.repo, ei.tag)
}

// NewErrImmutable ...
func NewErrImmutable(msg, tag string) ErrImmutable {
	return ErrImmutable{
		repo: msg,
		tag:  tag,
	}
}
