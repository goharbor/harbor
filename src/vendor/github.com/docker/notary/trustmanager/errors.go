package trustmanager

import "fmt"

// ErrAttemptsExceeded is returned when too many attempts have been made to decrypt a key
type ErrAttemptsExceeded struct{}

// ErrAttemptsExceeded is returned when too many attempts have been made to decrypt a key
func (err ErrAttemptsExceeded) Error() string {
	return "maximum number of passphrase attempts exceeded"
}

// ErrPasswordInvalid is returned when signing fails. It could also mean the signing
// key file was corrupted, but we have no way to distinguish.
type ErrPasswordInvalid struct{}

// ErrPasswordInvalid is returned when signing fails. It could also mean the signing
// key file was corrupted, but we have no way to distinguish.
func (err ErrPasswordInvalid) Error() string {
	return "password invalid, operation has failed."
}

// ErrKeyNotFound is returned when the keystore fails to retrieve a specific key.
type ErrKeyNotFound struct {
	KeyID string
}

// ErrKeyNotFound is returned when the keystore fails to retrieve a specific key.
func (err ErrKeyNotFound) Error() string {
	return fmt.Sprintf("signing key not found: %s", err.KeyID)
}
