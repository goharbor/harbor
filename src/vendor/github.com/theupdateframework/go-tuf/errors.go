package tuf

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInitNotAllowed               = errors.New("tuf: repository already initialized")
	ErrNewRepository                = errors.New("tuf: repository not yet committed")
	ErrChangePassphraseNotSupported = errors.New("tuf: store does not support changing passphrase")
)

type ErrMissingMetadata struct {
	Name string
}

func (e ErrMissingMetadata) Error() string {
	return fmt.Sprintf("tuf: missing metadata file %s", e.Name)
}

type ErrFileNotFound struct {
	Path string
}

func (e ErrFileNotFound) Error() string {
	return fmt.Sprintf("tuf: file not found %s", e.Path)
}

type ErrNoKeys struct {
	Name string
}

func (e ErrNoKeys) Error() string {
	return fmt.Sprintf("tuf: no keys available to sign %s", e.Name)
}

type ErrInsufficientSignatures struct {
	Name string
	Err  error
}

func (e ErrInsufficientSignatures) Error() string {
	return fmt.Sprintf("tuf: insufficient signatures for %s: %s", e.Name, e.Err)
}

type ErrInvalidRole struct {
	Role   string
	Reason string
}

func (e ErrInvalidRole) Error() string {
	return fmt.Sprintf("tuf: invalid role %s: %s", e.Role, e.Reason)
}

type ErrInvalidExpires struct {
	Expires time.Time
}

func (e ErrInvalidExpires) Error() string {
	return fmt.Sprintf("tuf: invalid expires: %s", e.Expires)
}

type ErrKeyNotFound struct {
	Role  string
	KeyID string
}

func (e ErrKeyNotFound) Error() string {
	return fmt.Sprintf(`tuf: no key with id "%s" exists for the %s role`, e.KeyID, e.Role)
}

type ErrNotEnoughKeys struct {
	Role      string
	Keys      int
	Threshold int
}

func (e ErrNotEnoughKeys) Error() string {
	return fmt.Sprintf("tuf: %s role has insufficient keys for threshold (has %d keys, threshold is %d)", e.Role, e.Keys, e.Threshold)
}

type ErrPassphraseRequired struct {
	Role string
}

func (e ErrPassphraseRequired) Error() string {
	return fmt.Sprintf("tuf: a passphrase is required to access the encrypted %s keys file", e.Role)
}

type ErrNoDelegatedTarget struct {
	Path string
}

func (e ErrNoDelegatedTarget) Error() string {
	return fmt.Sprintf("tuf: no delegated target for path %s", e.Path)
}
