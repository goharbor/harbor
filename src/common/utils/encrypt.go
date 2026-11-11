// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha1" // nolint:gosec // G505: blocklisted import kept for legacy PBKDF2-SHA1 password verification only
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"io"
	"strings"
)

const (
	// EncryptHeaderV1 ...
	EncryptHeaderV1 = "<enc-v1>"
	// SHA1 is the name of sha1 hash alg
	SHA1 = "sha1"
	// SHA256 is the name of sha256 hash alg
	SHA256 = "sha256"
	// PBKDF2SHA256 is the password hashing version used for all newly created
	// or updated user passwords. It uses PBKDF2-HMAC-SHA256 with a high
	// iteration count (pbkdf2Iterations) following current OWASP guidance.
	PBKDF2SHA256 = "pbkdf2_sha256"
)

const (
	// legacyPBKDF2Iterations is the iteration count used by credentials hashed
	// with PasswordVersion SHA1 or SHA256. It is retained ONLY to verify
	// pre-existing credentials and must not be used for new secrets.
	legacyPBKDF2Iterations = 4096
	// pbkdf2Iterations is the OWASP-recommended minimum iteration count for
	// PBKDF2-HMAC-SHA256, used for all newly created/updated user passwords
	// (PasswordVersion == PBKDF2SHA256).
	pbkdf2Iterations = 600000
)

// HashAlg maps a password version to the hash function we use for it.
//
// We only keep SHA1 around so that old users can still log in. Their passwords
// were hashed with PBKDF2-HMAC-SHA1 (PasswordVersion == SHA1) back in the day,
// so we need it to check those credentials. Anything new goes through
// PBKDF2-HMAC-SHA256 with a high iteration count instead (see
// pkg/user/manager.go). Please don't use SHA1 to hash any new secrets.
var HashAlg = map[string]func() hash.Hash{
	SHA1:         sha1.New, // nolint:gosec // G401/G505: weak hash kept only for legacy PBKDF2 password verification
	SHA256:       sha256.New,
	PBKDF2SHA256: sha256.New,
}

// pbkdf2Params picks the right hash function and PBKDF2 iteration count for a
// given credential version. New passwords get the strong iteration count, while
// the older SHA1/SHA256 ones stick with their original count so we can still
// verify existing credentials (both user passwords and robot secrets).
func pbkdf2Params(version string) (func() hash.Hash, int) {
	alg, ok := HashAlg[version]
	if !ok {
		// Unknown/new version: use the strongest scheme, i.e. SHA-256 with the
		// high iteration count, so a bad or typo'd version never falls back to a
		// low work factor.
		return sha256.New, pbkdf2Iterations
	}
	iterations := legacyPBKDF2Iterations
	if version == PBKDF2SHA256 {
		iterations = pbkdf2Iterations
	}
	return alg, iterations
}

// Encrypt encrypts the content with salt
func Encrypt(content string, salt string, encryptAlg string) string {
	alg, iterations := pbkdf2Params(encryptAlg)
	key, _ := pbkdf2.Key(alg, content, []byte(salt), iterations, 16)
	return fmt.Sprintf("%x", key)
}

// ReversibleEncrypt encrypts the str with aes/base64
func ReversibleEncrypt(str, key string) (string, error) {
	keyBytes := []byte(key)
	var block cipher.Block
	var err error

	if block, err = aes.NewCipher(keyBytes); err != nil {
		return "", err
	}

	// ensures the value is no larger than 64 MB, which fits comfortably within an int and avoids the overflow
	if len(str) > 64*1024*1024 {
		return "", errors.New("str value too large")
	}

	size := aes.BlockSize + len(str)
	cipherText := make([]byte, size)
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(cipherText[aes.BlockSize:], []byte(str))
	encrypted := EncryptHeaderV1 + base64.StdEncoding.EncodeToString(cipherText)
	return encrypted, nil
}

// ReversibleDecrypt decrypts the str with aes/base64 or base 64 depending on "header"
func ReversibleDecrypt(str, key string) (string, error) {
	if strings.HasPrefix(str, EncryptHeaderV1) {
		str = str[len(EncryptHeaderV1):]
		return decryptAES(str, key)
	}
	// fallback to base64
	return decodeB64(str)
}

func decodeB64(str string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(str)
	return string(cipherText), err
}

func decryptAES(str, key string) (string, error) {
	keyBytes := []byte(key)
	var block cipher.Block
	var cipherText []byte
	var err error

	if block, err = aes.NewCipher(keyBytes); err != nil {
		return "", err
	}
	if cipherText, err = base64.StdEncoding.DecodeString(str); err != nil {
		return "", err
	}
	if len(cipherText) < aes.BlockSize {
		err = errors.New("cipherText too short")
		return "", err
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(cipherText, cipherText)
	return string(cipherText), nil
}
