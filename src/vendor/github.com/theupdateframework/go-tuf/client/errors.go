package client

import (
	"errors"
	"fmt"
)

var (
	ErrNoRootKeys       = errors.New("tuf: no root keys found in local meta store")
	ErrInsufficientKeys = errors.New("tuf: insufficient keys to meet threshold")
	ErrNoLocalSnapshot  = errors.New("tuf: no snapshot stored locally")
)

type ErrMissingRemoteMetadata struct {
	Name string
}

func (e ErrMissingRemoteMetadata) Error() string {
	return fmt.Sprintf("tuf: missing remote metadata %s", e.Name)
}

type ErrDownloadFailed struct {
	File string
	Err  error
}

func (e ErrDownloadFailed) Error() string {
	return fmt.Sprintf("tuf: failed to download %s: %s", e.File, e.Err)
}

type ErrDecodeFailed struct {
	File string
	Err  error
}

func (e ErrDecodeFailed) Error() string {
	return fmt.Sprintf("tuf: failed to decode %s: %s", e.File, e.Err)
}

type ErrMaxDelegations struct {
	Target          string
	MaxDelegations  int
	SnapshotVersion int64
}

func (e ErrMaxDelegations) Error() string {
	return fmt.Sprintf("tuf: max delegation of %d reached searching for %s with snapshot version %d", e.MaxDelegations, e.Target, e.SnapshotVersion)
}

type ErrNotFound struct {
	File string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("tuf: file not found: %s", e.File)
}

func IsNotFound(err error) bool {
	_, ok := err.(ErrNotFound)
	return ok
}

type ErrWrongSize struct {
	File     string
	Actual   int64
	Expected int64
}

func (e ErrWrongSize) Error() string {
	return fmt.Sprintf("tuf: unexpected file size: %s (expected %d bytes, got %d bytes)", e.File, e.Expected, e.Actual)
}

type ErrUnknownTarget struct {
	Name            string
	SnapshotVersion int64
}

func (e ErrUnknownTarget) Error() string {
	return fmt.Sprintf("tuf: unknown target file: %s with snapshot version %d", e.Name, e.SnapshotVersion)
}

type ErrMetaTooLarge struct {
	Name    string
	Size    int64
	MaxSize int64
}

func (e ErrMetaTooLarge) Error() string {
	return fmt.Sprintf("tuf: %s size %d bytes greater than maximum %d bytes", e.Name, e.Size, e.MaxSize)
}

type ErrInvalidURL struct {
	URL string
}

func (e ErrInvalidURL) Error() string {
	return fmt.Sprintf("tuf: invalid repository URL %s", e.URL)
}

type ErrRoleNotInSnapshot struct {
	Role            string
	SnapshotVersion int64
}

func (e ErrRoleNotInSnapshot) Error() string {
	return fmt.Sprintf("tuf: role %s not in snapshot version %d", e.Role, e.SnapshotVersion)
}
