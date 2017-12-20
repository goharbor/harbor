package cryptoservice

import (
	"crypto/rand"
	"crypto/x509"
	"testing"
	"time"

	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/utils"
	"github.com/stretchr/testify/require"
)

func TestGenerateCertificate(t *testing.T) {
	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err, "could not generate key")

	keyStore := trustmanager.NewKeyMemoryStore(passphraseRetriever)

	err = keyStore.AddKey(trustmanager.KeyInfo{Role: data.CanonicalRootRole, Gun: ""}, privKey)
	require.NoError(t, err, "could not add key to store")

	// Check GenerateCertificate method
	var gun data.GUN = "docker.com/notary"
	startTime := time.Now()
	cert, err := GenerateCertificate(privKey, gun, startTime, startTime.AddDate(10, 0, 0))
	require.NoError(t, err, "could not generate certificate")

	// Check public key
	ecdsaPrivateKey, err := x509.ParseECPrivateKey(privKey.Private())
	require.NoError(t, err)
	ecdsaPublicKey := ecdsaPrivateKey.Public()
	require.Equal(t, ecdsaPublicKey, cert.PublicKey)

	// Check CommonName
	require.EqualValues(t, cert.Subject.CommonName, gun)
}
