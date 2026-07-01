package token

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"

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

func TestNewOptionsMethodDeprecated(t *testing.T) {
	// Method parameter is deprecated but still accepted for backward compatibility.
	// The method is determined from the key type regardless of the parameter value.
	keyFile := writeECKeyFile(t, elliptic.P256(), "EC PRIVATE KEY")
	defer os.Remove(keyFile)

	opt, err := NewOptions("RS256", "test-issuer", keyFile)
	require.NoError(t, err)
	// Despite passing "RS256", the signing method is determined from the ECDSA key
	assert.Equal(t, jwt.SigningMethodES256, opt.SignMethod)
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
	require.NoError(t, f.Close())
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
		{"P-521 PKCS8", elliptic.P521(), "PRIVATE KEY", "ES512"},
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
	t.Run("private key only (SEC1)", func(t *testing.T) {
		keyFile := writeECKeyFile(t, elliptic.P256(), "EC PRIVATE KEY")
		defer os.Remove(keyFile)

		opt, err := NewOptions("", "test-issuer", keyFile)
		require.NoError(t, err)

		key, err := opt.GetKey()
		require.NoError(t, err)
		assert.IsType(t, (*ecdsa.PrivateKey)(nil), key)
	})

	t.Run("private key only (PKCS8)", func(t *testing.T) {
		keyFile := writeECKeyFile(t, elliptic.P256(), "PRIVATE KEY")
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
		assert.Contains(t, err.Error(), "public key and private key do not match")
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

func TestNewAndRawWithECDSA(t *testing.T) {
	cases := []struct {
		name    string
		curve   elliptic.Curve
		pemType string
		wantAlg string
	}{
		{"P-256 SEC1", elliptic.P256(), "EC PRIVATE KEY", "ES256"},
		{"P-384 SEC1", elliptic.P384(), "EC PRIVATE KEY", "ES384"},
		{"P-521 SEC1", elliptic.P521(), "EC PRIVATE KEY", "ES512"},
		{"P-256 PKCS8", elliptic.P256(), "PRIVATE KEY", "ES256"},
		{"P-384 PKCS8", elliptic.P384(), "PRIVATE KEY", "ES384"},
		{"P-521 PKCS8", elliptic.P521(), "PRIVATE KEY", "ES512"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			keyFile := writeECKeyFile(t, tc.curve, tc.pemType)
			defer os.Remove(keyFile)

			opt, err := NewOptions("", "test-issuer", keyFile)
			require.NoError(t, err)

			claims := &jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				Issuer:    "test-issuer",
			}

			token, err := New(opt, claims)
			require.NoError(t, err)

			tokenStr, err := token.Raw()
			require.NoError(t, err)
			require.NotEmpty(t, tokenStr)

			parsedToken, err := Parse(opt, tokenStr, &jwt.RegisteredClaims{})
			require.NoError(t, err)
			require.NotNil(t, parsedToken)
			assert.Equal(t, tc.wantAlg, parsedToken.Header["alg"])
		})
	}
}

func TestGetKeyRSA(t *testing.T) {
	genRSAPEM := func(t *testing.T) (priv, pub []byte) {
		t.Helper()
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)
		priv = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
		pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		require.NoError(t, err)
		pub = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
		return
	}

	t.Run("no keys provided", func(t *testing.T) {
		opt := &Options{SignMethod: jwt.SigningMethodRS256}
		key, err := opt.GetKey()
		assert.Error(t, err)
		assert.Nil(t, key)
		assert.Contains(t, err.Error(), "no key provided")
	})

	t.Run("public key only", func(t *testing.T) {
		_, pubPEM := genRSAPEM(t)
		opt := &Options{SignMethod: jwt.SigningMethodRS256, PublicKey: pubPEM}
		key, err := opt.GetKey()
		require.NoError(t, err)
		assert.IsType(t, (*rsa.PublicKey)(nil), key)
	})

	t.Run("mismatched public private keys", func(t *testing.T) {
		priv1, _ := genRSAPEM(t)
		_, pub2 := genRSAPEM(t)
		opt := &Options{
			SignMethod: jwt.SigningMethodRS256,
			PrivateKey: priv1,
			PublicKey:  pub2,
		}
		key, err := opt.GetKey()
		assert.Error(t, err)
		assert.Nil(t, key)
		assert.Contains(t, err.Error(), "public key and private key do not match")
	})

	t.Run("matching public private keys", func(t *testing.T) {
		privPEM, pubPEM := genRSAPEM(t)
		opt := &Options{
			SignMethod: jwt.SigningMethodRS256,
			PrivateKey: privPEM,
			PublicKey:  pubPEM,
		}
		result, err := opt.GetKey()
		require.NoError(t, err)
		assert.IsType(t, (*rsa.PrivateKey)(nil), result)
	})

	t.Run("corrupted public key", func(t *testing.T) {
		privPEM, _ := genRSAPEM(t)
		corruptPubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: []byte("not-valid-der")})
		opt := &Options{
			SignMethod: jwt.SigningMethodRS256,
			PrivateKey: privPEM,
			PublicKey:  corruptPubPEM,
		}
		key, err := opt.GetKey()
		assert.Error(t, err)
		assert.Nil(t, key)
	})
}

func TestNewOptionsErrors(t *testing.T) {
	t.Run("file not found", func(t *testing.T) {
		_, err := NewOptions("", "issuer", "/nonexistent/path/to/key.pem")
		assert.Error(t, err)
	})

	t.Run("empty file", func(t *testing.T) {
		f, err := os.CreateTemp("", "harbor-empty-*.pem")
		require.NoError(t, err)
		f.Close()
		defer os.Remove(f.Name())
		_, err = NewOptions("", "issuer", f.Name())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode PEM")
	})

	t.Run("unsupported PEM type only", func(t *testing.T) {
		f, err := os.CreateTemp("", "harbor-cert-*.pem")
		require.NoError(t, err)
		require.NoError(t, pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: []byte("dummy")}))
		f.Close()
		defer os.Remove(f.Name())
		_, err = NewOptions("", "issuer", f.Name())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported private key type")
	})

	t.Run("corrupted RSA PKCS1 content", func(t *testing.T) {
		f, err := os.CreateTemp("", "harbor-corrupt-rsa-*.pem")
		require.NoError(t, err)
		require.NoError(t, pem.Encode(f, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: []byte("not-valid-der")}))
		f.Close()
		defer os.Remove(f.Name())
		_, err = NewOptions("", "issuer", f.Name())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse private key")
	})

	t.Run("corrupted EC PRIVATE KEY content", func(t *testing.T) {
		f, err := os.CreateTemp("", "harbor-corrupt-ec-*.pem")
		require.NoError(t, err)
		require.NoError(t, pem.Encode(f, &pem.Block{Type: "EC PRIVATE KEY", Bytes: []byte("not-valid-der")}))
		f.Close()
		defer os.Remove(f.Name())
		_, err = NewOptions("", "issuer", f.Name())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse private key")
	})

	t.Run("corrupted PKCS8 content", func(t *testing.T) {
		f, err := os.CreateTemp("", "harbor-corrupt-pkcs8-*.pem")
		require.NoError(t, err)
		require.NoError(t, pem.Encode(f, &pem.Block{Type: "PRIVATE KEY", Bytes: []byte("not-valid-der")}))
		f.Close()
		defer os.Remove(f.Name())
		_, err = NewOptions("", "issuer", f.Name())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse private key")
	})
}

func TestNewOptionsMultiBlockPEM(t *testing.T) {
	// OpenSSL sometimes generates EC private keys with a leading EC PARAMETERS
	// block. NewOptions should skip it and successfully load the key that follows.
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	f, err := os.CreateTemp("", "harbor-ec-params-*.pem")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	// P-256 curve OID as a minimal EC PARAMETERS block.
	p256OID := []byte{0x06, 0x08, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x03, 0x01, 0x07}
	require.NoError(t, pem.Encode(f, &pem.Block{Type: "EC PARAMETERS", Bytes: p256OID}))

	der, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)
	require.NoError(t, pem.Encode(f, &pem.Block{Type: "EC PRIVATE KEY", Bytes: der}))
	f.Close()

	opt, err := NewOptions("", "test-issuer", f.Name())
	require.NoError(t, err)
	assert.Equal(t, jwt.SigningMethodES256, opt.SignMethod)
}

func TestNewOptionsRSAPKCS8(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)

	f, err := os.CreateTemp("", "harbor-rsa-pkcs8-*.pem")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	require.NoError(t, pem.Encode(f, &pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	f.Close()

	opt, err := NewOptions("", "test-issuer", f.Name())
	require.NoError(t, err)
	assert.Equal(t, jwt.SigningMethodRS256, opt.SignMethod)

	// Verify the key can actually be parsed via GetKey()
	gotKey, err := opt.GetKey()
	require.NoError(t, err)
	assert.IsType(t, &rsa.PrivateKey{}, gotKey)
	privKey := gotKey.(*rsa.PrivateKey)
	assert.Equal(t, key.PublicKey.E, privKey.PublicKey.E)
}

func TestGetKeyErrors(t *testing.T) {
	t.Run("RSA private key parse error", func(t *testing.T) {
		opt := &Options{
			SignMethod: jwt.SigningMethodRS256,
			PrivateKey: pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: []byte("invalid-der")}),
		}
		key, err := opt.GetKey()
		assert.Error(t, err)
		assert.Nil(t, key)
	})

	t.Run("ECDSA private key parse error", func(t *testing.T) {
		opt := &Options{
			SignMethod: jwt.SigningMethodES256,
			PrivateKey: pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: []byte("invalid-der")}),
		}
		key, err := opt.GetKey()
		assert.Error(t, err)
		assert.Nil(t, key)
	})

	t.Run("ECDSA public key parse error", func(t *testing.T) {
		ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)
		privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: mustMarshalECPrivateKey(t, ecKey)})
		opt := &Options{
			SignMethod: jwt.SigningMethodES256,
			PrivateKey: privPEM,
			PublicKey:  pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: []byte("invalid-der")}),
		}
		k, err := opt.GetKey()
		assert.Error(t, err)
		assert.Nil(t, k)
	})

	t.Run("unsupported sign method", func(t *testing.T) {
		opt := &Options{SignMethod: jwt.SigningMethodHS256}
		key, err := opt.GetKey()
		assert.Error(t, err)
		assert.Nil(t, key)
		assert.Contains(t, err.Error(), "unsupported sign method")
	})
}

func mustMarshalECPrivateKey(t *testing.T, key *ecdsa.PrivateKey) []byte {
	t.Helper()
	der, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)
	return der
}

func TestNewOptionsUnsupportedCurve(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	require.NoError(t, err)

	der, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	f, err := os.CreateTemp("", "harbor-p224-*.pem")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	require.NoError(t, pem.Encode(f, &pem.Block{Type: "EC PRIVATE KEY", Bytes: der}))
	f.Close()

	_, err = NewOptions("", "test-issuer", f.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported ECDSA curve")
}

func TestNewOptionsEd25519(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	der, err := x509.MarshalPKCS8PrivateKey(priv)
	require.NoError(t, err)

	f, err := os.CreateTemp("", "harbor-ed25519-*.pem")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	require.NoError(t, pem.Encode(f, &pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	f.Close()

	_, err = NewOptions("", "test-issuer", f.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported private key type")
}
