package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	ctxu "github.com/docker/distribution/context"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"

	"github.com/docker/notary"
	"github.com/docker/notary/server/errors"
	"github.com/docker/notary/server/snapshot"
	"github.com/docker/notary/server/storage"
	"github.com/docker/notary/server/timestamp"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/docker/notary/tuf/validation"
	"github.com/docker/notary/utils"
)

// MainHandler is the default handler for the server
func MainHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// For now it only supports `GET`
	if r.Method != "GET" {
		return errors.ErrGenericNotFound.WithDetail(nil)
	}

	if _, err := w.Write([]byte("{}")); err != nil {
		return errors.ErrUnknown.WithDetail(err)
	}
	return nil
}

// AtomicUpdateHandler will accept multiple TUF files and ensure that the storage
// backend is atomically updated with all the new records.
func AtomicUpdateHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()
	vars := mux.Vars(r)
	return atomicUpdateHandler(ctx, w, r, vars)
}

func atomicUpdateHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	gun := data.GUN(vars["gun"])
	s := ctx.Value(notary.CtxKeyMetaStore)
	logger := ctxu.GetLoggerWithField(ctx, gun, "gun")
	store, ok := s.(storage.MetaStore)
	if !ok {
		logger.Error("500 POST unable to retrieve storage")
		return errors.ErrNoStorage.WithDetail(nil)
	}
	cryptoServiceVal := ctx.Value(notary.CtxKeyCryptoSvc)
	cryptoService, ok := cryptoServiceVal.(signed.CryptoService)
	if !ok {
		logger.Error("500 POST unable to retrieve signing service")
		return errors.ErrNoCryptoService.WithDetail(nil)
	}

	reader, err := r.MultipartReader()
	if err != nil {
		logger.Info("400 POST unable to parse TUF data")
		return errors.ErrMalformedUpload.WithDetail(nil)
	}
	var updates []storage.MetaUpdate
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		role := data.RoleName(strings.TrimSuffix(part.FileName(), ".json"))
		if role.String() == "" {
			logger.Info("400 POST empty role")
			return errors.ErrNoFilename.WithDetail(nil)
		} else if !data.ValidRole(role) {
			logger.Infof("400 POST invalid role: %s", role)
			return errors.ErrInvalidRole.WithDetail(role)
		}
		meta := &data.SignedMeta{}
		var input []byte
		inBuf := bytes.NewBuffer(input)
		dec := json.NewDecoder(io.TeeReader(part, inBuf))
		err = dec.Decode(meta)
		if err != nil {
			logger.Info("400 POST malformed update JSON")
			return errors.ErrMalformedJSON.WithDetail(nil)
		}
		version := meta.Signed.Version
		updates = append(updates, storage.MetaUpdate{
			Role:    role,
			Version: version,
			Data:    inBuf.Bytes(),
		})
	}
	updates, err = validateUpdate(cryptoService, gun, updates, store)
	if err != nil {
		serializable, serializableError := validation.NewSerializableError(err)
		if serializableError != nil {
			logger.Info("400 POST error validating update")
			return errors.ErrInvalidUpdate.WithDetail(nil)
		}
		return errors.ErrInvalidUpdate.WithDetail(serializable)
	}
	err = store.UpdateMany(gun, updates)
	if err != nil {
		// If we have an old version error, surface to user with error code
		if _, ok := err.(storage.ErrOldVersion); ok {
			logger.Info("400 POST old version error")
			return errors.ErrOldVersion.WithDetail(err)
		}
		// More generic storage update error, possibly due to attempted rollback
		logger.Errorf("500 POST error applying update request: %v", err)
		return errors.ErrUpdating.WithDetail(nil)
	}
	return nil
}

// GetHandler returns the json for a specified role and GUN.
func GetHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()
	vars := mux.Vars(r)
	return getHandler(ctx, w, r, vars)
}

func getHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	gun := data.GUN(vars["gun"])
	checksum := vars["checksum"]
	version := vars["version"]
	tufRole := vars["tufRole"]
	s := ctx.Value(notary.CtxKeyMetaStore)

	logger := ctxu.GetLoggerWithField(ctx, gun, "gun")

	store, ok := s.(storage.MetaStore)
	if !ok {
		logger.Error("500 GET: no storage exists")
		return errors.ErrNoStorage.WithDetail(nil)
	}

	lastModified, output, err := getRole(ctx, store, gun, data.RoleName(tufRole), checksum, version)
	if err != nil {
		logger.Infof("404 GET %s role", tufRole)
		return err
	}
	if lastModified != nil {
		// This shouldn't always be true, but in case it is nil, and the last modified headers
		// are not set, the cache control handler should set the last modified date to the beginning
		// of time.
		utils.SetLastModifiedHeader(w.Header(), *lastModified)
	} else {
		logger.Warnf("Got bytes out for %s's %s (checksum: %s), but missing lastModified date",
			gun, tufRole, checksum)
	}

	w.Write(output)
	return nil
}

// DeleteHandler deletes all data for a GUN. A 200 responses indicates success.
func DeleteHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	gun := data.GUN(vars["gun"])
	logger := ctxu.GetLoggerWithField(ctx, gun, "gun")
	s := ctx.Value(notary.CtxKeyMetaStore)
	store, ok := s.(storage.MetaStore)
	if !ok {
		logger.Error("500 DELETE repository: no storage exists")
		return errors.ErrNoStorage.WithDetail(nil)
	}
	err := store.Delete(gun)
	if err != nil {
		logger.Error("500 DELETE repository")
		return errors.ErrUnknown.WithDetail(err)
	}
	return nil
}

// GetKeyHandler returns a public key for the specified role, creating a new key-pair
// it if it doesn't yet exist
func GetKeyHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()
	vars := mux.Vars(r)
	return getKeyHandler(ctx, w, r, vars)
}

func getKeyHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	role, gun, keyAlgorithm, store, crypto, err := setupKeyHandler(ctx, w, r, vars, http.MethodGet)
	if err != nil {
		return err
	}
	var key data.PublicKey
	logger := ctxu.GetLoggerWithField(ctx, gun, "gun")
	switch role {
	case data.CanonicalTimestampRole:
		key, err = timestamp.GetOrCreateTimestampKey(gun, store, crypto, keyAlgorithm)
	case data.CanonicalSnapshotRole:
		key, err = snapshot.GetOrCreateSnapshotKey(gun, store, crypto, keyAlgorithm)
	default:
		logger.Infof("400 GET %s key: %v", role, err)
		return errors.ErrInvalidRole.WithDetail(role)
	}
	if err != nil {
		logger.Errorf("500 GET %s key: %v", role, err)
		return errors.ErrUnknown.WithDetail(err)
	}

	out, err := json.Marshal(key)
	if err != nil {
		logger.Errorf("500 GET %s key", role)
		return errors.ErrUnknown.WithDetail(err)
	}
	logger.Debugf("200 GET %s key", role)
	w.Write(out)
	return nil
}

// RotateKeyHandler rotates the remote key for the specified role, returning the public key
func RotateKeyHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()
	vars := mux.Vars(r)
	return rotateKeyHandler(ctx, w, r, vars)
}

func rotateKeyHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	role, gun, keyAlgorithm, store, crypto, err := setupKeyHandler(ctx, w, r, vars, http.MethodPost)
	if err != nil {
		return err
	}
	var key data.PublicKey
	logger := ctxu.GetLoggerWithField(ctx, gun, "gun")
	switch role {
	case data.CanonicalTimestampRole:
		key, err = timestamp.RotateTimestampKey(gun, store, crypto, keyAlgorithm)
	case data.CanonicalSnapshotRole:
		key, err = snapshot.RotateSnapshotKey(gun, store, crypto, keyAlgorithm)
	default:
		logger.Infof("400 POST %s key: %v", role, err)
		return errors.ErrInvalidRole.WithDetail(role)
	}
	if err != nil {
		logger.Errorf("500 POST %s key: %v", role, err)
		return errors.ErrUnknown.WithDetail(err)
	}

	out, err := json.Marshal(key)
	if err != nil {
		logger.Errorf("500 POST %s key", role)
		return errors.ErrUnknown.WithDetail(err)
	}
	logger.Debugf("200 POST %s key", role)
	w.Write(out)
	return nil
}

// To be called before getKeyHandler or rotateKeyHandler
func setupKeyHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string, actionVerb string) (data.RoleName, data.GUN, string, storage.MetaStore, signed.CryptoService, error) {
	gun := data.GUN(vars["gun"])
	logger := ctxu.GetLoggerWithField(ctx, gun, "gun")
	if gun == "" {
		logger.Infof("400 %s no gun in request", actionVerb)
		return "", "", "", nil, nil, errors.ErrUnknown.WithDetail("no gun")
	}

	role := data.RoleName(vars["tufRole"])
	if role == "" {
		logger.Infof("400 %s no role in request", actionVerb)
		return "", "", "", nil, nil, errors.ErrUnknown.WithDetail("no role")
	}

	s := ctx.Value(notary.CtxKeyMetaStore)
	store, ok := s.(storage.MetaStore)
	if !ok || store == nil {
		logger.Errorf("500 %s storage not configured", actionVerb)
		return "", "", "", nil, nil, errors.ErrNoStorage.WithDetail(nil)
	}
	c := ctx.Value(notary.CtxKeyCryptoSvc)
	crypto, ok := c.(signed.CryptoService)
	if !ok || crypto == nil {
		logger.Errorf("500 %s crypto service not configured", actionVerb)
		return "", "", "", nil, nil, errors.ErrNoCryptoService.WithDetail(nil)
	}
	algo := ctx.Value(notary.CtxKeyKeyAlgo)
	keyAlgo, ok := algo.(string)
	if !ok || keyAlgo == "" {
		logger.Errorf("500 %s key algorithm not configured", actionVerb)
		return "", "", "", nil, nil, errors.ErrNoKeyAlgorithm.WithDetail(nil)
	}

	return role, gun, keyAlgo, store, crypto, nil
}

// NotFoundHandler is used as a generic catch all handler to return the ErrMetadataNotFound
// 404 response
func NotFoundHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return errors.ErrMetadataNotFound.WithDetail(nil)
}
