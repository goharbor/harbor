package trustmanager

import (
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/theupdateframework/notary"
	tufdata "github.com/theupdateframework/notary/tuf/data"
	"github.com/theupdateframework/notary/tuf/utils"
)

// Exporter is a simple interface for the two functions we need from the Storage interface
type Exporter interface {
	Get(string) ([]byte, error)
	ListFiles() []string
}

// Importer is a simple interface for the one function we need from the Storage interface
type Importer interface {
	Set(string, []byte) error
}

// ExportKeysByGUN exports all keys filtered to a GUN
func ExportKeysByGUN(to io.Writer, s Exporter, gun string) error {
	keys := s.ListFiles()
	sort.Strings(keys) // ensure consistency. ListFiles has no order guarantee
	for _, loc := range keys {
		keyFile, err := s.Get(loc)
		if err != nil {
			logrus.Warn("Could not parse key file at ", loc)
			continue
		}
		block, _ := pem.Decode(keyFile)
		keyGun := block.Headers["gun"]
		if keyGun == gun { // must be full GUN match
			if err := ExportKeys(to, s, loc); err != nil {
				return err
			}
		}
	}
	return nil
}

// ExportKeysByID exports all keys matching the given ID
func ExportKeysByID(to io.Writer, s Exporter, ids []string) error {
	want := make(map[string]struct{})
	for _, id := range ids {
		want[id] = struct{}{}
	}
	keys := s.ListFiles()
	for _, k := range keys {
		id := filepath.Base(k)
		if _, ok := want[id]; ok {
			if err := ExportKeys(to, s, k); err != nil {
				return err
			}
		}
	}
	return nil
}

// ExportKeys copies a key from the store to the io.Writer
func ExportKeys(to io.Writer, s Exporter, from string) error {
	// get PEM block
	k, err := s.Get(from)
	if err != nil {
		return err
	}

	// parse PEM blocks if there are more than one
	for block, rest := pem.Decode(k); block != nil; block, rest = pem.Decode(rest) {
		// add from path in a header for later import
		block.Headers["path"] = from
		// write serialized PEM
		err = pem.Encode(to, block)
		if err != nil {
			return err
		}
	}
	return nil
}

// ImportKeys expects an io.Reader containing one or more PEM blocks.
// It reads PEM blocks one at a time until pem.Decode returns a nil
// block.
// Each block is written to the subpath indicated in the "path" PEM
// header. If the file already exists, the file is truncated. Multiple
// adjacent PEMs with the same "path" header are appended together.
func ImportKeys(from io.Reader, to []Importer, fallbackRole string, fallbackGUN string, passRet notary.PassRetriever) error {
	// importLogic.md contains a small flowchart I made to clear up my understand while writing the cases in this function
	// it is very rough, but it may help while reading this piece of code
	data, err := ioutil.ReadAll(from)
	if err != nil {
		return err
	}
	var (
		writeTo   string
		toWrite   []byte
		errBlocks []string
	)
	for block, rest := pem.Decode(data); block != nil; block, rest = pem.Decode(rest) {
		handleLegacyPath(block)
		setFallbacks(block, fallbackGUN, fallbackRole)

		loc, err := checkValidity(block)
		if err != nil {
			// already logged in checkValidity
			errBlocks = append(errBlocks, err.Error())
			continue
		}

		// the path header is not of any use once we've imported the key so strip it away
		delete(block.Headers, "path")

		// we are now all set for import but let's first encrypt the key
		blockBytes := pem.EncodeToMemory(block)
		// check if key is encrypted, note: if it is encrypted at this point, it will have had a path header
		if privKey, err := utils.ParsePEMPrivateKey(blockBytes, ""); err == nil {
			// Key is not encrypted- ask for a passphrase and encrypt this key
			var chosenPassphrase string
			for attempts := 0; ; attempts++ {
				var giveup bool
				chosenPassphrase, giveup, err = passRet(loc, block.Headers["role"], true, attempts)
				if err == nil {
					break
				}
				if giveup || attempts > 10 {
					return errors.New("maximum number of passphrase attempts exceeded")
				}
			}
			blockBytes, err = utils.ConvertPrivateKeyToPKCS8(privKey, tufdata.RoleName(block.Headers["role"]), tufdata.GUN(block.Headers["gun"]), chosenPassphrase)
			if err != nil {
				return errors.New("failed to encrypt key with given passphrase")
			}
		}

		if loc != writeTo {
			// next location is different from previous one. We've finished aggregating
			// data for the previous file. If we have data, write the previous file,
			// clear toWrite and set writeTo to the next path we're going to write
			if toWrite != nil {
				if err = importToStores(to, writeTo, toWrite); err != nil {
					return err
				}
			}
			// set up for aggregating next file's data
			toWrite = nil
			writeTo = loc
		}

		toWrite = append(toWrite, blockBytes...)
	}
	if toWrite != nil { // close out final iteration if there's data left
		return importToStores(to, writeTo, toWrite)
	}
	if len(errBlocks) > 0 {
		return fmt.Errorf("failed to import all keys: %s", strings.Join(errBlocks, ", "))
	}
	return nil
}

func handleLegacyPath(block *pem.Block) {
	// if there is a legacy path then we set the gun header from this path
	// this is the case when a user attempts to import a key bundle generated by an older client
	if rawPath := block.Headers["path"]; rawPath != "" && rawPath != filepath.Base(rawPath) {
		// this is a legacy filepath and we should try to deduce the gun name from it
		pathWOFileName := filepath.Dir(rawPath)
		if strings.HasPrefix(pathWOFileName, notary.NonRootKeysSubdir) {
			// remove the notary keystore-specific segment of the path, and any potential leading or trailing slashes
			gunName := strings.Trim(strings.TrimPrefix(pathWOFileName, notary.NonRootKeysSubdir), "/")
			if gunName != "" {
				block.Headers["gun"] = gunName
			}
		}
		block.Headers["path"] = filepath.Base(rawPath)
	}
}

func setFallbacks(block *pem.Block, fallbackGUN, fallbackRole string) {
	if block.Headers["gun"] == "" {
		if fallbackGUN != "" {
			block.Headers["gun"] = fallbackGUN
		}
	}

	if block.Headers["role"] == "" {
		if fallbackRole == "" {
			block.Headers["role"] = notary.DefaultImportRole
		} else {
			block.Headers["role"] = fallbackRole
		}
	}
}

// checkValidity ensures the fields in the pem headers are valid and parses out the location.
// While importing a collection of keys, errors from this function should result in only the
// current pem block being skipped.
func checkValidity(block *pem.Block) (string, error) {
	// A root key or a delegations key should not have a gun
	// Note that a key that is not any of the canonical roles (except root) is a delegations key and should not have a gun
	switch block.Headers["role"] {
	case tufdata.CanonicalSnapshotRole.String(), tufdata.CanonicalTargetsRole.String(), tufdata.CanonicalTimestampRole.String():
		// check if the key is missing a gun header or has an empty gun and error out since we don't know what gun it belongs to
		if block.Headers["gun"] == "" {
			logrus.Warnf("failed to import key (%s) to store: Cannot have canonical role key without a gun, don't know what gun it belongs to", block.Headers["path"])
			return "", errors.New("invalid key pem block")
		}
	default:
		delete(block.Headers, "gun")
	}

	loc, ok := block.Headers["path"]
	// only if the path isn't specified do we get into this parsing path logic
	if !ok || loc == "" {
		// if the path isn't specified, we will try to infer the path rel to trust dir from the role (and then gun)
		// parse key for the keyID which we will save it by.
		// if the key is encrypted at this point, we will generate an error and continue since we don't know the ID to save it by

		decodedKey, err := utils.ParsePEMPrivateKey(pem.EncodeToMemory(block), "")
		if err != nil {
			logrus.Warn("failed to import key to store: Invalid key generated, key may be encrypted and does not contain path header")
			return "", errors.New("invalid key pem block")
		}
		loc = decodedKey.ID()
	}
	return loc, nil
}

func importToStores(to []Importer, path string, bytes []byte) error {
	var err error
	for _, i := range to {
		if err = i.Set(path, bytes); err != nil {
			logrus.Errorf("failed to import key to store: %s", err.Error())
			continue
		}
		break
	}
	return err
}
