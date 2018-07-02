package trustpinning_test

import (
	"bytes"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cloudflare/cfssl/config"
	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/helpers"
	"github.com/cloudflare/cfssl/initca"
	"github.com/cloudflare/cfssl/signer"
	"github.com/cloudflare/cfssl/signer/local"
	"github.com/docker/go/canonical/json"
	"github.com/docker/notary"
	"github.com/docker/notary/cryptoservice"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/trustpinning"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/docker/notary/tuf/testutils"
	"github.com/docker/notary/tuf/utils"
	"github.com/stretchr/testify/require"
)

type SignedRSARootTemplate struct {
	RootPem string
}

var passphraseRetriever = func(string, string, bool, int) (string, bool, error) { return "passphrase", false, nil }

type rootData struct {
	rootMeta                      *data.Signed
	rootPubKeyID, targetsPubkeyID string
}

type certChain struct {
	rootCert, intermediateCert, leafCert []byte
	rootKey, intermediateKey, leafKey    data.PrivateKey
}

var (
	_sampleRootData  *rootData
	_sampleCertChain *certChain
)

func sampleRootData(t *testing.T) *rootData {
	if _sampleRootData == nil {
		var err error
		_sampleRootData = new(rootData)
		// generate a single test repo we can use for testing
		tufRepo, _, err := testutils.EmptyRepo("docker.com/notary")
		require.NoError(t, err)
		_sampleRootData.rootPubKeyID = tufRepo.Root.Signed.Roles[data.CanonicalRootRole].KeyIDs[0]
		_sampleRootData.targetsPubkeyID = tufRepo.Root.Signed.Roles[data.CanonicalTargetsRole].KeyIDs[0]
		tufRepo.Root.Signed.Version++
		_sampleRootData.rootMeta, err = tufRepo.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
		require.NoError(t, err)
	}
	return _sampleRootData
}

func sampleCertChain(t *testing.T) *certChain {
	if _sampleCertChain == nil {
		// generate a CA, an intermediate, and a leaf certificate using CFSSL
		// Create a simple CSR for the CA using the default CA validator and policy
		req := &csr.CertificateRequest{
			CN:         "docker.io/notary/root",
			KeyRequest: csr.NewBasicKeyRequest(),
			CA:         &csr.CAConfig{},
		}

		// Generate the CA and get the certificate and private key
		rootCert, _, rootKey, _ := initca.New(req)
		priv, _ := helpers.ParsePrivateKeyPEM(rootKey)
		cert, _ := helpers.ParseCertificatePEM(rootCert)
		s, _ := local.NewSigner(priv, cert, signer.DefaultSigAlgo(priv), initca.CAPolicy())

		req.CN = "docker.io/notary/intermediate"
		intCSR, intKey, _ := csr.ParseRequest(req)
		intCert, _ := s.Sign(signer.SignRequest{
			Request: string(intCSR),
			Subject: &signer.Subject{CN: req.CN},
		})

		priv, _ = helpers.ParsePrivateKeyPEM(intKey)
		cert, _ = helpers.ParseCertificatePEM(intCert)
		s, _ = local.NewSigner(priv, cert, signer.DefaultSigAlgo(priv), &config.Signing{
			Default: config.DefaultConfig(),
		})
		req.CA = nil
		req.CN = "docker.io/notary/leaf"
		leafCSR, leafKey, _ := csr.ParseRequest(req)
		leafCert, _ := s.Sign(signer.SignRequest{
			Request: string(leafCSR),
			Subject: &signer.Subject{CN: req.CN},
		})

		parsedRootKey, _ := utils.ParsePEMPrivateKey(rootKey, "")
		parsedIntKey, _ := utils.ParsePEMPrivateKey(intKey, "")
		parsedLeafKey, _ := utils.ParsePEMPrivateKey(leafKey, "")

		_sampleCertChain = &certChain{
			rootCert:         rootCert,
			intermediateCert: intCert,
			leafCert:         leafCert,

			rootKey:         parsedRootKey,
			intermediateKey: parsedIntKey,
			leafKey:         parsedLeafKey,
		}
	}
	return _sampleCertChain
}

func TestValidateRoot(t *testing.T) {
	// This call to trustpinning.ValidateRoot will succeed since we are using a valid PEM
	// encoded certificate, and have no other certificates for this CN
	_, err := trustpinning.ValidateRoot(nil, sampleRootData(t).rootMeta, "docker.com/notary", trustpinning.TrustPinConfig{})
	require.NoError(t, err)

	// This call to trustpinning.ValidateRoot will fail since we are passing in a dnsName that
	// doesn't match the CN of the certificate.
	_, err = trustpinning.ValidateRoot(nil, sampleRootData(t).rootMeta, "diogomonica.com/notary", trustpinning.TrustPinConfig{})
	require.Error(t, err, "An error was expected")
	require.Equal(t, err, &trustpinning.ErrValidationFail{Reason: "unable to retrieve valid leaf certificates"})

	// --- now we mess around with changing the keys, so we need to create a custom TUF repo that we can re-sign
	tufRepo, cs, err := testutils.EmptyRepo("docker.com/notary")
	require.NoError(t, err)
	tufRepo.Root.Signed.Version++
	rootKeyID := tufRepo.Root.Signed.Roles[data.CanonicalRootRole].KeyIDs[0]
	pubKey := tufRepo.Root.Signed.Keys[rootKeyID]

	rootMeta, err := tufRepo.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
	require.NoError(t, err)

	//
	// This call to trustpinning.ValidateRoot will fail since we are passing an unparsable RootSigned
	//
	keyBytes, err := json.MarshalCanonical(&pubKey)
	require.NoError(t, err)

	rawJSONBytes, err := json.Marshal(rootMeta.Signed)
	require.NoError(t, err)
	rawJSONBytes = bytes.Replace(rawJSONBytes, keyBytes, []byte(`"------ ABSOLUTELY NOT A BASE64 PEM -------"`), -1)
	require.NoError(t, json.Unmarshal(rawJSONBytes, rootMeta.Signed))
	require.NoError(t, signed.Sign(cs, rootMeta, []data.PublicKey{pubKey}, 1, nil))

	_, err = trustpinning.ValidateRoot(nil, rootMeta, "docker.com/notary", trustpinning.TrustPinConfig{})
	require.Error(t, err, "illegal base64 data at input byte")

	//
	// This call to trustpinning.ValidateRoot will fail since we are passing an invalid PEM cert
	//
	tufRepo.Root.Signed.Keys[rootKeyID] = data.NewECDSAx509PublicKey([]byte("-----BEGIN CERTIFICATE-----\ninvalid PEM\n-----END CERTIFICATE-----\n"))
	rootMeta, err = tufRepo.Root.ToSigned()
	require.NoError(t, err)
	require.NoError(t, signed.Sign(cs, rootMeta, []data.PublicKey{pubKey}, 1, nil))

	_, err = trustpinning.ValidateRoot(nil, rootMeta, "docker.com/notary", trustpinning.TrustPinConfig{})
	require.Error(t, err, "An error was expected")
	require.Equal(t, err, &trustpinning.ErrValidationFail{Reason: "unable to retrieve valid leaf certificates"})

	tufRepo.Root.Signed.Keys[rootKeyID] = pubKey // put things back the way they were

	//
	// This call to trustpinning.ValidateRoot will fail since we are passing only CA certificate
	// This will fail due to the lack of a leaf certificate
	//
	pubKey = data.NewECDSAx509PublicKey(sampleCertChain(t).rootCert)
	tufRepo.Root.Signed.Keys[pubKey.ID()] = pubKey
	tufRepo.Root.Signed.Roles[data.CanonicalRootRole].KeyIDs = []string{pubKey.ID()}
	require.NoError(t, cs.AddKey(data.CanonicalRootRole, "docker.io/notary/root", sampleCertChain(t).rootKey))

	rootMeta, err = tufRepo.Root.ToSigned()
	require.NoError(t, err)
	require.NoError(t, signed.Sign(cs, rootMeta, []data.PublicKey{pubKey}, 1, nil))

	_, err = trustpinning.ValidateRoot(nil, rootMeta, "secure.example.com", trustpinning.TrustPinConfig{})
	require.Error(t, err, "An error was expected")
	require.Equal(t, err, &trustpinning.ErrValidationFail{Reason: "unable to retrieve valid leaf certificates"})

	//
	// This call to trustpinning.ValidateRoot could succeed in getting to the TUF validation, since
	// we are using a valid PEM encoded certificate chain of intermediate + leaf cert
	// that are signed by a trusted root authority and the leaf cert has a correct CN.
	// It will, however, fail to validate, because the leaf cert does not precede the
	// intermediate in the certificate bundle
	//
	pubKey = data.NewECDSAx509PublicKey(
		append(append(sampleCertChain(t).intermediateCert, sampleCertChain(t).leafCert...), sampleCertChain(t).rootCert...))
	tufRepo.Root.Signed.Keys[pubKey.ID()] = pubKey
	tufRepo.Root.Signed.Roles[data.CanonicalRootRole].KeyIDs = []string{pubKey.ID()}
	require.NoError(t, cs.AddKey(data.CanonicalRootRole, "docker.io/notary/intermediate", sampleCertChain(t).intermediateKey))
	require.NoError(t, cs.AddKey(data.CanonicalRootRole, "docker.io/notary/leaf", sampleCertChain(t).leafKey))

	rootMeta, err = tufRepo.Root.ToSigned()
	require.NoError(t, err)
	require.NoError(t, signed.Sign(cs, rootMeta, []data.PublicKey{pubKey}, 1, nil))

	_, err = trustpinning.ValidateRoot(nil, rootMeta, "docker.io/notary/intermediate", trustpinning.TrustPinConfig{})
	require.Error(t, err, "An error was expected")
	require.Equal(t, err, &trustpinning.ErrValidationFail{Reason: "unable to retrieve valid leaf certificates"})

	//
	// This call to trustpinning.ValidateRoot will succeed in getting to the TUF validation, since
	// we are using a valid PEM encoded certificate chain of leaf cert + intermediate cert + root cert
	pubKey = data.NewECDSAx509PublicKey(
		append(append(sampleCertChain(t).leafCert, sampleCertChain(t).intermediateCert...), sampleCertChain(t).rootCert...))
	tufRepo.Root.Signed.Keys[pubKey.ID()] = pubKey
	tufRepo.Root.Signed.Roles[data.CanonicalRootRole].KeyIDs = []string{pubKey.ID()}

	rootMeta, err = tufRepo.Root.ToSigned()
	require.NoError(t, err)
	require.NoError(t, signed.Sign(cs, rootMeta, []data.PublicKey{pubKey}, 1, nil))

	_, err = trustpinning.ValidateRoot(nil, rootMeta, "docker.io/notary/leaf", trustpinning.TrustPinConfig{})
	require.NoError(t, err)
}

func TestValidateRootWithoutTOFUS(t *testing.T) {
	// This call to trustpinning.ValidateRoot will fail since we are explicitly disabling TOFU and have no local certs
	_, err := trustpinning.ValidateRoot(nil, sampleRootData(t).rootMeta, "docker.com/notary", trustpinning.TrustPinConfig{DisableTOFU: true})
	require.Error(t, err)
}

func TestValidateRootWithPinnedCert(t *testing.T) {
	typedSignedRoot, err := data.RootFromSigned(sampleRootData(t).rootMeta)
	require.NoError(t, err)

	// This call to trustpinning.ValidateRoot should succeed with the correct Cert ID (same as root public key ID)
	validatedSignedRoot, err := trustpinning.ValidateRoot(nil, sampleRootData(t).rootMeta, "docker.com/notary",
		trustpinning.TrustPinConfig{Certs: map[string][]string{"docker.com/notary": {sampleRootData(t).rootPubKeyID}}, DisableTOFU: true})
	require.NoError(t, err)
	typedSignedRoot.Signatures[0].IsValid = true
	require.Equal(t, validatedSignedRoot, typedSignedRoot)

	// This call to trustpinning.ValidateRoot should also succeed with the correct Cert ID (same as root public key ID), even though we passed an extra bad one
	validatedSignedRoot, err = trustpinning.ValidateRoot(nil, sampleRootData(t).rootMeta, "docker.com/notary",
		trustpinning.TrustPinConfig{Certs: map[string][]string{"docker.com/notary": {sampleRootData(t).rootPubKeyID, "invalidID"}}, DisableTOFU: true})
	require.NoError(t, err)
	// This extra assignment is necessary because ValidateRoot calls through to a successful VerifySignature which marks IsValid
	typedSignedRoot.Signatures[0].IsValid = true
	require.Equal(t, typedSignedRoot, validatedSignedRoot)
}

func TestValidateRootWithPinnedCertAndIntermediates(t *testing.T) {
	now := time.Now()
	memStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cs := cryptoservice.NewCryptoService(memStore)

	ecdsax509Key := data.NewECDSAx509PublicKey(append(sampleCertChain(t).leafCert, sampleCertChain(t).intermediateCert...))
	require.NoError(t, cs.AddKey(data.CanonicalRootRole, "docker.io/notary/leaf", sampleCertChain(t).leafKey))

	otherKey, err := cs.Create(data.CanonicalTargetsRole, "docker.io/notary/leaf", data.ED25519Key)
	require.NoError(t, err)

	root := data.SignedRoot{
		Signatures: make([]data.Signature, 0),
		Signed: data.Root{
			SignedCommon: data.SignedCommon{
				Type:    "Root",
				Expires: now.Add(time.Hour),
				Version: 1,
			},
			Keys: map[string]data.PublicKey{
				ecdsax509Key.ID(): ecdsax509Key,
				otherKey.ID():     otherKey,
			},
			Roles: map[data.RoleName]*data.RootRole{
				"root": {
					KeyIDs:    []string{ecdsax509Key.ID()},
					Threshold: 1,
				},
				"targets": {
					KeyIDs:    []string{otherKey.ID()},
					Threshold: 1,
				},
				"snapshot": {
					KeyIDs:    []string{otherKey.ID()},
					Threshold: 1,
				},
				"timestamp": {
					KeyIDs:    []string{otherKey.ID()},
					Threshold: 1,
				},
			},
		},
		Dirty: true,
	}

	signedRoot, err := root.ToSigned()
	require.NoError(t, err)
	err = signed.Sign(cs, signedRoot, []data.PublicKey{ecdsax509Key}, 1, nil)
	require.NoError(t, err)

	typedSignedRoot, err := data.RootFromSigned(signedRoot)
	require.NoError(t, err)

	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	defer os.RemoveAll(tempBaseDir)
	require.NoError(t, err, "failed to create a temporary directory: %s", err)

	validatedRoot, err := trustpinning.ValidateRoot(
		nil,
		signedRoot,
		"docker.io/notary/leaf",
		trustpinning.TrustPinConfig{
			Certs: map[string][]string{
				"docker.io/notary/leaf": {ecdsax509Key.ID()},
			},
			DisableTOFU: true,
		},
	)
	require.NoError(t, err, "failed to validate certID with intermediate")
	for idx, sig := range typedSignedRoot.Signatures {
		if sig.KeyID == ecdsax509Key.ID() {
			typedSignedRoot.Signatures[idx].IsValid = true
		}
	}
	require.Equal(t, typedSignedRoot, validatedRoot)

	// test it also works with a wildcarded gun in certs
	validatedRoot, err = trustpinning.ValidateRoot(
		nil,
		signedRoot,
		"docker.io/notary/leaf",
		trustpinning.TrustPinConfig{
			Certs: map[string][]string{
				"docker.io/notar*": {ecdsax509Key.ID()},
			},
			DisableTOFU: true,
		},
	)
	require.NoError(t, err, "failed to validate certID with intermediate")
	for idx, sig := range typedSignedRoot.Signatures {
		if sig.KeyID == ecdsax509Key.ID() {
			typedSignedRoot.Signatures[idx].IsValid = true
		}
	}
	require.Equal(t, typedSignedRoot, validatedRoot)

	// incorrect key id on wildcard match should fail
	_, err = trustpinning.ValidateRoot(
		nil,
		signedRoot,
		"docker.io/notary/leaf",
		trustpinning.TrustPinConfig{
			Certs: map[string][]string{
				"docker.io/notar*": {"badID"},
			},
			DisableTOFU: true,
		},
	)
	require.Error(t, err, "failed to validate certID with intermediate")

	// exact match should take precedence even if it fails validation
	_, err = trustpinning.ValidateRoot(
		nil,
		signedRoot,
		"docker.io/notary/leaf",
		trustpinning.TrustPinConfig{
			Certs: map[string][]string{
				"docker.io/notary/leaf": {"badID"},
				"docker.io/notar*":      {ecdsax509Key.ID()},
			},
			DisableTOFU: true,
		},
	)
	require.Error(t, err, "failed to validate certID with intermediate")

	// exact match should take precedence
	validatedRoot, err = trustpinning.ValidateRoot(
		nil,
		signedRoot,
		"docker.io/notary/leaf",
		trustpinning.TrustPinConfig{
			Certs: map[string][]string{
				"docker.io/notary/leaf": {ecdsax509Key.ID()},
				"docker.io/notar*":      {"badID"},
			},
			DisableTOFU: true,
		},
	)
	require.NoError(t, err, "failed to validate certID with intermediate")
	for idx, sig := range typedSignedRoot.Signatures {
		if sig.KeyID == ecdsax509Key.ID() {
			typedSignedRoot.Signatures[idx].IsValid = true
		}
	}
	require.Equal(t, typedSignedRoot, validatedRoot)
}

func TestValidateRootFailuresWithPinnedCert(t *testing.T) {
	typedSignedRoot, err := data.RootFromSigned(sampleRootData(t).rootMeta)
	require.NoError(t, err)

	// This call to trustpinning.ValidateRoot should fail due to an incorrect cert ID
	_, err = trustpinning.ValidateRoot(nil, sampleRootData(t).rootMeta, "docker.com/notary",
		trustpinning.TrustPinConfig{Certs: map[string][]string{"docker.com/notary": {"ABSOLUTELY NOT A CERT ID"}}, DisableTOFU: true})
	require.Error(t, err)

	// This call to trustpinning.ValidateRoot should fail due to an empty cert ID
	_, err = trustpinning.ValidateRoot(nil, sampleRootData(t).rootMeta, "docker.com/notary",
		trustpinning.TrustPinConfig{Certs: map[string][]string{"docker.com/notary": {""}}, DisableTOFU: true})
	require.Error(t, err)

	// This call to trustpinning.ValidateRoot should fail due to an invalid GUN (even though the cert ID is correct), and TOFUS is set to false
	_, err = trustpinning.ValidateRoot(nil, sampleRootData(t).rootMeta, "docker.com/notary",
		trustpinning.TrustPinConfig{Certs: map[string][]string{"not_a_gun": {sampleRootData(t).rootPubKeyID}}, DisableTOFU: true})
	require.Error(t, err)

	// This call to trustpinning.ValidateRoot should fail due to an invalid cert ID, even though it's a valid key ID for targets
	_, err = trustpinning.ValidateRoot(nil, sampleRootData(t).rootMeta, "docker.com/notary",
		trustpinning.TrustPinConfig{Certs: map[string][]string{"docker.com/notary": {sampleRootData(t).targetsPubkeyID}}, DisableTOFU: true})
	require.Error(t, err)

	// This call to trustpinning.ValidateRoot should succeed because we fall through to TOFUS because we have no matching GUNs under Certs
	validatedRoot, err := trustpinning.ValidateRoot(nil, sampleRootData(t).rootMeta, "docker.com/notary",
		trustpinning.TrustPinConfig{Certs: map[string][]string{"not_a_gun": {sampleRootData(t).rootPubKeyID}}, DisableTOFU: false})
	require.NoError(t, err)
	typedSignedRoot.Signatures[0].IsValid = true
	require.Equal(t, typedSignedRoot, validatedRoot)
}

func TestValidateRootWithPinnedCA(t *testing.T) {
	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory: %s", err)
	defer os.RemoveAll(tempBaseDir)

	typedSignedRoot, err := data.RootFromSigned(sampleRootData(t).rootMeta)
	require.NoError(t, err)

	// This call to trustpinning.ValidateRoot will fail because we have an invalid path for the CA
	_, err = trustpinning.ValidateRoot(nil, sampleRootData(t).rootMeta, "docker.com/notary",
		trustpinning.TrustPinConfig{CA: map[string]string{"docker.com/notary": filepath.Join(tempBaseDir, "nonexistent")}})
	require.Error(t, err)

	// This call to trustpinning.ValidateRoot will fail because we have no valid GUNs to use, and TOFUS is disabled
	_, err = trustpinning.ValidateRoot(nil, sampleRootData(t).rootMeta, "docker.com/notary",
		trustpinning.TrustPinConfig{CA: map[string]string{"othergun": filepath.Join(tempBaseDir, "nonexistent")}, DisableTOFU: true})
	require.Error(t, err)

	// This call to trustpinning.ValidateRoot will succeed because we have no valid GUNs to use and we fall back to enabled TOFUS
	validatedRoot, err := trustpinning.ValidateRoot(nil, sampleRootData(t).rootMeta, "docker.com/notary",
		trustpinning.TrustPinConfig{CA: map[string]string{"othergun": filepath.Join(tempBaseDir, "nonexistent")}, DisableTOFU: false})
	require.NoError(t, err)
	typedSignedRoot.Signatures[0].IsValid = true
	require.Equal(t, typedSignedRoot, validatedRoot)

	// Write an invalid CA cert (not even a PEM) to the tempDir and ensure validation fails when using it
	invalidCAFilepath := filepath.Join(tempBaseDir, "invalid.ca")
	require.NoError(t, ioutil.WriteFile(invalidCAFilepath, []byte("ABSOLUTELY NOT A PEM"), 0644))

	// Using this invalid CA cert should fail on trustpinning.ValidateRoot
	_, err = trustpinning.ValidateRoot(nil, sampleRootData(t).rootMeta, "docker.com/notary",
		trustpinning.TrustPinConfig{CA: map[string]string{"docker.com/notary": invalidCAFilepath}, DisableTOFU: true})
	require.Error(t, err)

	validCAFilepath := "../fixtures/root-ca.crt"

	// If we pass an invalid Certs entry in addition to this valid CA entry, since Certs has priority for pinning we will fail
	_, err = trustpinning.ValidateRoot(nil, sampleRootData(t).rootMeta, "docker.com/notary",
		trustpinning.TrustPinConfig{
			Certs:       map[string][]string{"docker.com/notary": {"invalidID"}},
			CA:          map[string]string{"docker.com/notary": validCAFilepath},
			DisableTOFU: true})
	require.Error(t, err)

	// Now construct a new root with a valid cert chain, such that signatures are correct over the 'notary-signer' GUN.  Pin the root-ca and validate
	certChain, err := utils.LoadCertBundleFromFile("../fixtures/notary-signer.crt")
	require.NoError(t, err)

	pemChainBytes, err := utils.CertChainToPEM(certChain)
	require.NoError(t, err)

	newRootKey := data.NewPublicKey(data.RSAx509Key, pemChainBytes)

	rootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{newRootKey.ID()}, nil)
	require.NoError(t, err)

	testRoot, err := data.NewRoot(
		map[string]data.PublicKey{newRootKey.ID(): newRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole},
		false,
	)
	testRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	keyReader, err := os.Open("../fixtures/notary-signer.key")
	require.NoError(t, err, "could not open key file")
	pemBytes, err := ioutil.ReadAll(keyReader)
	require.NoError(t, err, "could not read key file")
	privKey, err := utils.ParsePEMPrivateKey(pemBytes, "")
	require.NoError(t, err)

	store, err := trustmanager.NewKeyFileStore(tempBaseDir, passphraseRetriever)
	require.NoError(t, err)
	cs := cryptoservice.NewCryptoService(store)

	err = store.AddKey(trustmanager.KeyInfo{Role: data.CanonicalRootRole, Gun: "notary-signer"}, privKey)
	require.NoError(t, err)

	newTestSignedRoot, err := testRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, newTestSignedRoot, []data.PublicKey{newRootKey}, 1, nil)
	require.NoError(t, err)

	newTypedSignedRoot, err := data.RootFromSigned(newTestSignedRoot)
	require.NoError(t, err)

	// Check that we validate correctly against a pinned CA and provided bundle
	validatedRoot, err = trustpinning.ValidateRoot(nil, newTestSignedRoot, "notary-signer", trustpinning.TrustPinConfig{CA: map[string]string{"notary-signer": validCAFilepath}, DisableTOFU: true})
	require.NoError(t, err)
	for idx, sig := range newTypedSignedRoot.Signatures {
		if sig.KeyID == newRootKey.ID() {
			newTypedSignedRoot.Signatures[idx].IsValid = true
		}
	}
	require.Equal(t, newTypedSignedRoot, validatedRoot)

	// Add an expired CA for the same gun to our previous pinned bundle, ensure that we still validate correctly
	goodRootCABundle, err := utils.LoadCertBundleFromFile(validCAFilepath)
	require.NoError(t, err)
	memKeyStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cryptoService := cryptoservice.NewCryptoService(memKeyStore)
	testPubKey, err := cryptoService.Create("root", "notary-signer", data.ECDSAKey)
	require.NoError(t, err)
	testPrivKey, _, err := memKeyStore.GetKey(testPubKey.ID())
	require.NoError(t, err)
	expiredCert, err := generateExpiredTestingCertificate(testPrivKey, "notary-signer")
	require.NoError(t, err)
	bundleWithExpiredCert, err := utils.CertChainToPEM(append(goodRootCABundle, expiredCert))
	require.NoError(t, err)
	bundleWithExpiredCertPath := filepath.Join(tempBaseDir, "bundle_with_expired_cert.pem")
	require.NoError(t, ioutil.WriteFile(bundleWithExpiredCertPath, bundleWithExpiredCert, 0644))

	// Check that we validate correctly against a pinned CA and provided bundle
	validatedRoot, err = trustpinning.ValidateRoot(nil, newTestSignedRoot, "notary-signer", trustpinning.TrustPinConfig{CA: map[string]string{"notary-signer": bundleWithExpiredCertPath}, DisableTOFU: true})
	require.NoError(t, err)
	require.Equal(t, newTypedSignedRoot, validatedRoot)

	testPubKey2, err := cryptoService.Create("root", "notary-signer", data.ECDSAKey)
	require.NoError(t, err)
	testPrivKey2, _, err := memKeyStore.GetKey(testPubKey2.ID())
	require.NoError(t, err)
	expiredCert2, err := generateExpiredTestingCertificate(testPrivKey2, "notary-signer")
	require.NoError(t, err)
	allExpiredCertBundle, err := utils.CertChainToPEM([]*x509.Certificate{expiredCert, expiredCert2})
	require.NoError(t, err)
	allExpiredCertPath := filepath.Join(tempBaseDir, "all_expired_cert.pem")
	require.NoError(t, ioutil.WriteFile(allExpiredCertPath, allExpiredCertBundle, 0644))
	// Now only use expired certs in the bundle, we should fail
	_, err = trustpinning.ValidateRoot(nil, newTestSignedRoot, "notary-signer", trustpinning.TrustPinConfig{CA: map[string]string{"notary-signer": allExpiredCertPath}, DisableTOFU: true})
	require.Error(t, err)

	// Add a CA cert for a that won't validate against the root leaf certificate
	testPubKey3, err := cryptoService.Create("root", "notary-signer", data.ECDSAKey)
	require.NoError(t, err)
	testPrivKey3, _, err := memKeyStore.GetKey(testPubKey3.ID())
	require.NoError(t, err)
	validCert, err := cryptoservice.GenerateCertificate(testPrivKey3, "notary-signer", time.Now(), time.Now().AddDate(1, 0, 0))
	require.NoError(t, err)
	bundleWithWrongCert, err := utils.CertChainToPEM([]*x509.Certificate{validCert})
	require.NoError(t, err)
	bundleWithWrongCertPath := filepath.Join(tempBaseDir, "bundle_with_expired_cert.pem")
	require.NoError(t, ioutil.WriteFile(bundleWithWrongCertPath, bundleWithWrongCert, 0644))
	_, err = trustpinning.ValidateRoot(nil, newTestSignedRoot, "notary-signer", trustpinning.TrustPinConfig{CA: map[string]string{"notary-signer": bundleWithWrongCertPath}, DisableTOFU: true})
	require.Error(t, err)
}

// TestValidateSuccessfulRootRotation runs through a full root certificate rotation
// We test this with both an RSA and ECDSA root certificate
func TestValidateSuccessfulRootRotation(t *testing.T) {
	testValidateSuccessfulRootRotation(t, data.ECDSAKey, data.ECDSAx509Key)
	if !testing.Short() {
		testValidateSuccessfulRootRotation(t, data.RSAKey, data.RSAx509Key)
	}
}

func testValidateSuccessfulRootRotation(t *testing.T, keyAlg, rootKeyType string) {
	// The gun to test
	var gun data.GUN = "docker.com/notary"

	memKeyStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cs := cryptoservice.NewCryptoService(memKeyStore)

	// TUF key with PEM-encoded x509 certificate
	origRootKey, err := testutils.CreateKey(cs, gun, data.CanonicalRootRole, keyAlg)
	require.NoError(t, err)

	origRootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{origRootKey.ID()}, nil)
	require.NoError(t, err)

	origTestRoot, err := data.NewRoot(
		map[string]data.PublicKey{origRootKey.ID(): origRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &origRootRole.RootRole,
			data.CanonicalTargetsRole:   &origRootRole.RootRole,
			data.CanonicalSnapshotRole:  &origRootRole.RootRole,
			data.CanonicalTimestampRole: &origRootRole.RootRole,
		},
		false,
	)
	origTestRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	signedOrigTestRoot, err := origTestRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, signedOrigTestRoot, []data.PublicKey{origRootKey}, 1, nil)
	require.NoError(t, err)
	prevRoot, err := data.RootFromSigned(signedOrigTestRoot)
	require.NoError(t, err)

	// TUF key with PEM-encoded x509 certificate
	replRootKey, err := testutils.CreateKey(cs, gun, data.CanonicalRootRole, keyAlg)
	require.NoError(t, err)

	rootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{replRootKey.ID()}, nil)
	require.NoError(t, err)

	testRoot, err := data.NewRoot(
		map[string]data.PublicKey{replRootKey.ID(): replRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole},
		false,
	)
	testRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	signedTestRoot, err := testRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, signedTestRoot, []data.PublicKey{replRootKey, origRootKey}, 2, nil)
	require.NoError(t, err)

	typedSignedRoot, err := data.RootFromSigned(signedTestRoot)
	require.NoError(t, err)

	// This call to trustpinning.ValidateRoot will succeed since we are using a valid PEM
	// encoded certificate, and have no other certificates for this CN
	validatedRoot, err := trustpinning.ValidateRoot(prevRoot, signedTestRoot, gun, trustpinning.TrustPinConfig{})
	require.NoError(t, err)
	for idx, sig := range typedSignedRoot.Signatures {
		if sig.KeyID == replRootKey.ID() {
			typedSignedRoot.Signatures[idx].IsValid = true
		}
	}
	require.Equal(t, typedSignedRoot, validatedRoot)
}

// TestValidateRootRotationMissingOrigSig runs through a full root certificate rotation
// where we are missing the original root key signature. Verification should fail.
// We test this with both an RSA and ECDSA root certificate
func TestValidateRootRotationMissingOrigSig(t *testing.T) {
	testValidateRootRotationMissingOrigSig(t, data.ECDSAKey, data.ECDSAx509Key)
	if !testing.Short() {
		testValidateRootRotationMissingOrigSig(t, data.RSAKey, data.RSAx509Key)
	}
}

func testValidateRootRotationMissingOrigSig(t *testing.T, keyAlg, rootKeyType string) {
	var gun data.GUN = "docker.com/notary"

	memKeyStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cs := cryptoservice.NewCryptoService(memKeyStore)

	// TUF key with PEM-encoded x509 certificate
	origRootKey, err := testutils.CreateKey(cs, gun, data.CanonicalRootRole, keyAlg)
	require.NoError(t, err)

	origRootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{origRootKey.ID()}, nil)
	require.NoError(t, err)

	origTestRoot, err := data.NewRoot(
		map[string]data.PublicKey{origRootKey.ID(): origRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &origRootRole.RootRole,
			data.CanonicalTargetsRole:   &origRootRole.RootRole,
			data.CanonicalSnapshotRole:  &origRootRole.RootRole,
			data.CanonicalTimestampRole: &origRootRole.RootRole,
		},
		false,
	)
	origTestRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	signedOrigTestRoot, err := origTestRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, signedOrigTestRoot, []data.PublicKey{origRootKey}, 1, nil)
	require.NoError(t, err)
	prevRoot, err := data.RootFromSigned(signedOrigTestRoot)
	require.NoError(t, err)

	// TUF key with PEM-encoded x509 certificate
	replRootKey, err := testutils.CreateKey(cs, gun, data.CanonicalRootRole, keyAlg)
	require.NoError(t, err)

	rootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{replRootKey.ID()}, nil)
	require.NoError(t, err)

	testRoot, err := data.NewRoot(
		map[string]data.PublicKey{replRootKey.ID(): replRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
		},
		false,
	)
	testRoot.Signed.Version = 2
	require.NoError(t, err, "Failed to create new root")

	signedTestRoot, err := testRoot.ToSigned()
	require.NoError(t, err)

	// We only sign with the new key, and not with the original one.
	err = signed.Sign(cs, signedTestRoot, []data.PublicKey{replRootKey}, 1, nil)
	require.NoError(t, err)

	// This call to trustpinning.ValidateRoot will fail since we don't have the original key's signature
	_, err = trustpinning.ValidateRoot(prevRoot, signedTestRoot, gun, trustpinning.TrustPinConfig{})
	require.Error(t, err, "insufficient signatures on root")

	// If we clear out an valid certs from the prevRoot, this will still fail
	prevRoot.Signed.Keys = nil
	_, err = trustpinning.ValidateRoot(prevRoot, signedTestRoot, gun, trustpinning.TrustPinConfig{})
	require.Error(t, err, "insufficient signatures on root")
}

// TestValidateRootRotationMissingNewSig runs through a full root certificate rotation
// where we are missing the new root key signature. Verification should fail.
// We test this with both an RSA and ECDSA root certificate
func TestValidateRootRotationMissingNewSig(t *testing.T) {
	testValidateRootRotationMissingNewSig(t, data.ECDSAKey, data.ECDSAx509Key)
	if !testing.Short() {
		testValidateRootRotationMissingNewSig(t, data.RSAKey, data.RSAx509Key)
	}
}

func testValidateRootRotationMissingNewSig(t *testing.T, keyAlg, rootKeyType string) {
	var gun data.GUN = "docker.com/notary"

	memKeyStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cs := cryptoservice.NewCryptoService(memKeyStore)

	// TUF key with PEM-encoded x509 certificate
	origRootKey, err := testutils.CreateKey(cs, gun, data.CanonicalRootRole, keyAlg)
	require.NoError(t, err)

	origRootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{origRootKey.ID()}, nil)
	require.NoError(t, err)

	origTestRoot, err := data.NewRoot(
		map[string]data.PublicKey{origRootKey.ID(): origRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &origRootRole.RootRole,
			data.CanonicalTargetsRole:   &origRootRole.RootRole,
			data.CanonicalSnapshotRole:  &origRootRole.RootRole,
			data.CanonicalTimestampRole: &origRootRole.RootRole,
		},
		false,
	)
	origTestRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	signedOrigTestRoot, err := origTestRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, signedOrigTestRoot, []data.PublicKey{origRootKey}, 1, nil)
	require.NoError(t, err)
	prevRoot, err := data.RootFromSigned(signedOrigTestRoot)
	require.NoError(t, err)

	// TUF key with PEM-encoded x509 certificate
	replRootKey, err := testutils.CreateKey(cs, gun, data.CanonicalRootRole, keyAlg)
	require.NoError(t, err)

	rootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{replRootKey.ID()}, nil)
	require.NoError(t, err)

	testRoot, err := data.NewRoot(
		map[string]data.PublicKey{replRootKey.ID(): replRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
		},
		false,
	)
	require.NoError(t, err, "Failed to create new root")

	signedTestRoot, err := testRoot.ToSigned()
	require.NoError(t, err)

	// We only sign with the old key, and not with the new one
	err = signed.Sign(cs, signedTestRoot, []data.PublicKey{origRootKey}, 1, nil)
	require.NoError(t, err)

	// This call to trustpinning.ValidateRoot will succeed since we are using a valid PEM
	// encoded certificate, and have no other certificates for this CN
	_, err = trustpinning.ValidateRoot(prevRoot, signedTestRoot, gun, trustpinning.TrustPinConfig{})
	require.Error(t, err, "insufficient signatures on root")
}

// TestValidateRootRotationTrustPinning runs a full root certificate rotation but ensures that
// the specified trust pinning is respected with the new root for the Certs and TOFUs settings
func TestValidateRootRotationTrustPinning(t *testing.T) {
	// The gun to test
	var gun data.GUN = "docker.com/notary"

	memKeyStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cs := cryptoservice.NewCryptoService(memKeyStore)

	// TUF key with PEM-encoded x509 certificate
	origRootKey, err := testutils.CreateKey(cs, gun, data.CanonicalRootRole, data.RSAKey)
	require.NoError(t, err)

	origRootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{origRootKey.ID()}, nil)
	require.NoError(t, err)

	origTestRoot, err := data.NewRoot(
		map[string]data.PublicKey{origRootKey.ID(): origRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &origRootRole.RootRole,
			data.CanonicalTargetsRole:   &origRootRole.RootRole,
			data.CanonicalSnapshotRole:  &origRootRole.RootRole,
			data.CanonicalTimestampRole: &origRootRole.RootRole,
		},
		false,
	)
	origTestRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	signedOrigTestRoot, err := origTestRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, signedOrigTestRoot, []data.PublicKey{origRootKey}, 1, nil)
	require.NoError(t, err)
	prevRoot, err := data.RootFromSigned(signedOrigTestRoot)
	require.NoError(t, err)

	// TUF key with PEM-encoded x509 certificate
	replRootKey, err := testutils.CreateKey(cs, gun, data.CanonicalRootRole, data.RSAKey)
	require.NoError(t, err)

	rootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{replRootKey.ID()}, nil)
	require.NoError(t, err)

	testRoot, err := data.NewRoot(
		map[string]data.PublicKey{replRootKey.ID(): replRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole},
		false,
	)
	testRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	signedTestRoot, err := testRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, signedTestRoot, []data.PublicKey{replRootKey, origRootKey}, 2, nil)
	require.NoError(t, err)

	typedSignedRoot, err := data.RootFromSigned(signedTestRoot)
	require.NoError(t, err)

	// This call to trustpinning.ValidateRoot will fail due to the trust pinning mismatch in certs
	invalidCertConfig := trustpinning.TrustPinConfig{
		Certs: map[string][]string{
			gun.String(): {origRootKey.ID()},
		},
		DisableTOFU: true,
	}
	_, err = trustpinning.ValidateRoot(prevRoot, signedTestRoot, gun, invalidCertConfig)
	require.Error(t, err)

	// This call will succeed since we include the new root cert ID (and the old one)
	validCertConfig := trustpinning.TrustPinConfig{
		Certs: map[string][]string{
			gun.String(): {origRootKey.ID(), replRootKey.ID()},
		},
		DisableTOFU: true,
	}

	validatedRoot, err := trustpinning.ValidateRoot(prevRoot, signedTestRoot, gun, validCertConfig)
	require.NoError(t, err)
	for idx, sig := range typedSignedRoot.Signatures {
		if sig.KeyID == replRootKey.ID() {
			typedSignedRoot.Signatures[idx].IsValid = true
		}
	}
	require.Equal(t, typedSignedRoot, validatedRoot)

	// This call will also succeed since we only need the new replacement root ID to be pinned
	validCertConfig = trustpinning.TrustPinConfig{
		Certs: map[string][]string{
			gun.String(): {replRootKey.ID()},
		},
		DisableTOFU: true,
	}
	validatedRoot, err = trustpinning.ValidateRoot(prevRoot, signedTestRoot, gun, validCertConfig)
	require.NoError(t, err)
	require.Equal(t, typedSignedRoot, validatedRoot)

	// Even if we disable TOFU in the trustpinning, since we have a previously trusted root we should honor a valid rotation
	validatedRoot, err = trustpinning.ValidateRoot(prevRoot, signedTestRoot, gun, trustpinning.TrustPinConfig{DisableTOFU: true})
	require.NoError(t, err)
	require.Equal(t, typedSignedRoot, validatedRoot)
}

// TestValidateRootRotationTrustPinningInvalidCA runs a full root certificate rotation but ensures that
// the specified trust pinning rejects the new root for not being signed by the specified CA
func TestValidateRootRotationTrustPinningInvalidCA(t *testing.T) {
	var gun data.GUN = "notary-signer"
	keyAlg := data.RSAKey
	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	defer os.RemoveAll(tempBaseDir)
	require.NoError(t, err, "failed to create a temporary directory: %s", err)

	leafCert, err := utils.LoadCertFromFile("../fixtures/notary-signer.crt")
	require.NoError(t, err)

	intermediateCert, err := utils.LoadCertFromFile("../fixtures/intermediate-ca.crt")
	require.NoError(t, err)

	pemChainBytes, err := utils.CertChainToPEM([]*x509.Certificate{leafCert, intermediateCert})
	require.NoError(t, err)

	origRootKey := data.NewPublicKey(data.RSAx509Key, pemChainBytes)

	rootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{origRootKey.ID()}, nil)
	require.NoError(t, err)

	testRoot, err := data.NewRoot(
		map[string]data.PublicKey{origRootKey.ID(): origRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole},
		false,
	)
	testRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	pemBytes, err := ioutil.ReadFile("../fixtures/notary-signer.key")
	require.NoError(t, err, "could not read key file")
	privKey, err := utils.ParsePEMPrivateKey(pemBytes, "")
	require.NoError(t, err)

	store, err := trustmanager.NewKeyFileStore(tempBaseDir, passphraseRetriever)
	require.NoError(t, err)
	cs := cryptoservice.NewCryptoService(store)

	err = store.AddKey(trustmanager.KeyInfo{Role: data.CanonicalRootRole, Gun: gun}, privKey)
	require.NoError(t, err)

	origSignedTestRoot, err := testRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, origSignedTestRoot, []data.PublicKey{origRootKey}, 1, nil)
	require.NoError(t, err)
	prevRoot, err := data.RootFromSigned(origSignedTestRoot)
	require.NoError(t, err)

	// generate a new TUF key with PEM-encoded x509 certificate, not signed by our pinned CA
	replRootKey, err := testutils.CreateKey(cs, gun, data.CanonicalRootRole, keyAlg)
	require.NoError(t, err)

	_, err = data.NewRole(data.CanonicalRootRole, 1, []string{replRootKey.ID()}, nil)
	require.NoError(t, err)
	newRoot, err := data.NewRoot(
		map[string]data.PublicKey{replRootKey.ID(): replRootKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole},
		false,
	)
	newRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	newSignedTestRoot, err := newRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, newSignedTestRoot, []data.PublicKey{replRootKey, origRootKey}, 2, nil)
	require.NoError(t, err)

	// Check that we respect the trust pinning on rotation
	validCAFilepath := "../fixtures/root-ca.crt"
	_, err = trustpinning.ValidateRoot(prevRoot, newSignedTestRoot, gun, trustpinning.TrustPinConfig{CA: map[string]string{gun.String(): validCAFilepath}, DisableTOFU: true})
	require.Error(t, err)
}

func generateTestingCertificate(rootKey data.PrivateKey, gun data.GUN, timeToExpire time.Duration) (*x509.Certificate, error) {
	startTime := time.Now()
	return cryptoservice.GenerateCertificate(rootKey, gun, startTime, startTime.Add(timeToExpire))
}

func generateExpiredTestingCertificate(rootKey data.PrivateKey, gun data.GUN) (*x509.Certificate, error) {
	startTime := time.Now().AddDate(-10, 0, 0)
	return cryptoservice.GenerateCertificate(rootKey, gun, startTime, startTime.AddDate(1, 0, 0))
}

func TestParsePEMPublicKey(t *testing.T) {
	var gun data.GUN = "notary"
	memStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cs := cryptoservice.NewCryptoService(memStore)

	// can parse ECDSA PEM
	ecdsaPubKey, err := cs.Create("root", "docker.io/notary/test", data.ECDSAKey)
	require.NoError(t, err)
	ecdsaPemBytes := pem.EncodeToMemory(&pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   ecdsaPubKey.Public(),
	})

	ecdsaParsedPubKey, err := utils.ParsePEMPublicKey(ecdsaPemBytes)
	require.NoError(t, err, "no key: %s", ecdsaParsedPubKey.Public())

	// can parse certificates
	ecdsaPrivKey, _, err := memStore.GetKey(ecdsaPubKey.ID())
	require.NoError(t, err)
	cert, err := generateTestingCertificate(ecdsaPrivKey, gun, notary.Day*30)
	require.NoError(t, err)
	ecdsaPubKeyFromCert, err := utils.ParsePEMPublicKey(utils.CertToPEM(cert))
	require.NoError(t, err)

	thatData := []byte{1, 2, 3, 4}
	sig, err := ecdsaPrivKey.Sign(rand.Reader, thatData, nil)
	require.NoError(t, err)
	err = signed.ECDSAVerifier{}.Verify(ecdsaPubKeyFromCert, sig, thatData)
	require.NoError(t, err)

	// can parse RSA PEM
	rsaPemBytes := []byte(`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA7HQxZ0fDsxPTFIABQXNX
i9b25AZWtBoR+k8myrrI0cb08ISoB2NBpYwDbxhxLvjN1OpjFzCOjbmK+sD2zCkt
Rxg1Z9NimY4J/p9uWF2EcRklmCqdHJ2KW7QD3j5uy7e7KsSyLPcsMtIrRYVtk2Z8
oGKEOQUsTudXoH0W9lVtBNgQi0S3FiuesRXKc0jDsZRXxtQUB0MzzRJ8zjgZbuKw
6XBlfidMEo3E10jQk8lrV1iio0xpkYuW+sbfefgNDyGBoSpsSG9Kh0sDHCyRteCm
zKJV1ck/b6x3x7eLNtsAErkJfp6aNKcvGrXMUgB/pZTaC4lpfxKq4s3+zY6sgabr
jwIDAQAB
-----END PUBLIC KEY-----`)
	_, err = utils.ParsePEMPublicKey(rsaPemBytes)
	require.NoError(t, err)

	// unsupported key type
	unsupportedPemBytes := pem.EncodeToMemory(&pem.Block{
		Type:    "PRIVATE KEY",
		Headers: nil,
		Bytes:   []byte{0, 0, 0, 0},
	})
	_, err = utils.ParsePEMPublicKey(unsupportedPemBytes)
	require.Error(t, err)

	// bad key
	badPemBytes := pem.EncodeToMemory(&pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   []byte{0, 0, 0, 0},
	})
	_, err = utils.ParsePEMPublicKey(badPemBytes)
	require.Error(t, err)
}

func TestCheckingCertExpiry(t *testing.T) {
	var gun data.GUN = "notary"
	memStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cs := cryptoservice.NewCryptoService(memStore)
	testPubKey, err := cs.Create(data.CanonicalRootRole, gun, data.ECDSAKey)
	require.NoError(t, err)
	testPrivKey, _, err := memStore.GetKey(testPubKey.ID())
	require.NoError(t, err)

	almostExpiredCert, err := generateTestingCertificate(testPrivKey, gun, notary.Day*30)
	require.NoError(t, err)
	almostExpiredPubKey, err := utils.ParsePEMPublicKey(utils.CertToPEM(almostExpiredCert))
	require.NoError(t, err)

	// set up a logrus logger to capture warning output
	origLevel := logrus.GetLevel()
	logrus.SetLevel(logrus.WarnLevel)
	defer logrus.SetLevel(origLevel)
	logBuf := bytes.NewBuffer(nil)
	logrus.SetOutput(logBuf)

	rootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{almostExpiredPubKey.ID()}, nil)
	require.NoError(t, err)
	testRoot, err := data.NewRoot(
		map[string]data.PublicKey{almostExpiredPubKey.ID(): almostExpiredPubKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole},
		false,
	)
	testRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	signedTestRoot, err := testRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, signedTestRoot, []data.PublicKey{almostExpiredPubKey}, 1, nil)
	require.NoError(t, err)

	// This is a valid root certificate, but check that we get a Warn-level message that the certificate is near expiry
	_, err = trustpinning.ValidateRoot(nil, signedTestRoot, gun, trustpinning.TrustPinConfig{})
	require.NoError(t, err)
	require.Contains(t, logBuf.String(), fmt.Sprintf("certificate with CN %s is near expiry", gun))

	expiredCert, err := generateExpiredTestingCertificate(testPrivKey, gun)
	require.NoError(t, err)
	expiredPubKey := utils.CertToKey(expiredCert)

	rootRole, err = data.NewRole(data.CanonicalRootRole, 1, []string{expiredPubKey.ID()}, nil)
	require.NoError(t, err)
	testRoot, err = data.NewRoot(
		map[string]data.PublicKey{expiredPubKey.ID(): expiredPubKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole},
		false,
	)
	testRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	signedTestRoot, err = testRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, signedTestRoot, []data.PublicKey{expiredPubKey}, 1, nil)
	require.NoError(t, err)

	// This is an invalid root certificate since it's expired
	_, err = trustpinning.ValidateRoot(nil, signedTestRoot, gun, trustpinning.TrustPinConfig{})
	require.Error(t, err)
}

func TestValidateRootWithExpiredIntermediate(t *testing.T) {
	now := time.Now()
	memStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cs := cryptoservice.NewCryptoService(memStore)

	rootCert, err := helpers.ParseCertificatePEM(sampleCertChain(t).rootCert)
	require.NoError(t, err)

	expTemplate, err := helpers.ParseCertificatePEM(sampleCertChain(t).intermediateCert)
	require.NoError(t, err)
	expTemplate.NotBefore = now.Add(-2 * notary.Year)
	expTemplate.NotAfter = now.Add(-notary.Year)
	expiredIntermediate, err := x509.CreateCertificate(rand.Reader, expTemplate, rootCert,
		sampleCertChain(t).rootKey.CryptoSigner().Public(), sampleCertChain(t).rootKey.CryptoSigner())
	require.NoError(t, err)

	ecdsax509Key := data.NewECDSAx509PublicKey(append(sampleCertChain(t).leafCert, expiredIntermediate...))
	require.NoError(t, cs.AddKey(data.CanonicalRootRole, "docker.io/notary/leaf", sampleCertChain(t).leafKey))

	otherKey, err := cs.Create(data.CanonicalTargetsRole, "docker.io/notary/leaf", data.ED25519Key)
	require.NoError(t, err)

	root := data.SignedRoot{
		Signatures: make([]data.Signature, 0),
		Signed: data.Root{
			SignedCommon: data.SignedCommon{
				Type:    "Root",
				Expires: now.Add(time.Hour),
				Version: 1,
			},
			Keys: map[string]data.PublicKey{
				ecdsax509Key.ID(): ecdsax509Key,
				otherKey.ID():     otherKey,
			},
			Roles: map[data.RoleName]*data.RootRole{
				"root": {
					KeyIDs:    []string{ecdsax509Key.ID()},
					Threshold: 1,
				},
				"targets": {
					KeyIDs:    []string{otherKey.ID()},
					Threshold: 1,
				},
				"snapshot": {
					KeyIDs:    []string{otherKey.ID()},
					Threshold: 1,
				},
				"timestamp": {
					KeyIDs:    []string{otherKey.ID()},
					Threshold: 1,
				},
			},
		},
		Dirty: true,
	}

	signedRoot, err := root.ToSigned()
	require.NoError(t, err)
	err = signed.Sign(cs, signedRoot, []data.PublicKey{ecdsax509Key}, 1, nil)
	require.NoError(t, err)

	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	defer os.RemoveAll(tempBaseDir)
	require.NoError(t, err, "failed to create a temporary directory: %s", err)

	_, err = trustpinning.ValidateRoot(
		nil,
		signedRoot,
		"docker.io/notary/leaf",
		trustpinning.TrustPinConfig{},
	)
	require.Error(t, err, "failed to invalidate expired intermediate certificate")
}

func TestCheckingWildcardCert(t *testing.T) {
	memStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)
	cs := cryptoservice.NewCryptoService(memStore)
	testPubKey, err := cs.Create(data.CanonicalRootRole, "docker.io/notary/*", data.ECDSAKey)
	require.NoError(t, err)
	testPrivKey, _, err := memStore.GetKey(testPubKey.ID())
	require.NoError(t, err)

	testCert, err := generateTestingCertificate(testPrivKey, "docker.io/notary/*", notary.Year)
	require.NoError(t, err)
	testCertPubKey, err := utils.ParsePEMPublicKey(utils.CertToPEM(testCert))
	require.NoError(t, err)

	rootRole, err := data.NewRole(data.CanonicalRootRole, 1, []string{testCertPubKey.ID()}, nil)
	require.NoError(t, err)
	testRoot, err := data.NewRoot(
		map[string]data.PublicKey{testCertPubKey.ID(): testCertPubKey},
		map[data.RoleName]*data.RootRole{
			data.CanonicalRootRole:      &rootRole.RootRole,
			data.CanonicalTimestampRole: &rootRole.RootRole,
			data.CanonicalTargetsRole:   &rootRole.RootRole,
			data.CanonicalSnapshotRole:  &rootRole.RootRole},
		false,
	)
	testRoot.Signed.Version = 1
	require.NoError(t, err, "Failed to create new root")

	signedTestRoot, err := testRoot.ToSigned()
	require.NoError(t, err)

	err = signed.Sign(cs, signedTestRoot, []data.PublicKey{testCertPubKey}, 1, nil)
	require.NoError(t, err)

	_, err = trustpinning.ValidateRoot(
		nil,
		signedTestRoot,
		"docker.io/notary/test",
		trustpinning.TrustPinConfig{},
	)
	require.NoError(t, err, "expected wildcard cert to validate")

	_, err = trustpinning.ValidateRoot(
		nil,
		signedTestRoot,
		"docker.io/not-a-match",
		trustpinning.TrustPinConfig{},
	)
	require.Error(t, err, "expected wildcard cert not to validate")
}

func TestWildcardMatching(t *testing.T) {
	var wildcardTests = []struct {
		CN  string
		gun string
		out bool
	}{
		{"docker.com/*", "docker.com/notary", true},
		{"docker.com/**", "docker.com/notary", true},
		{"*", "docker.com/any", true},
		{"*", "", true},
		{"**", "docker.com/any", true},
		{"test/*******", "test/many/wildcard", true},
		{"test/**/*/", "test/test", false},
		{"test/*/wild", "test/test/wild", false},
		{"*/all", "test/all", false},
		{"docker.com/*/*", "docker.com/notary/test", false},
		{"docker.com/*/**", "docker.com/notary/test", false},
		{"", "*", false},
		{"*abc*", "abc", false},
		{"test/*/wild*", "test/test/wild", false},
	}
	for _, tt := range wildcardTests {
		require.Equal(t, trustpinning.MatchCNToGun(tt.CN, data.GUN(tt.gun)), tt.out)
	}
}
