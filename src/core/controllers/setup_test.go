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

package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidSetupPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{"empty password", "", false},
		{"too short", "Abc1", false},
		{"exactly 7 chars", "Abcde1!", false},
		{"exactly 8 chars valid", "Abcdefg1", true},
		{"no uppercase", "abcdefg1", false},
		{"no lowercase", "ABCDEFG1", false},
		{"no number", "Abcdefgh", false},
		{"valid mixed", "Harbor12345", true},
		{"valid complex", "MyP@ssw0rd!", true},
		{"128 chars valid", string(makePassword(128)), true},
		{"129 chars too long", string(makePassword(129)), false},
		{"only numbers", "12345678", false},
		{"only letters", "Abcdefgh", false},
		{"valid minimum requirements", "aB345678", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validSetupPassword(tt.password)
			assert.Equal(t, tt.expected, result, "password: %q", tt.password)
		})
	}
}

// makePassword creates a password of specified length that meets requirements
func makePassword(length int) []byte {
	if length < 3 {
		return []byte{}
	}
	pw := make([]byte, length)
	pw[0] = 'A' // uppercase
	pw[1] = 'a' // lowercase
	pw[2] = '1' // digit
	for i := 3; i < length; i++ {
		pw[i] = 'x'
	}
	return pw
}
