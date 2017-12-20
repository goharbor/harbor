// +build pkcs11

package main

import (
	"github.com/docker/notary"
	store "github.com/docker/notary/storage"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/trustmanager/yubikey"
	"github.com/docker/notary/utils"
)

func getYubiStore(fileKeyStore trustmanager.KeyStore, ret notary.PassRetriever) (*yubikey.YubiStore, error) {
	return yubikey.NewYubiStore(fileKeyStore, ret)
}

func getImporters(baseDir string, ret notary.PassRetriever) ([]utils.Importer, error) {

	var importers []utils.Importer
	if yubikey.IsAccessible() {
		yubiStore, err := getYubiStore(nil, ret)
		if err == nil {
			importers = append(
				importers,
				yubikey.NewImporter(yubiStore, ret),
			)
		}
	}
	fileStore, err := store.NewPrivateKeyFileStorage(baseDir, notary.KeyExtension)
	if err == nil {
		importers = append(
			importers,
			fileStore,
		)
	} else if len(importers) == 0 {
		return nil, err // couldn't initialize any stores
	}
	return importers, nil
}
