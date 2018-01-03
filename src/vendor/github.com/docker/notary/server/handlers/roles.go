package handlers

import (
	"strconv"
	"time"

	"golang.org/x/net/context"

	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/docker/notary"
	"github.com/docker/notary/server/errors"
	"github.com/docker/notary/server/storage"
	"github.com/docker/notary/server/timestamp"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
)

func getRole(ctx context.Context, store storage.MetaStore, gun data.GUN, role data.RoleName, checksum, version string) (*time.Time, []byte, error) {
	var (
		lastModified *time.Time
		out          []byte
		err          error
	)
	if checksum != "" {
		lastModified, out, err = store.GetChecksum(gun, role, checksum)
	} else if version != "" {
		v, vErr := strconv.Atoi(version)
		if vErr != nil {
			return nil, nil, errors.ErrMetadataNotFound.WithDetail(vErr)
		}
		lastModified, out, err = store.GetVersion(gun, role, v)
	} else {
		// the timestamp and snapshot might be server signed so are
		// handled specially
		switch role {
		case data.CanonicalTimestampRole, data.CanonicalSnapshotRole:
			return getMaybeServerSigned(ctx, store, gun, role)
		}
		lastModified, out, err = store.GetCurrent(gun, role)

	}

	if err != nil {
		if _, ok := err.(storage.ErrNotFound); ok {
			return nil, nil, errors.ErrMetadataNotFound.WithDetail(err)
		}
		return nil, nil, errors.ErrUnknown.WithDetail(err)
	}
	if out == nil {
		return nil, nil, errors.ErrMetadataNotFound.WithDetail(nil)
	}

	return lastModified, out, nil
}

// getMaybeServerSigned writes the current snapshot or timestamp (based on the
// role passed) to the provided writer or returns an error. In retrieving
// the timestamp and snapshot, based on the keys held by the server, a new one
// might be generated and signed due to expiry of the previous one or updates
// to other roles.
func getMaybeServerSigned(ctx context.Context, store storage.MetaStore, gun data.GUN, role data.RoleName) (*time.Time, []byte, error) {
	cryptoServiceVal := ctx.Value(notary.CtxKeyCryptoSvc)
	cryptoService, ok := cryptoServiceVal.(signed.CryptoService)
	if !ok {
		return nil, nil, errors.ErrNoCryptoService.WithDetail(nil)
	}

	var (
		lastModified *time.Time
		out          []byte
		err          error
	)
	if role != data.CanonicalTimestampRole && role != data.CanonicalSnapshotRole {
		return nil, nil, fmt.Errorf("role %s cannot be server signed", role.String())
	}
	lastModified, out, err = timestamp.GetOrCreateTimestamp(gun, store, cryptoService)
	if err != nil {
		switch err.(type) {
		case *storage.ErrNoKey, storage.ErrNotFound:
			return nil, nil, errors.ErrMetadataNotFound.WithDetail(err)
		default:
			return nil, nil, errors.ErrUnknown.WithDetail(err)
		}
	}

	// If we wanted the snapshot, get it by checksum from the timestamp data
	if role == data.CanonicalSnapshotRole {
		ts := new(data.SignedTimestamp)
		if err := json.Unmarshal(out, ts); err != nil {
			return nil, nil, err
		}
		snapshotChecksums, err := ts.GetSnapshot()
		if err != nil || snapshotChecksums == nil {
			return nil, nil, fmt.Errorf("could not retrieve latest snapshot checksum")
		}
		if snapshotSHA256Bytes, ok := snapshotChecksums.Hashes[notary.SHA256]; ok {
			snapshotSHA256Hex := hex.EncodeToString(snapshotSHA256Bytes[:])
			return store.GetChecksum(gun, role, snapshotSHA256Hex)
		}
		return nil, nil, fmt.Errorf("could not retrieve sha256 snapshot checksum")
	}

	return lastModified, out, nil
}
