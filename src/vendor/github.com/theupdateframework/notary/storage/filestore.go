package storage

import (
	"bytes"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/theupdateframework/notary"
)

// NewFileStore creates a fully configurable file store
func NewFileStore(baseDir, fileExt string) (*FilesystemStore, error) {
	baseDir = filepath.Clean(baseDir)
	if err := createDirectory(baseDir, notary.PrivExecPerms); err != nil {
		return nil, err
	}
	if !strings.HasPrefix(fileExt, ".") {
		fileExt = "." + fileExt
	}

	return &FilesystemStore{
		baseDir: baseDir,
		ext:     fileExt,
	}, nil
}

// NewPrivateKeyFileStorage initializes a new filestore for private keys, appending
// the notary.PrivDir to the baseDir.
func NewPrivateKeyFileStorage(baseDir, fileExt string) (*FilesystemStore, error) {
	baseDir = filepath.Join(baseDir, notary.PrivDir)
	myStore, err := NewFileStore(baseDir, fileExt)
	myStore.migrateTo0Dot4()
	return myStore, err
}

// NewPrivateSimpleFileStore is a wrapper to create an owner readable/writeable
// _only_ filestore
func NewPrivateSimpleFileStore(baseDir, fileExt string) (*FilesystemStore, error) {
	return NewFileStore(baseDir, fileExt)
}

// FilesystemStore is a store in a locally accessible directory
type FilesystemStore struct {
	baseDir string
	ext     string
}

func (f *FilesystemStore) moveKeyTo0Dot4Location(file string) {
	keyID := filepath.Base(file)
	fileDir := filepath.Dir(file)
	d, _ := f.Get(file)
	block, _ := pem.Decode(d)
	if block == nil {
		logrus.Warn("Key data for", file, "could not be decoded as a valid PEM block. The key will not been migrated and may not be available")
		return
	}
	fileDir = strings.TrimPrefix(fileDir, notary.RootKeysSubdir)
	fileDir = strings.TrimPrefix(fileDir, notary.NonRootKeysSubdir)
	if fileDir != "" {
		block.Headers["gun"] = filepath.ToSlash(fileDir[1:])
	}
	if strings.Contains(keyID, "_") {
		role := strings.Split(keyID, "_")[1]
		keyID = strings.TrimSuffix(keyID, "_"+role)
		block.Headers["role"] = role
	}
	var keyPEM bytes.Buffer
	// since block came from decoding the PEM bytes in the first place, and all we're doing is adding some headers we ignore the possibility of an error while encoding the block
	pem.Encode(&keyPEM, block)
	f.Set(keyID, keyPEM.Bytes())
}

func (f *FilesystemStore) migrateTo0Dot4() {
	rootKeysSubDir := filepath.Clean(filepath.Join(f.Location(), notary.RootKeysSubdir))
	nonRootKeysSubDir := filepath.Clean(filepath.Join(f.Location(), notary.NonRootKeysSubdir))
	if _, err := os.Stat(rootKeysSubDir); !os.IsNotExist(err) && f.Location() != rootKeysSubDir {
		if rootKeysSubDir == "" || rootKeysSubDir == "/" {
			// making sure we don't remove a user's homedir
			logrus.Warn("The directory for root keys is an unsafe value, we are not going to delete the directory. Please delete it manually")
		} else {
			// root_keys exists, migrate things from it
			listOnlyRootKeysDirStore, _ := NewFileStore(rootKeysSubDir, f.ext)
			for _, file := range listOnlyRootKeysDirStore.ListFiles() {
				f.moveKeyTo0Dot4Location(filepath.Join(notary.RootKeysSubdir, file))
			}
			// delete the old directory
			os.RemoveAll(rootKeysSubDir)
		}
	}

	if _, err := os.Stat(nonRootKeysSubDir); !os.IsNotExist(err) && f.Location() != nonRootKeysSubDir {
		if nonRootKeysSubDir == "" || nonRootKeysSubDir == "/" {
			// making sure we don't remove a user's homedir
			logrus.Warn("The directory for non root keys is an unsafe value, we are not going to delete the directory. Please delete it manually")
		} else {
			// tuf_keys exists, migrate things from it
			listOnlyNonRootKeysDirStore, _ := NewFileStore(nonRootKeysSubDir, f.ext)
			for _, file := range listOnlyNonRootKeysDirStore.ListFiles() {
				f.moveKeyTo0Dot4Location(filepath.Join(notary.NonRootKeysSubdir, file))
			}
			// delete the old directory
			os.RemoveAll(nonRootKeysSubDir)
		}
	}

	// if we have a trusted_certificates folder, let's delete for a complete migration since it is unused by new clients
	certsSubDir := filepath.Join(f.Location(), "trusted_certificates")
	if certsSubDir == "" || certsSubDir == "/" {
		logrus.Warn("The directory for trusted certificate is an unsafe value, we are not going to delete the directory. Please delete it manually")
	} else {
		os.RemoveAll(certsSubDir)
	}
}

func (f *FilesystemStore) getPath(name string) (string, error) {
	fileName := fmt.Sprintf("%s%s", name, f.ext)
	fullPath := filepath.Join(f.baseDir, fileName)

	if !strings.HasPrefix(fullPath, f.baseDir) {
		return "", ErrPathOutsideStore
	}
	return fullPath, nil
}

// GetSized returns the meta for the given name (a role) up to size bytes
// If size is "NoSizeLimit", this corresponds to "infinite," but we cut off at a
// predefined threshold "notary.MaxDownloadSize". If the file is larger than size
// we return ErrMaliciousServer for consistency with the HTTPStore
func (f *FilesystemStore) GetSized(name string, size int64) ([]byte, error) {
	p, err := f.getPath(name)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(p, os.O_RDONLY, notary.PrivNoExecPerms)
	if err != nil {
		if os.IsNotExist(err) {
			err = ErrMetaNotFound{Resource: name}
		}
		return nil, err
	}
	defer file.Close()

	if size == NoSizeLimit {
		size = notary.MaxDownloadSize
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if stat.Size() > size {
		return nil, ErrMaliciousServer{}
	}

	l := io.LimitReader(file, size)
	return ioutil.ReadAll(l)
}

// Get returns the meta for the given name.
func (f *FilesystemStore) Get(name string) ([]byte, error) {
	p, err := f.getPath(name)
	if err != nil {
		return nil, err
	}
	meta, err := ioutil.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			err = ErrMetaNotFound{Resource: name}
		}
		return nil, err
	}
	return meta, nil
}

// SetMulti sets the metadata for multiple roles in one operation
func (f *FilesystemStore) SetMulti(metas map[string][]byte) error {
	for role, blob := range metas {
		err := f.Set(role, blob)
		if err != nil {
			return err
		}
	}
	return nil
}

// Set sets the meta for a single role
func (f *FilesystemStore) Set(name string, meta []byte) error {
	fp, err := f.getPath(name)
	if err != nil {
		return err
	}

	// Ensures the parent directories of the file we are about to write exist
	err = os.MkdirAll(filepath.Dir(fp), notary.PrivExecPerms)
	if err != nil {
		return err
	}

	// if something already exists, just delete it and re-write it
	os.RemoveAll(fp)

	// Write the file to disk
	return ioutil.WriteFile(fp, meta, notary.PrivNoExecPerms)
}

// RemoveAll clears the existing filestore by removing its base directory
func (f *FilesystemStore) RemoveAll() error {
	return os.RemoveAll(f.baseDir)
}

// Remove removes the metadata for a single role - if the metadata doesn't
// exist, no error is returned
func (f *FilesystemStore) Remove(name string) error {
	p, err := f.getPath(name)
	if err != nil {
		return err
	}
	return os.RemoveAll(p) // RemoveAll succeeds if path doesn't exist
}

// Location returns a human readable name for the storage location
func (f FilesystemStore) Location() string {
	return f.baseDir
}

// ListFiles returns a list of all the filenames that can be used with Get*
// to retrieve content from this filestore
func (f FilesystemStore) ListFiles() []string {
	files := make([]string, 0, 0)
	filepath.Walk(f.baseDir, func(fp string, fi os.FileInfo, err error) error {
		// If there are errors, ignore this particular file
		if err != nil {
			return nil
		}
		// Ignore if it is a directory
		if fi.IsDir() {
			return nil
		}

		// If this is a symlink, ignore it
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil
		}

		// Only allow matches that end with our certificate extension (e.g. *.crt)
		matched, _ := filepath.Match("*"+f.ext, fi.Name())

		if matched {
			// Find the relative path for this file relative to the base path.
			fp, err = filepath.Rel(f.baseDir, fp)
			if err != nil {
				return err
			}
			trimmed := strings.TrimSuffix(fp, f.ext)
			files = append(files, trimmed)
		}
		return nil
	})
	return files
}

// createDirectory receives a string of the path to a directory.
// It does not support passing files, so the caller has to remove
// the filename by doing filepath.Dir(full_path_to_file)
func createDirectory(dir string, perms os.FileMode) error {
	// This prevents someone passing /path/to/dir and 'dir' not being created
	// If two '//' exist, MkdirAll deals it with correctly
	dir = dir + "/"
	return os.MkdirAll(dir, perms)
}
