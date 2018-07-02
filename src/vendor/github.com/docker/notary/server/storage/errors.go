package storage

import (
	"fmt"
)

// ErrOldVersion is returned when a newer version of TUF metadata is already available
type ErrOldVersion struct{}

// ErrOldVersion is returned when a newer version of TUF metadata is already available
func (err ErrOldVersion) Error() string {
	return fmt.Sprintf("Error updating metadata. A newer version is already available")
}

// ErrNotFound is returned when TUF metadata isn't found for a specific record
type ErrNotFound struct{}

// Error implements error
func (err ErrNotFound) Error() string {
	return fmt.Sprintf("No record found")
}

// ErrKeyExists is returned when a key already exists
type ErrKeyExists struct {
	gun  string
	role string
}

// ErrKeyExists is returned when a key already exists
func (err ErrKeyExists) Error() string {
	return fmt.Sprintf("Error, timestamp key already exists for %s:%s", err.gun, err.role)
}

// ErrNoKey is returned when no timestamp key is found
type ErrNoKey struct {
	gun string
}

// ErrNoKey is returned when no timestamp key is found
func (err ErrNoKey) Error() string {
	return fmt.Sprintf("Error, no timestamp key found for %s", err.gun)
}

// ErrBadQuery is used when the parameters provided cannot be appropriately
// coerced.
type ErrBadQuery struct {
	msg string
}

func (err ErrBadQuery) Error() string {
	return fmt.Sprintf("did not recognize parameters: %s", err.msg)
}
