// Package utils contains tuf related utility functions however this file is hard
// forked from https://github.com/youmark/pkcs8 package. It has been further modified
// based on the requirements of Notary. For converting keys into PKCS#8 format,
// original package expected *crypto.PrivateKey interface, which then type inferred
// to either *rsa.PrivateKey or *ecdsa.PrivateKey depending on the need and later
// converted to ASN.1 DER encoded form, this whole process was superfluous here as
// keys are already being kept in ASN.1 DER format wrapped in data.PrivateKey
// structure. With these changes, package has became tightly coupled with notary as
// most of the method signatures have been updated. Moreover support for ED25519
// keys has been added as well. License for original package is following:
//
// The MIT License (MIT)
//
// Copyright (c) 2014 youmark
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"fmt"

	"golang.org/x/crypto/pbkdf2"

	"github.com/theupdateframework/notary/tuf/data"
)

// Copy from crypto/x509
var (
	oidPublicKeyRSA   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}
	oidPublicKeyDSA   = asn1.ObjectIdentifier{1, 2, 840, 10040, 4, 1}
	oidPublicKeyECDSA = asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}
	// crypto/x509 doesn't have support for ED25519
	// http://www.oid-info.com/get/1.3.6.1.4.1.11591.15.1
	oidPublicKeyED25519 = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 11591, 15, 1}
)

// Copy from crypto/x509
var (
	oidNamedCurveP224 = asn1.ObjectIdentifier{1, 3, 132, 0, 33}
	oidNamedCurveP256 = asn1.ObjectIdentifier{1, 2, 840, 10045, 3, 1, 7}
	oidNamedCurveP384 = asn1.ObjectIdentifier{1, 3, 132, 0, 34}
	oidNamedCurveP521 = asn1.ObjectIdentifier{1, 3, 132, 0, 35}
)

// Copy from crypto/x509
func oidFromNamedCurve(curve elliptic.Curve) (asn1.ObjectIdentifier, bool) {
	switch curve {
	case elliptic.P224():
		return oidNamedCurveP224, true
	case elliptic.P256():
		return oidNamedCurveP256, true
	case elliptic.P384():
		return oidNamedCurveP384, true
	case elliptic.P521():
		return oidNamedCurveP521, true
	}

	return nil, false
}

// Unecrypted PKCS8
var (
	oidPKCS5PBKDF2 = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 5, 12}
	oidPBES2       = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 5, 13}
	oidAES256CBC   = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 1, 42}
)

type ecPrivateKey struct {
	Version       int
	PrivateKey    []byte
	NamedCurveOID asn1.ObjectIdentifier `asn1:"optional,explicit,tag:0"`
	PublicKey     asn1.BitString        `asn1:"optional,explicit,tag:1"`
}

type privateKeyInfo struct {
	Version             int
	PrivateKeyAlgorithm []asn1.ObjectIdentifier
	PrivateKey          []byte
}

// Encrypted PKCS8
type pbkdf2Params struct {
	Salt           []byte
	IterationCount int
}

type pbkdf2Algorithms struct {
	IDPBKDF2     asn1.ObjectIdentifier
	PBKDF2Params pbkdf2Params
}

type pbkdf2Encs struct {
	EncryAlgo asn1.ObjectIdentifier
	IV        []byte
}

type pbes2Params struct {
	KeyDerivationFunc pbkdf2Algorithms
	EncryptionScheme  pbkdf2Encs
}

type pbes2Algorithms struct {
	IDPBES2     asn1.ObjectIdentifier
	PBES2Params pbes2Params
}

type encryptedPrivateKeyInfo struct {
	EncryptionAlgorithm pbes2Algorithms
	EncryptedData       []byte
}

// pkcs8 reflects an ASN.1, PKCS#8 PrivateKey.
// copied from https://github.com/golang/go/blob/964639cc338db650ccadeafb7424bc8ebb2c0f6c/src/crypto/x509/pkcs8.go#L17
type pkcs8 struct {
	Version    int
	Algo       pkix.AlgorithmIdentifier
	PrivateKey []byte
}

func parsePKCS8ToTufKey(der []byte) (data.PrivateKey, error) {
	var key pkcs8

	if _, err := asn1.Unmarshal(der, &key); err != nil {
		if _, ok := err.(asn1.StructuralError); ok {
			return nil, errors.New("could not decrypt private key")
		}
		return nil, err
	}

	if key.Algo.Algorithm.Equal(oidPublicKeyED25519) {
		tufED25519PrivateKey, err := ED25519ToPrivateKey(key.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("could not convert ed25519.PrivateKey to data.PrivateKey: %v", err)
		}

		return tufED25519PrivateKey, nil
	}

	privKey, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, err
	}

	switch priv := privKey.(type) {
	case *rsa.PrivateKey:
		tufRSAPrivateKey, err := RSAToPrivateKey(priv)
		if err != nil {
			return nil, fmt.Errorf("could not convert rsa.PrivateKey to data.PrivateKey: %v", err)
		}

		return tufRSAPrivateKey, nil
	case *ecdsa.PrivateKey:
		tufECDSAPrivateKey, err := ECDSAToPrivateKey(priv)
		if err != nil {
			return nil, fmt.Errorf("could not convert ecdsa.PrivateKey to data.PrivateKey: %v", err)
		}

		return tufECDSAPrivateKey, nil
	}

	return nil, errors.New("unsupported key type")
}

// ParsePKCS8ToTufKey requires PKCS#8 key in DER format and returns data.PrivateKey
// Password should be provided in case of Encrypted PKCS#8 key, else it should be nil.
func ParsePKCS8ToTufKey(der []byte, password []byte) (data.PrivateKey, error) {
	if password == nil {
		return parsePKCS8ToTufKey(der)
	}

	var privKey encryptedPrivateKeyInfo
	if _, err := asn1.Unmarshal(der, &privKey); err != nil {
		return nil, errors.New("pkcs8: only PKCS #5 v2.0 supported")
	}

	if !privKey.EncryptionAlgorithm.IDPBES2.Equal(oidPBES2) {
		return nil, errors.New("pkcs8: only PBES2 supported")
	}

	if !privKey.EncryptionAlgorithm.PBES2Params.KeyDerivationFunc.IDPBKDF2.Equal(oidPKCS5PBKDF2) {
		return nil, errors.New("pkcs8: only PBKDF2 supported")
	}

	encParam := privKey.EncryptionAlgorithm.PBES2Params.EncryptionScheme
	kdfParam := privKey.EncryptionAlgorithm.PBES2Params.KeyDerivationFunc.PBKDF2Params

	switch {
	case encParam.EncryAlgo.Equal(oidAES256CBC):
		iv := encParam.IV
		salt := kdfParam.Salt
		iter := kdfParam.IterationCount

		encryptedKey := privKey.EncryptedData
		symkey := pbkdf2.Key(password, salt, iter, 32, sha1.New)
		block, err := aes.NewCipher(symkey)
		if err != nil {
			return nil, err
		}
		mode := cipher.NewCBCDecrypter(block, iv)
		mode.CryptBlocks(encryptedKey, encryptedKey)

		// no need to explicitly remove padding, as ASN.1 unmarshalling will automatically discard it
		key, err := parsePKCS8ToTufKey(encryptedKey)
		if err != nil {
			return nil, errors.New("pkcs8: incorrect password")
		}

		return key, nil
	default:
		return nil, errors.New("pkcs8: only AES-256-CBC supported")
	}

}

func convertTUFKeyToPKCS8(priv data.PrivateKey) ([]byte, error) {
	var pkey privateKeyInfo

	switch priv.Algorithm() {
	case data.RSAKey, data.RSAx509Key:
		// Per RFC5958, if publicKey is present, then version is set to v2(1) else version is set to v1(0).
		// But openssl set to v1 even publicKey is present
		pkey.Version = 0
		pkey.PrivateKeyAlgorithm = make([]asn1.ObjectIdentifier, 1)
		pkey.PrivateKeyAlgorithm[0] = oidPublicKeyRSA
		pkey.PrivateKey = priv.Private()
	case data.ECDSAKey, data.ECDSAx509Key:
		// To extract Curve value, parsing ECDSA key to *ecdsa.PrivateKey
		eckey, err := x509.ParseECPrivateKey(priv.Private())
		if err != nil {
			return nil, err
		}

		oidNamedCurve, ok := oidFromNamedCurve(eckey.Curve)
		if !ok {
			return nil, errors.New("pkcs8: unknown elliptic curve")
		}

		// Per RFC5958, if publicKey is present, then version is set to v2(1) else version is set to v1(0).
		// But openssl set to v1 even publicKey is present
		pkey.Version = 1
		pkey.PrivateKeyAlgorithm = make([]asn1.ObjectIdentifier, 2)
		pkey.PrivateKeyAlgorithm[0] = oidPublicKeyECDSA
		pkey.PrivateKeyAlgorithm[1] = oidNamedCurve
		pkey.PrivateKey = priv.Private()
	case data.ED25519Key:
		pkey.Version = 0
		pkey.PrivateKeyAlgorithm = make([]asn1.ObjectIdentifier, 1)
		pkey.PrivateKeyAlgorithm[0] = oidPublicKeyED25519
		pkey.PrivateKey = priv.Private()
	default:
		return nil, fmt.Errorf("algorithm %s not supported", priv.Algorithm())
	}

	return asn1.Marshal(pkey)
}

func convertTUFKeyToPKCS8Encrypted(priv data.PrivateKey, password []byte) ([]byte, error) {
	// Convert private key into PKCS8 format
	pkey, err := convertTUFKeyToPKCS8(priv)
	if err != nil {
		return nil, err
	}

	// Calculate key from password based on PKCS5 algorithm
	// Use 8 byte salt, 16 byte IV, and 2048 iteration
	iter := 2048
	salt := make([]byte, 8)
	iv := make([]byte, 16)
	_, err = rand.Reader.Read(salt)
	if err != nil {
		return nil, err
	}

	_, err = rand.Reader.Read(iv)
	if err != nil {
		return nil, err
	}

	key := pbkdf2.Key(password, salt, iter, 32, sha1.New)

	// Use AES256-CBC mode, pad plaintext with PKCS5 padding scheme
	padding := aes.BlockSize - len(pkey)%aes.BlockSize
	if padding > 0 {
		n := len(pkey)
		pkey = append(pkey, make([]byte, padding)...)
		for i := 0; i < padding; i++ {
			pkey[n+i] = byte(padding)
		}
	}

	encryptedKey := make([]byte, len(pkey))
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encryptedKey, pkey)

	pbkdf2algo := pbkdf2Algorithms{oidPKCS5PBKDF2, pbkdf2Params{salt, iter}}
	pbkdf2encs := pbkdf2Encs{oidAES256CBC, iv}
	pbes2algo := pbes2Algorithms{oidPBES2, pbes2Params{pbkdf2algo, pbkdf2encs}}

	encryptedPkey := encryptedPrivateKeyInfo{pbes2algo, encryptedKey}
	return asn1.Marshal(encryptedPkey)
}

// ConvertTUFKeyToPKCS8 converts a private key (data.Private) to PKCS#8 and returns in DER format
// if password is not nil, it would convert the Private Key to Encrypted PKCS#8.
func ConvertTUFKeyToPKCS8(priv data.PrivateKey, password []byte) ([]byte, error) {
	if password == nil {
		return convertTUFKeyToPKCS8(priv)
	}
	return convertTUFKeyToPKCS8Encrypted(priv, password)
}
