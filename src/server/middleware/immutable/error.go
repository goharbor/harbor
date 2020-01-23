package immutable

import "fmt"

// ErrImmutable ...
type ErrImmutable struct {
	repo string
	tag  string
}

// Error ...
func (ei *ErrImmutable) Error() string {
	return fmt.Sprintf("Failed to process request due to '%s:%s' configured as immutable.", ei.repo, ei.tag)
}

// Unwrap ...
func (ei *ErrImmutable) Unwrap() error {
	return nil
}

// NewErrImmutable ...
func NewErrImmutable(msg, tag string) error {
	return &ErrImmutable{
		repo: msg,
		tag:  tag,
	}
}
