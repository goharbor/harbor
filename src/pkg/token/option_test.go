package token

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOptions(t *testing.T) {
	defaultOpt := DefaultTokenOptions()
	if defaultOpt == nil {
		assert.NotNil(t, defaultOpt)
		return
	}
	assert.Equal(t, defaultOpt.SignMethod, jwt.GetSigningMethod("RS256"))
	assert.Equal(t, defaultOpt.Issuer, "harbor-token-defaultIssuer")
}

func TestGetKey(t *testing.T) {
	defaultOpt := DefaultTokenOptions()
	if defaultOpt == nil {
		assert.NotNil(t, defaultOpt)
		return
	}
	key, err := defaultOpt.GetKey()
	assert.Nil(t, err)
	assert.NotNil(t, key)
}

// writeECKeyFile generates an ECDSA key with the given curve and writes it to
// a temp PEM file in the specified format ("EC PRIVATE KEY" or "PRIVATE KEY").
// The caller is responsible for removing the file.
func writeECKeyFile(t *testing.T, curve elliptic.Curve, pemType string) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	require.NoError(t, err)

	var der []byte
	if pemType == "EC PRIVATE KEY" {
		der, err = x509.MarshalECPrivateKey(key)
	} else {
		der, err = x509.MarshalPKCS8PrivateKey(key)
	}
	require.NoError(t, err)

	f, err := os.CreateTemp("", "harbor-ec-key-*.pem")
	require.NoError(t, err)
	require.NoError(t, pem.Encode(f, &pem.Block{Type: pemType, Bytes: der}))
	f.Close()
	return f.Name()
}

func TestNewOptionsECDSA(t *testing.T) {
	cases := []struct {
		name       string
		curve      elliptic.Curve
		pemType    string
		wantMethod string
	}{
		{"P-256 SEC1", elliptic.P256(), "EC PRIVATE KEY", "ES256"},
		{"P-384 SEC1", elliptic.P384(), "EC PRIVATE KEY", "ES384"},
		{"P-521 SEC1", elliptic.P521(), "EC PRIVATE KEY", "ES512"},
		{"P-256 PKCS8", elliptic.P256(), "PRIVATE KEY", "ES256"},
		{"P-384 PKCS8", elliptic.P384(), "PRIVATE KEY", "ES384"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			keyFile := writeECKeyFile(t, tc.curve, tc.pemType)
			defer os.Remove(keyFile)

			opt, err := NewOptions("", "test-issuer", keyFile)
			require.NoError(t, err)
			assert.Equal(t, jwt.GetSigningMethod(tc.wantMethod), opt.SignMethod)
			assert.Equal(t, "test-issuer", opt.Issuer)
		})
	}
}

func TestGetKeyECDSA(t *testing.T) {
	t.Run("private key only", func(t *testing.T) {
		keyFile := writeECKeyFile(t, elliptic.P256(), "EC PRIVATE KEY")
		defer os.Remove(keyFile)

		opt, err := NewOptions("", "test-issuer", keyFile)
		require.NoError(t, err)

		key, err := opt.GetKey()
		require.NoError(t, err)
		assert.IsType(t, (*ecdsa.PrivateKey)(nil), key)
	})

	t.Run("no keys provided", func(t *testing.T) {
		opt := &Options{
			SignMethod: jwt.SigningMethodES256,
		}

		key, err := opt.GetKey()
		assert.Error(t, err)
		assert.Nil(t, key)
		assert.Contains(t, err.Error(), "no key provided")
	})

	t.Run("mismatched public private keys", func(t *testing.T) {
		// Generate two different ECDSA keys
		key1, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)
		key2, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		// Encode private key from key1
		privDER, err := x509.MarshalECPrivateKey(key1)
		require.NoError(t, err)
		privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privDER})

		// Encode public key from key2 (mismatch!)
		pubDER, err := x509.MarshalPKIXPublicKey(&key2.PublicKey)
		require.NoError(t, err)
		pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})

		opt := &Options{
			SignMethod: jwt.SigningMethodES256,
			PrivateKey: privPEM,
			PublicKey:  pubPEM,
		}

		key, err := opt.GetKey()
		assert.Error(t, err)
		assert.Nil(t, key)
		assert.Contains(t, err.Error(), "the public key and private key are not match")
	})

	t.Run("matching public private keys", func(t *testing.T) {
		// Generate one ECDSA key
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		// Encode private key
		privDER, err := x509.MarshalECPrivateKey(key)
		require.NoError(t, err)
		privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privDER})

		// Encode matching public key
		pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		require.NoError(t, err)
		pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})

		opt := &Options{
			SignMethod: jwt.SigningMethodES256,
			PrivateKey: privPEM,
			PublicKey:  pubPEM,
		}

		result, err := opt.GetKey()
		require.NoError(t, err)
		assert.IsType(t, (*ecdsa.PrivateKey)(nil), result)
		assert.Equal(t, key, result)
	})

	t.Run("public key only", func(t *testing.T) {
		// Generate one ECDSA key and extract public key
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		// Encode public key only
		pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		require.NoError(t, err)
		pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})

		opt := &Options{
			SignMethod: jwt.SigningMethodES256,
			PublicKey:  pubPEM,
		}

		result, err := opt.GetKey()
		require.NoError(t, err)
		assert.IsType(t, (*ecdsa.PublicKey)(nil), result)
		assert.Equal(t, &key.PublicKey, result)
	})
}
