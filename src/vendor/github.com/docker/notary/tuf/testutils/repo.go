package testutils

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/docker/go/canonical/json"
	"github.com/docker/notary/cryptoservice"
	"github.com/docker/notary/passphrase"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/utils"
	"github.com/stretchr/testify/require"

	"github.com/docker/notary/tuf"
	"github.com/docker/notary/tuf/signed"
)

// CreateKey creates a new key inside the cryptoservice for the given role and gun,
// returning the public key.  If the role is a root role, create an x509 key.
func CreateKey(cs signed.CryptoService, gun data.GUN, role data.RoleName, keyAlgorithm string) (data.PublicKey, error) {
	key, err := cs.Create(role, gun, keyAlgorithm)
	if err != nil {
		return nil, err
	}
	if role == data.CanonicalRootRole {
		start := time.Now().AddDate(0, 0, -1)
		privKey, _, err := cs.GetPrivateKey(key.ID())
		if err != nil {
			return nil, err
		}
		cert, err := cryptoservice.GenerateCertificate(
			privKey, gun, start, start.AddDate(1, 0, 0),
		)
		if err != nil {
			return nil, err
		}
		// Keep the x509 key type consistent with the key's algorithm
		switch keyAlgorithm {
		case data.RSAKey:
			key = data.NewRSAx509PublicKey(utils.CertToPEM(cert))
		case data.ECDSAKey:
			key = data.NewECDSAx509PublicKey(utils.CertToPEM(cert))
		default:
			// This should be impossible because of the Create() call above, but just in case
			return nil, fmt.Errorf("invalid key algorithm type")
		}

	}
	return key, nil
}

// CopyKeys copies keys of a particular role to a new cryptoservice, and returns that cryptoservice
func CopyKeys(t *testing.T, from signed.CryptoService, roles ...data.RoleName) signed.CryptoService {
	memKeyStore := trustmanager.NewKeyMemoryStore(passphrase.ConstantRetriever("pass"))
	for _, role := range roles {
		for _, keyID := range from.ListKeys(role) {
			key, _, err := from.GetPrivateKey(keyID)
			require.NoError(t, err)
			memKeyStore.AddKey(trustmanager.KeyInfo{Role: role}, key)
		}
	}
	return cryptoservice.NewCryptoService(memKeyStore)
}

// EmptyRepo creates an in memory crypto service
// and initializes a repo with no targets.  Delegations are only created
// if delegation roles are passed in.
func EmptyRepo(gun data.GUN, delegationRoles ...data.RoleName) (*tuf.Repo, signed.CryptoService, error) {
	cs := cryptoservice.NewCryptoService(trustmanager.NewKeyMemoryStore(passphrase.ConstantRetriever("")))
	r := tuf.NewRepo(cs)

	baseRoles := map[data.RoleName]data.BaseRole{}
	for _, role := range data.BaseRoles {
		key, err := CreateKey(cs, gun, role, data.ECDSAKey)
		if err != nil {
			return nil, nil, err
		}
		baseRoles[role] = data.NewBaseRole(
			role,
			1,
			key,
		)
	}

	r.InitRoot(
		baseRoles[data.CanonicalRootRole],
		baseRoles[data.CanonicalTimestampRole],
		baseRoles[data.CanonicalSnapshotRole],
		baseRoles[data.CanonicalTargetsRole],
		false,
	)
	r.InitTargets(data.CanonicalTargetsRole)
	r.InitSnapshot()
	r.InitTimestamp()

	// sort the delegation roles so that we make sure to create the parents
	// first
	// TODO: go back and fix this when we upgrade to Go 1.8 with the new
	//       slice sorting support. We should only need to define a `Less(i, j {}interface)`
	//       on RoleName to be able to call sort.Slice(delegationRoles) (or something like that)
	var roleNames []string
	for _, role := range delegationRoles {
		roleNames = append(roleNames, role.String())
	}

	sort.Strings(roleNames)
	for _, delgName := range roleNames {
		// create a delegations key and a delegation in the TUF repo
		delgKey, err := CreateKey(cs, gun, data.RoleName(delgName), data.ECDSAKey)
		if err != nil {
			return nil, nil, err
		}
		if err := r.UpdateDelegationKeys(data.RoleName(delgName), []data.PublicKey{delgKey}, []string{}, 1); err != nil {
			return nil, nil, err
		}
		if err := r.UpdateDelegationPaths(data.RoleName(delgName), []string{""}, []string{}, false); err != nil {
			return nil, nil, err
		}
	}

	return r, cs, nil
}

// NewRepoMetadata creates a TUF repo and returns the metadata
func NewRepoMetadata(gun data.GUN, delegationRoles ...data.RoleName) (map[data.RoleName][]byte, signed.CryptoService, error) {
	tufRepo, cs, err := EmptyRepo(gun, delegationRoles...)
	if err != nil {
		return nil, nil, err
	}

	meta, err := SignAndSerialize(tufRepo)
	if err != nil {
		return nil, nil, err
	}

	return meta, cs, nil
}

// CopyRepoMetadata makes a copy of a metadata->bytes mapping
func CopyRepoMetadata(from map[data.RoleName][]byte) map[data.RoleName][]byte {
	copied := make(map[data.RoleName][]byte)
	for roleName, metaBytes := range from {
		copied[roleName] = metaBytes
	}
	return copied
}

// SignAndSerialize calls Sign and then Serialize to get the repo metadata out
func SignAndSerialize(tufRepo *tuf.Repo) (map[data.RoleName][]byte, error) {
	meta := make(map[data.RoleName][]byte)

	for delgName := range tufRepo.Targets {
		// we'll sign targets later
		if delgName == data.CanonicalTargetsRole {
			continue
		}

		signedThing, err := tufRepo.SignTargets(delgName, data.DefaultExpires("targets"))
		if err != nil {
			return nil, err
		}
		metaBytes, err := json.MarshalCanonical(signedThing)
		if err != nil {
			return nil, err
		}

		meta[delgName] = metaBytes
	}

	// these need to be generated after the delegations are created and signed so
	// the snapshot will have the delegation metadata
	rs, tgs, ss, ts, err := Sign(tufRepo)
	if err != nil {
		return nil, err
	}

	rf, tgf, sf, tf, err := Serialize(rs, tgs, ss, ts)
	if err != nil {
		return nil, err
	}

	meta[data.CanonicalRootRole] = rf
	meta[data.CanonicalSnapshotRole] = sf
	meta[data.CanonicalTargetsRole] = tgf
	meta[data.CanonicalTimestampRole] = tf

	return meta, nil
}

// Sign signs all top level roles in a repo in the appropriate order
func Sign(repo *tuf.Repo) (root, targets, snapshot, timestamp *data.Signed, err error) {
	root, err = repo.SignRoot(data.DefaultExpires("root"), nil)
	if _, ok := err.(data.ErrInvalidRole); err != nil && !ok {
		return nil, nil, nil, nil, err
	}
	targets, err = repo.SignTargets(data.CanonicalTargetsRole, data.DefaultExpires("targets"))
	if _, ok := err.(data.ErrInvalidRole); err != nil && !ok {
		return nil, nil, nil, nil, err
	}
	snapshot, err = repo.SignSnapshot(data.DefaultExpires("snapshot"))
	if _, ok := err.(data.ErrInvalidRole); err != nil && !ok {
		return nil, nil, nil, nil, err
	}
	timestamp, err = repo.SignTimestamp(data.DefaultExpires("timestamp"))
	if _, ok := err.(data.ErrInvalidRole); err != nil && !ok {
		return nil, nil, nil, nil, err
	}
	return
}

// Serialize takes the Signed objects for the 4 top level roles and serializes them all to JSON
func Serialize(sRoot, sTargets, sSnapshot, sTimestamp *data.Signed) (root, targets, snapshot, timestamp []byte, err error) {
	root, err = json.Marshal(sRoot)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	targets, err = json.Marshal(sTargets)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	snapshot, err = json.Marshal(sSnapshot)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	timestamp, err = json.Marshal(sTimestamp)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return
}
