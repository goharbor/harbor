package error

import (
	"fmt"
)

// ErrImmutable ...
type ErrImmutable struct {
	repo string
}

// Error ...
func (ei ErrImmutable) Error() string {
	return fmt.Sprintf("Failed to process request, due to immutable. '%s'", ei.repo)
}

// NewErrImmutable ...
func NewErrImmutable(msg string) ErrImmutable {
	return ErrImmutable{repo: msg}
}
