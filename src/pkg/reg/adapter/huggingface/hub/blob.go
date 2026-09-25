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

package hub

import (
	"context"
	"crypto/sha1" // nolint:gosec // G505: git blob ids are sha1, used only to check downloads against the Hub listing
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"strconv"

	"github.com/opencontainers/go-digest"
)

// maxNonLFSSize caps the download of a non-LFS file whose sha256 has to be computed.
var maxNonLFSSize int64 = 64 << 20

// GitBlobID returns the git blob sha1 of content, as the Hub reports it in blobId.
func GitBlobID(content []byte) string {
	h := newGitBlobHash(int64(len(content)))
	_, _ = h.Write(content)
	return hex.EncodeToString(h.Sum(nil))
}

func newGitBlobHash(size int64) hash.Hash {
	h := sha1.New() // nolint:gosec // G401: git blob id, not a security boundary
	_, _ = io.WriteString(h, "blob "+strconv.FormatInt(size, 10)+"\x00")
	return h
}

// ContentDigest downloads a non-LFS file at commit, verifies it against the git blob id of the
// snapshot and returns its sha256 digest.
func (c *Client) ContentDigest(ctx context.Context, modelID, commit string, f File) (digest.Digest, error) {
	if f.IsLFS() {
		return digest.NewDigestFromEncoded(digest.SHA256, f.SHA256), nil
	}
	if f.Size > maxNonLFSSize {
		return "", fmt.Errorf("non-LFS file %s of model %s is %d bytes, above the %d byte limit", f.Path, modelID, f.Size, maxNonLFSSize)
	}
	body, err := c.Open(ctx, modelID, commit, f.Path, 0, -1)
	if err != nil {
		return "", err
	}
	defer body.Close()

	gitHash := newGitBlobHash(f.Size)
	digester := digest.SHA256.Digester()
	n, err := io.Copy(io.MultiWriter(gitHash, digester.Hash()), io.LimitReader(body, f.Size+1))
	if err != nil {
		return "", fmt.Errorf("failed to download %s of model %s: %v", f.Path, modelID, err)
	}
	if n != f.Size {
		return "", fmt.Errorf("downloaded %d bytes of %s of model %s, expected %d", n, f.Path, modelID, f.Size)
	}
	if got := hex.EncodeToString(gitHash.Sum(nil)); got != f.BlobID {
		return "", fmt.Errorf("git blob id of %s of model %s is %s, expected %s", f.Path, modelID, got, f.BlobID)
	}
	return digester.Digest(), nil
}
