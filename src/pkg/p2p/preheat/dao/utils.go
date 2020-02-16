package dao

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

// ErrObjectNotFound represents the not found error
var ErrObjectNotFound = errors.New("object not found")

// UUID generates an unique ID
func UUID() string {
	b := make([]byte, 12)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", b)
}
