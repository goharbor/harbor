// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

// Encrypt encrypts the content with salt
func Encrypt(content string, salt string) string {
	return fmt.Sprintf("%x", pbkdf2.Key([]byte(content), []byte(salt), 4096, 16, sha1.New))
}

const (
	// EncryptHeaderV1 ...
	EncryptHeaderV1 = "<enc-v1>"
)

// ReversibleEncrypt encrypts the str with aes/base64
func ReversibleEncrypt(str, key string) (string, error) {
	keyBytes := []byte(key)
	var block cipher.Block
	var err error

	if block, err = aes.NewCipher(keyBytes); err != nil {
		return "", err
	}
	cipherText := make([]byte, aes.BlockSize+len(str))
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
	//fallback to base64
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
