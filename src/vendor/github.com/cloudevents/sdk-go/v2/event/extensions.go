/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// DataContentEncodingKey is the key to DeprecatedDataContentEncoding for versions that do not support data content encoding
	// directly.
	DataContentEncodingKey = "datacontentencoding"
)

var (
	// This determines the behavior of validateExtensionName(). For MaxExtensionNameLength > 0, an error will be returned,
	// if len(key) > MaxExtensionNameLength
	MaxExtensionNameLength = 0
)

func caseInsensitiveSearch(key string, space map[string]interface{}) (interface{}, bool) {
	lkey := strings.ToLower(key)
	for k, v := range space {
		if strings.EqualFold(lkey, strings.ToLower(k)) {
			return v, true
		}
	}
	return nil, false
}

func IsExtensionNameValid(key string) bool {
	if err := validateExtensionName(key); err != nil {
		return false
	}
	return true
}

func validateExtensionName(key string) error {
	if len(key) < 1 {
		return errors.New("bad key, CloudEvents attribute names MUST NOT be empty")
	}
	if MaxExtensionNameLength > 0 && len(key) > MaxExtensionNameLength {
		return fmt.Errorf("bad key, CloudEvents attribute name '%s' is longer than %d characters", key, MaxExtensionNameLength)
	}

	for _, c := range key {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return errors.New("bad key, CloudEvents attribute names MUST consist of lower-case letters ('a' to 'z') or digits ('0' to '9') from the ASCII character set")
		}
	}
	return nil
}
