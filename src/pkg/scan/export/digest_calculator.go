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

package export

import (
	"crypto/sha256"
	"io"
	"os"

	"github.com/opencontainers/go-digest"
)

// ArtifactDigestCalculator is an interface to be implemented by all file hash calculators
type ArtifactDigestCalculator interface {
	// Calculate returns the hash for a file
	Calculate(fileName string) (digest.Digest, error)
}

type SHA256ArtifactDigestCalculator struct{}

func (calc *SHA256ArtifactDigestCalculator) Calculate(fileName string) (digest.Digest, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return digest.NewDigest(digest.SHA256, hash), nil
}
