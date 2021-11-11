package registryctl

import (
	"github.com/stretchr/testify/mock"
)

type Mockclient struct {
	mock.Mock
}

// Health ...
func (c *Mockclient) Health() error {
	return nil
}

// DeleteBlob ...
func (c *Mockclient) DeleteBlob(reference string) (err error) {
	return nil
}

// DeleteManifest ...
func (c *Mockclient) DeleteManifest(repository, reference string) (err error) {
	return nil
}

// Purge ...
func (c *Mockclient) Purge(olderThan int64, dryRun, logOut, async bool) (err error) {
	return nil
}
