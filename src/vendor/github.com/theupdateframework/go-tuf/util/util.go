package util

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/theupdateframework/go-tuf/data"
)

type ErrWrongLength struct {
	Expected int64
	Actual   int64
}

func (e ErrWrongLength) Error() string {
	return fmt.Sprintf("wrong length, expected %d got %d", e.Expected, e.Actual)
}

type ErrWrongVersion struct {
	Expected int64
	Actual   int64
}

func (e ErrWrongVersion) Error() string {
	return fmt.Sprintf("wrong version, expected %d got %d", e.Expected, e.Actual)
}

type ErrWrongHash struct {
	Type     string
	Expected data.HexBytes
	Actual   data.HexBytes
}

func (e ErrWrongHash) Error() string {
	return fmt.Sprintf("wrong %s hash, expected %s got %s", e.Type, hex.EncodeToString(e.Expected), hex.EncodeToString(e.Actual))
}

type ErrNoCommonHash struct {
	Expected data.Hashes
	Actual   data.Hashes
}

func (e ErrNoCommonHash) Error() string {
	types := func(a data.Hashes) []string {
		t := make([]string, 0, len(a))
		for typ := range a {
			t = append(t, typ)
		}
		return t
	}
	return fmt.Sprintf("no common hash function, expected one of %s, got %s", types(e.Expected), types(e.Actual))
}

type ErrUnknownHashAlgorithm struct {
	Name string
}

func (e ErrUnknownHashAlgorithm) Error() string {
	return fmt.Sprintf("unknown hash algorithm: %s", e.Name)
}

type PassphraseFunc func(role string, confirm bool, change bool) ([]byte, error)

func FileMetaEqual(actual data.FileMeta, expected data.FileMeta) error {
	if actual.Length != expected.Length {
		return ErrWrongLength{expected.Length, actual.Length}
	}

	if err := hashEqual(actual.Hashes, expected.Hashes); err != nil {
		return err
	}

	return nil
}

func BytesMatchLenAndHashes(fetched []byte, length int64, hashes data.Hashes) error {
	flen := int64(len(fetched))
	if length != 0 && flen != length {
		return ErrWrongLength{length, flen}
	}

	for alg, expected := range hashes {
		var h hash.Hash
		switch alg {
		case "sha256":
			h = sha256.New()
		case "sha512":
			h = sha512.New()
		default:
			return ErrUnknownHashAlgorithm{alg}
		}
		h.Write(fetched)
		hash := h.Sum(nil)
		if !hmac.Equal(hash, expected) {
			return ErrWrongHash{alg, expected, hash}
		}
	}

	return nil
}

func hashEqual(actual data.Hashes, expected data.Hashes) error {
	hashChecked := false
	for typ, hash := range expected {
		if h, ok := actual[typ]; ok {
			hashChecked = true
			if !hmac.Equal(h, hash) {
				return ErrWrongHash{typ, hash, h}
			}
		}
	}
	if !hashChecked {
		return ErrNoCommonHash{expected, actual}
	}
	return nil
}

func VersionEqual(actual int64, expected int64) error {
	if actual != expected {
		return ErrWrongVersion{expected, actual}
	}
	return nil
}

func SnapshotFileMetaEqual(actual data.SnapshotFileMeta, expected data.SnapshotFileMeta) error {
	// TUF-1.0 no longer considers the length and hashes to be a required
	// member of snapshots. However they are considering requiring hashes
	// for delegated roles to avoid an attack described in Section 5.6 of
	// the Mercury paper:
	// https://github.com/theupdateframework/specification/pull/40
	if expected.Length != 0 && actual.Length != expected.Length {
		return ErrWrongLength{expected.Length, actual.Length}
	}
	// 5.6.2 - Check against snapshot role's targets hash
	if len(expected.Hashes) != 0 {
		if err := hashEqual(actual.Hashes, expected.Hashes); err != nil {
			return err
		}
	}
	// 5.6.4 - Check against snapshot role's snapshot version
	if err := VersionEqual(actual.Version, expected.Version); err != nil {
		return err
	}

	return nil
}

func TargetFileMetaEqual(actual data.TargetFileMeta, expected data.TargetFileMeta) error {
	return FileMetaEqual(actual.FileMeta, expected.FileMeta)
}

func TimestampFileMetaEqual(actual data.TimestampFileMeta, expected data.TimestampFileMeta) error {
	// TUF no longer considers the length and hashes to be a required
	// member of Timestamp.
	if expected.Length != 0 && actual.Length != expected.Length {
		return ErrWrongLength{expected.Length, actual.Length}
	}
	// 5.5.2 - Check against timestamp role's snapshot hash
	if len(expected.Hashes) != 0 {
		if err := hashEqual(actual.Hashes, expected.Hashes); err != nil {
			return err
		}
	}
	// 5.5.4 - Check against timestamp role's snapshot version
	if err := VersionEqual(actual.Version, expected.Version); err != nil {
		return err
	}

	return nil
}

const defaultHashAlgorithm = "sha512"

func GenerateFileMeta(r io.Reader, hashAlgorithms ...string) (data.FileMeta, error) {
	if len(hashAlgorithms) == 0 {
		hashAlgorithms = []string{defaultHashAlgorithm}
	}
	hashes := make(map[string]hash.Hash, len(hashAlgorithms))
	for _, hashAlgorithm := range hashAlgorithms {
		var h hash.Hash
		switch hashAlgorithm {
		case "sha256":
			h = sha256.New()
		case "sha512":
			h = sha512.New()
		default:
			return data.FileMeta{}, ErrUnknownHashAlgorithm{hashAlgorithm}
		}
		hashes[hashAlgorithm] = h
		r = io.TeeReader(r, h)
	}
	n, err := io.Copy(io.Discard, r)
	if err != nil {
		return data.FileMeta{}, err
	}
	m := data.FileMeta{Length: n, Hashes: make(data.Hashes, len(hashes))}
	for hashAlgorithm, h := range hashes {
		m.Hashes[hashAlgorithm] = h.Sum(nil)
	}
	return m, nil
}

type versionedMeta struct {
	Version int64 `json:"version"`
}

func generateVersionedFileMeta(r io.Reader, hashAlgorithms ...string) (data.FileMeta, int64, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return data.FileMeta{}, 0, err
	}

	m, err := GenerateFileMeta(bytes.NewReader(b), hashAlgorithms...)
	if err != nil {
		return data.FileMeta{}, 0, err
	}

	s := data.Signed{}
	if err := json.Unmarshal(b, &s); err != nil {
		return data.FileMeta{}, 0, err
	}

	vm := versionedMeta{}
	if err := json.Unmarshal(s.Signed, &vm); err != nil {
		return data.FileMeta{}, 0, err
	}

	return m, vm.Version, nil
}

func GenerateSnapshotFileMeta(r io.Reader, hashAlgorithms ...string) (data.SnapshotFileMeta, error) {
	m, v, err := generateVersionedFileMeta(r, hashAlgorithms...)
	if err != nil {
		return data.SnapshotFileMeta{}, err
	}
	return data.SnapshotFileMeta{
		Length:  m.Length,
		Hashes:  m.Hashes,
		Version: v,
	}, nil
}

func GenerateTargetFileMeta(r io.Reader, hashAlgorithms ...string) (data.TargetFileMeta, error) {
	m, err := GenerateFileMeta(r, hashAlgorithms...)
	if err != nil {
		return data.TargetFileMeta{}, err
	}
	return data.TargetFileMeta{
		FileMeta: m,
	}, nil
}

func GenerateTimestampFileMeta(r io.Reader, hashAlgorithms ...string) (data.TimestampFileMeta, error) {
	m, v, err := generateVersionedFileMeta(r, hashAlgorithms...)
	if err != nil {
		return data.TimestampFileMeta{}, err
	}
	return data.TimestampFileMeta{
		Length:  m.Length,
		Hashes:  m.Hashes,
		Version: v,
	}, nil
}

func NormalizeTarget(p string) string {
	// FIXME(TUF-0.9) TUF-1.0 is considering banning paths that begin with
	// a leading path separator, to avoid surprising behavior when joining
	// target and delgated paths. python-tuf raises an exception if any
	// path starts with '/', but since we need to be cross compatible with
	// TUF-0.9 we still need to support leading slashes. For now, we will
	// just strip them out, but eventually we should also consider turning
	// them into an error.
	return strings.TrimPrefix(path.Join("/", p), "/")
}

func VersionedPath(p string, version int64) string {
	return path.Join(path.Dir(p), strconv.FormatInt(version, 10)+"."+path.Base(p))
}

func HashedPaths(p string, hashes data.Hashes) []string {
	paths := make([]string, 0, len(hashes))
	for _, hash := range hashes {
		hashedPath := path.Join(path.Dir(p), hash.String()+"."+path.Base(p))
		paths = append(paths, hashedPath)
	}
	return paths
}

func AtomicallyWriteFile(filename string, data []byte, perm os.FileMode) error {
	dir, name := filepath.Split(filename)
	f, err := os.CreateTemp(dir, name)
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	if err != nil {
		f.Close()
		os.Remove(f.Name())
		return err
	}

	if err = f.Chmod(perm); err != nil {
		f.Close()
		os.Remove(f.Name())
		return err
	}

	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		return err
	}

	if err := os.Rename(f.Name(), filename); err != nil {
		os.Remove(f.Name())
		return err
	}

	return nil
}
