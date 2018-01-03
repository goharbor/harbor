package storage

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"crypto/rand"
	"fmt"
	"strconv"

	"github.com/docker/notary"
	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	testDir, err := ioutil.TempDir("", "testdir")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	s, err := NewFileStore(filepath.Join(testDir, "metadata"), "json")
	require.Nil(t, err, "Initializing FilesystemStore returned unexpected error: %v", err)
	defer os.RemoveAll(testDir)

	testContent := []byte("test data")

	err = s.Set("testMeta", testContent)
	require.Nil(t, err, "Set returned unexpected error: %v", err)

	content, err := ioutil.ReadFile(filepath.Join(testDir, "metadata", "testMeta.json"))
	require.Nil(t, err, "Error reading file: %v", err)
	require.Equal(t, testContent, content, "Content written to file was corrupted.")
}

func TestSetWithNoParentDirectory(t *testing.T) {
	testDir, err := ioutil.TempDir("", "testdir")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	s, err := NewFileStore(filepath.Join(testDir, "metadata"), "json")
	require.Nil(t, err, "Initializing FilesystemStore returned unexpected error: %v", err)
	defer os.RemoveAll(testDir)

	testContent := []byte("test data")

	err = s.Set("noexist/"+"testMeta", testContent)
	require.Nil(t, err, "Set returned unexpected error: %v", err)

	content, err := ioutil.ReadFile(filepath.Join(testDir, "metadata", "noexist/testMeta.json"))
	require.Nil(t, err, "Error reading file: %v", err)
	require.Equal(t, testContent, content, "Content written to file was corrupted.")
}

// if something already existed there, remove it first and write a new file
func TestSetRemovesExistingFileBeforeWriting(t *testing.T) {
	testDir, err := ioutil.TempDir("", "testdir")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	s, err := NewFileStore(filepath.Join(testDir, "metadata"), "json")
	require.Nil(t, err, "Initializing FilesystemStore returned unexpected error: %v", err)
	defer os.RemoveAll(testDir)

	// make a directory where we want metadata to go
	os.Mkdir(filepath.Join(testDir, "metadata", "root.json"), 0700)

	testContent := []byte("test data")
	err = s.Set("root", testContent)
	require.NoError(t, err, "Set returned unexpected error: %v", err)

	content, err := ioutil.ReadFile(filepath.Join(testDir, "metadata", "root.json"))
	require.NoError(t, err, "Error reading file: %v", err)
	require.Equal(t, testContent, content, "Content written to file was corrupted.")
}

func TestGetSized(t *testing.T) {
	testDir, err := ioutil.TempDir("", "testdir")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	s, err := NewFileStore(filepath.Join(testDir, "metadata"), "json")
	require.Nil(t, err, "Initializing FilesystemStore returned unexpected error: %v", err)
	defer os.RemoveAll(testDir)

	testContent := []byte("test data")

	ioutil.WriteFile(filepath.Join(testDir, "metadata", "testMeta.json"), testContent, 0600)

	content, err := s.GetSized("testMeta", int64(len(testContent)))
	require.Nil(t, err, "GetSized returned unexpected error: %v", err)

	require.Equal(t, testContent, content, "Content read from file was corrupted.")

	// Check that NoSizeLimit size reads everything
	content, err = s.GetSized("testMeta", NoSizeLimit)
	require.Nil(t, err, "GetSized returned unexpected error: %v", err)

	require.Equal(t, testContent, content, "Content read from file was corrupted.")

	// Check that we error if the file is larger than the expected size
	content, err = s.GetSized("testMeta", 4)
	require.Error(t, err)
	require.Len(t, content, 0)
}

func TestGetSizedSet(t *testing.T) {
	testDir, err := ioutil.TempDir("", "testdir")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	s, err := NewFileStore(filepath.Join(testDir, "metadata"), "json")
	require.NoError(t, err, "Initializing FilesystemStore returned unexpected error", err)
	defer os.RemoveAll(testDir)

	testGetSetMeta(t, func() MetadataStore { return s })
}

func TestRemove(t *testing.T) {
	testDir, err := ioutil.TempDir("", "testdir")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	s, err := NewFileStore(filepath.Join(testDir, "metadata"), "json")
	require.NoError(t, err, "Initializing FilesystemStore returned unexpected error", err)
	defer os.RemoveAll(testDir)

	testRemove(t, func() MetadataStore { return s })
}

func TestRemoveAll(t *testing.T) {
	testDir, err := ioutil.TempDir("", "testdir")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	s, err := NewFileStore(filepath.Join(testDir, "metadata"), "json")
	require.Nil(t, err, "Initializing FilesystemStore returned unexpected error: %v", err)
	defer os.RemoveAll(testDir)

	testContent := []byte("test data")

	// Write some files in metadata and targets dirs
	metaPath := filepath.Join(testDir, "metadata", "testMeta.json")
	ioutil.WriteFile(metaPath, testContent, 0600)

	// Remove all
	err = s.RemoveAll()
	require.Nil(t, err, "Removing all from FilesystemStore returned unexpected error: %v", err)

	// Test that files no longer exist
	_, err = ioutil.ReadFile(metaPath)
	require.True(t, os.IsNotExist(err))

	// Removing the empty filestore returns nil
	require.Nil(t, s.RemoveAll())
}

// Tests originally from Trustmanager ensuring the FilesystemStore satisfies the
// necessary behaviour
func TestAddFile(t *testing.T) {
	testData := []byte("This test data should be part of the file.")
	testName := "docker.com/notary/certificate"
	testExt := ".crt"

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempBaseDir)

	// Since we're generating this manually we need to add the extension '.'
	expectedFilePath := filepath.Join(tempBaseDir, testName+testExt)

	// Create our FilesystemStore
	store := &FilesystemStore{
		baseDir: tempBaseDir,
		ext:     testExt,
	}

	// Call the Set function
	err = store.Set(testName, testData)
	require.NoError(t, err)

	// Check to see if file exists
	b, err := ioutil.ReadFile(expectedFilePath)
	require.NoError(t, err)
	require.Equal(t, testData, b, "unexpected content in the file: %s", expectedFilePath)
}

func TestRemoveFile(t *testing.T) {
	testName := "docker.com/notary/certificate"
	testExt := ".crt"
	perms := os.FileMode(0755)

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempBaseDir)

	// Since we're generating this manually we need to add the extension '.'
	expectedFilePath := filepath.Join(tempBaseDir, testName+testExt)

	_, err = generateRandomFile(expectedFilePath, perms)
	require.NoError(t, err)

	// Create our FilesystemStore
	store := &FilesystemStore{
		baseDir: tempBaseDir,
		ext:     testExt,
	}

	// Call the Remove function
	err = store.Remove(testName)
	require.NoError(t, err)

	// Check to see if file exists
	_, err = os.Stat(expectedFilePath)
	require.Error(t, err)
}

func TestListFiles(t *testing.T) {
	testName := "docker.com/notary/certificate"
	testExt := "crt"
	perms := os.FileMode(0755)

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempBaseDir)

	var expectedFilePath string
	// Create 10 randomfiles
	for i := 1; i <= 10; i++ {
		// Since we're generating this manually we need to add the extension '.'
		expectedFilename := testName + strconv.Itoa(i) + "." + testExt
		expectedFilePath = filepath.Join(tempBaseDir, expectedFilename)
		_, err = generateRandomFile(expectedFilePath, perms)
		require.NoError(t, err)
	}

	// Create our FilesystemStore
	store := &FilesystemStore{
		baseDir: tempBaseDir,
		ext:     testExt,
	}

	// Call the List function. Expect 10 files
	files := store.ListFiles()
	require.Len(t, files, 10)
}

func TestGetPath(t *testing.T) {
	testExt := ".crt"

	// Create our FilesystemStore
	store := &FilesystemStore{
		baseDir: "",
		ext:     testExt,
	}

	firstPath := "diogomonica.com/openvpn/0xdeadbeef.crt"
	secondPath := "/docker.io/testing-dashes/@#$%^&().crt"

	result, err := store.getPath("diogomonica.com/openvpn/0xdeadbeef")
	require.Equal(t, firstPath, result, "unexpected error from GetPath: %v", err)

	result, err = store.getPath("/docker.io/testing-dashes/@#$%^&()")
	require.Equal(t, secondPath, result, "unexpected error from GetPath: %v", err)
}

func TestGetPathProtection(t *testing.T) {
	testExt := ".crt"

	// Create our FilesystemStore
	store := &FilesystemStore{
		baseDir: "/path/to/filestore/",
		ext:     testExt,
	}

	// Should deny requests for paths outside the filestore
	_, err := store.getPath("../../etc/passwd")
	require.Error(t, err)
	require.Equal(t, ErrPathOutsideStore, err)

	_, err = store.getPath("private/../../../etc/passwd")
	require.Error(t, err)
	require.Equal(t, ErrPathOutsideStore, err)

	// Convoluted paths should work as long as they end up inside the store
	expected := "/path/to/filestore/filename.crt"
	result, err := store.getPath("private/../../filestore/./filename")
	require.NoError(t, err)
	require.Equal(t, expected, result)

	// Repeat tests with a relative baseDir
	relStore := &FilesystemStore{
		baseDir: "relative/file/path",
		ext:     testExt,
	}

	// Should deny requests for paths outside the filestore
	_, err = relStore.getPath("../../etc/passwd")
	require.Error(t, err)
	require.Equal(t, ErrPathOutsideStore, err)
	_, err = relStore.getPath("private/../../../etc/passwd")
	require.Error(t, err)
	require.Equal(t, ErrPathOutsideStore, err)

	// Convoluted paths should work as long as they end up inside the store
	expected = "relative/file/path/filename.crt"
	result, err = relStore.getPath("private/../../path/./filename")
	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func TestGetData(t *testing.T) {
	testName := "docker.com/notary/certificate"
	testExt := ".crt"
	perms := os.FileMode(0755)

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempBaseDir)

	// Since we're generating this manually we need to add the extension '.'
	expectedFilePath := filepath.Join(tempBaseDir, testName+testExt)

	expectedData, err := generateRandomFile(expectedFilePath, perms)
	require.NoError(t, err)

	// Create our FilesystemStore
	store := &FilesystemStore{
		baseDir: tempBaseDir,
		ext:     testExt,
	}
	testData, err := store.Get(testName)
	require.NoError(t, err, "failed to get data from: %s", testName)
	require.Equal(t, expectedData, testData, "unexpected content for the file: %s", expectedFilePath)
}

func TestCreateDirectory(t *testing.T) {
	testDir := "fake/path/to/directory"

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempBaseDir)

	dirPath := filepath.Join(tempBaseDir, testDir)

	// Call createDirectory
	createDirectory(dirPath, notary.PrivExecPerms)

	// Check to see if file exists
	fi, err := os.Stat(dirPath)
	require.NoError(t, err)

	// Check to see if it is a directory
	require.True(t, fi.IsDir(), "expected to be directory: %s", dirPath)

	// Check to see if the permissions match
	require.Equal(t, "drwx------", fi.Mode().String(), "permissions are wrong for: %s. Got: %s", dirPath, fi.Mode().String())
}

func TestCreatePrivateDirectory(t *testing.T) {
	testDir := "fake/path/to/private/directory"

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempBaseDir)

	dirPath := filepath.Join(tempBaseDir, testDir)

	// Call createDirectory
	createDirectory(dirPath, notary.PrivExecPerms)

	// Check to see if file exists
	fi, err := os.Stat(dirPath)
	require.NoError(t, err)

	// Check to see if it is a directory
	require.True(t, fi.IsDir(), "expected to be directory: %s", dirPath)

	// Check to see if the permissions match
	require.Equal(t, "drwx------", fi.Mode().String(), "permissions are wrong for: %s. Got: %s", dirPath, fi.Mode().String())
}

func generateRandomFile(filePath string, perms os.FileMode) ([]byte, error) {
	rndBytes := make([]byte, 10)
	_, err := rand.Read(rndBytes)
	if err != nil {
		return nil, err
	}

	os.MkdirAll(filepath.Dir(filePath), perms)
	if err = ioutil.WriteFile(filePath, rndBytes, perms); err != nil {
		return nil, err
	}

	return rndBytes, nil
}

func TestFileStoreConsistency(t *testing.T) {
	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempBaseDir)

	tempBaseDir2, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempBaseDir2)

	s, err := NewPrivateSimpleFileStore(tempBaseDir, "txt")
	require.NoError(t, err)

	s2, err := NewPrivateSimpleFileStore(tempBaseDir2, ".txt")
	require.NoError(t, err)

	file1Data := make([]byte, 20)
	_, err = rand.Read(file1Data)
	require.NoError(t, err)

	file2Data := make([]byte, 20)
	_, err = rand.Read(file2Data)
	require.NoError(t, err)

	file3Data := make([]byte, 20)
	_, err = rand.Read(file3Data)
	require.NoError(t, err)

	file1Path := "file1"
	file2Path := "path/file2"
	file3Path := "long/path/file3"

	for _, s := range []*FilesystemStore{s, s2} {
		s.Set(file1Path, file1Data)
		s.Set(file2Path, file2Data)
		s.Set(file3Path, file3Data)

		paths := map[string][]byte{
			file1Path: file1Data,
			file2Path: file2Data,
			file3Path: file3Data,
		}
		for _, p := range s.ListFiles() {
			_, ok := paths[p]
			require.True(t, ok, fmt.Sprintf("returned path not found: %s", p))
			d, err := s.Get(p)
			require.NoError(t, err)
			require.Equal(t, paths[p], d)
		}
	}

}
