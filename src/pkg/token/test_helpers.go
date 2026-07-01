package token

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
)

// writeECKeyFile generates an ECDSA key with the given curve and writes it to
// a temporary file. It returns the file path.
func writeECKeyFile(t *testing.T, curve elliptic.Curve, pemType string) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ECDSA key: %v", err)
	}

	var der []byte
	if pemType == "EC PRIVATE KEY" {
		der, err = x509.MarshalECPrivateKey(key)
	} else {
		der, err = x509.MarshalPKCS8PrivateKey(key)
	}
	if err != nil {
		t.Fatalf("failed to marshal ECDSA key: %v", err)
	}

	f, err := os.CreateTemp("", "harbor-ec-key-*.pem")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if err := pem.Encode(f, &pem.Block{Type: pemType, Bytes: der}); err != nil {
		f.Close()
		os.Remove(f.Name())
		t.Fatalf("failed to encode PEM: %v", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		t.Fatalf("failed to close file: %v", err)
	}
	return f.Name()
}
