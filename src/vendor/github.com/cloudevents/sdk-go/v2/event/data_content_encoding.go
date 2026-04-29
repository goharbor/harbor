/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event

const (
	Base64 = "base64"
)

// StringOfBase64 returns a string pointer to "Base64"
func StringOfBase64() *string {
	a := Base64
	return &a
}
