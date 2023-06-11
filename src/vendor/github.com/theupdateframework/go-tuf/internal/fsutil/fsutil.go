// Package fsutil defiens a set of internal utility functions used to
// interact with the file system.
package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// IsMetaFile tests wheter a DirEntry appears to be a metadata file or not.
func IsMetaFile(e os.DirEntry) (bool, error) {
	if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
		return false, nil
	}

	info, err := e.Info()
	if err != nil {
		return false, fmt.Errorf("error retrieving FileInfo for %s: %w", e.Name(), err)
	}

	return info.Mode().IsRegular(), nil
}
