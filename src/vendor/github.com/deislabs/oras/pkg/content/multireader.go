package content

import (
	"context"
	"fmt"

	"github.com/containerd/containerd/content"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// MultiReader store to read content from multiple stores. It finds the content by asking each underlying
// store to find the content, which it does based on the hash.
//
// Example:
//        fileStore := NewFileStore(rootPath)
//        memoryStore := NewMemoryStore()
//        // load up content in fileStore and memoryStore
//        multiStore := MultiReader([]content.Provider{fileStore, memoryStore})
//
// You now can use multiStore anywhere that content.Provider is accepted
type MultiReader struct {
	stores []content.Provider
}

// AddStore add a store to read from
func (m *MultiReader) AddStore(store ...content.Provider) {
	m.stores = append(m.stores, store...)
}

// ReaderAt get a reader
func (m MultiReader) ReaderAt(ctx context.Context, desc ocispec.Descriptor) (content.ReaderAt, error) {
	for _, store := range m.stores {
		r, err := store.ReaderAt(ctx, desc)
		if r != nil && err == nil {
			return r, nil
		}
	}
	// we did not find any
	return nil, fmt.Errorf("not found")
}
