// +build !pkcs11

package main

import (
	"errors"

	"github.com/docker/notary"
	store "github.com/docker/notary/storage"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/utils"
)

func getYubiStore(fileKeyStore trustmanager.KeyStore, ret notary.PassRetriever) (trustmanager.KeyStore, error) {
	return nil, errors.New("Not built with hardware support")
}

func getImporters(baseDir string, _ notary.PassRetriever) ([]utils.Importer, error) {
	fileStore, err := store.NewPrivateKeyFileStorage(baseDir, notary.KeyExtension)
	if err != nil {
		return nil, err
	}
	return []utils.Importer{fileStore}, nil
}
