package driver

import (
	"context"
	"fmt"
	"testing"
)

type changingFileSystem struct {
	StorageDriver
	fileset   []string
	keptFiles map[string]bool
}

func (cfs *changingFileSystem) List(ctx context.Context, path string) ([]string, error) {
	return cfs.fileset, nil
}
func (cfs *changingFileSystem) Stat(ctx context.Context, path string) (FileInfo, error) {
	kept, ok := cfs.keptFiles[path]
	if ok && kept {
		return &FileInfoInternal{
			FileInfoFields: FileInfoFields{
				Path: path,
			},
		}, nil
	}
	return nil, PathNotFoundError{}
}
func TestWalkFileRemoved(t *testing.T) {
	d := &changingFileSystem{
		fileset: []string{"zoidberg", "bender"},
		keptFiles: map[string]bool{
			"zoidberg": true,
		},
	}
	infos := []FileInfo{}
	err := WalkFallback(context.Background(), d, "", func(fileInfo FileInfo) error {
		infos = append(infos, fileInfo)
		return nil
	})
	if len(infos) != 1 || infos[0].Path() != "zoidberg" {
		t.Errorf(fmt.Sprintf("unexpected path set during walk: %s", infos))
	}
	if err != nil {
		t.Fatalf(err.Error())
	}
}
