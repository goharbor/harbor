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

package azurecr

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	registryclient "github.com/goharbor/harbor/src/pkg/registry"
)

// TestChunkedUploadToACR is an integration test that validates chunked blob
// upload to a real Azure Container Registry. It mirrors the exact flow that
// Harbor's replication uses when SupportedCopyByChunk is enabled.
//
// Run with environment variables:
//
//	ACR_URL=https://min01rg72arc.azurecr.io ACR_USER=admin ACR_PASS=xxx \
//	  go test -v -run TestChunkedUploadToACR ./pkg/reg/adapter/azurecr/...
func TestChunkedUploadToACR(t *testing.T) {
	acrURL := os.Getenv("ACR_URL")
	acrUser := os.Getenv("ACR_USER")
	acrPass := os.Getenv("ACR_PASS")

	if acrURL == "" || acrUser == "" || acrPass == "" {
		t.Skip("Skipping integration test: set ACR_URL, ACR_USER, ACR_PASS env vars")
	}

	repository := "test/chunked-upload-test"
	chunkSize := int64(5 * 1024 * 1024) // 5MB chunks

	// Generate a 25MB random blob (larger than the ~20MB monolithic limit)
	blobSize := int64(25 * 1024 * 1024)
	blobData := make([]byte, blobSize)
	_, err := rand.Read(blobData)
	require.NoError(t, err, "failed to generate random blob")

	digest := fmt.Sprintf("sha256:%x", sha256.Sum256(blobData))

	t.Logf("ACR URL: %s", acrURL)
	t.Logf("Repository: %s", repository)
	t.Logf("Blob size: %dMB, Chunk size: %dMB", blobSize/1024/1024, chunkSize/1024/1024)
	t.Logf("Digest: %s", digest)

	// Create registry using the same ACR authorizer that Harbor replication uses
	registry := &model.Registry{
		URL: acrURL,
		Credential: &model.Credential{
			AccessKey:    acrUser,
			AccessSecret: acrPass,
		},
		Insecure: false,
	}

	// Use NewClient which auto-detects auth scheme (Bearer for ACR)
	client := registryclient.NewClient(acrURL, acrUser, acrPass, false)

	// Also create a client via the adapter factory (end-to-end validation)
	adp := &adapter{
		Adapter:     native.NewAdapterWithAuthorizer(registry, newAuthorizer(registry)),
		registryURL: acrURL,
	}
	info, err := adp.Info()
	require.NoError(t, err)
	assert.True(t, info.SupportedCopyByChunk, "Azure ACR adapter must declare SupportedCopyByChunk")

	// --- Test 1: Monolithic upload should fail with 413 ---
	t.Run("monolithic_upload_fails_413", func(t *testing.T) {
		err := client.PushBlob(repository, digest, blobSize, io.NopCloser(bytes.NewReader(blobData)))
		if err != nil {
			t.Logf("Monolithic upload failed as expected: %v", err)
			assert.Contains(t, err.Error(), "413", "expected 413 error for monolithic upload")
		} else {
			t.Log("Monolithic upload succeeded unexpectedly (ACR may have increased limits)")
		}
	})

	// --- Test 2: Chunked upload should succeed ---
	t.Run("chunked_upload_succeeds", func(t *testing.T) {
		var location string
		var end int64 = -1
		endRange := blobSize - 1

		for {
			start := end + 1
			end = start + chunkSize - 1
			if end > endRange {
				end = endRange
			}

			chunk := blobData[start : end+1]
			t.Logf("Uploading chunk %d-%d (%d bytes, last=%v)",
				start, end, len(chunk), end == endRange)

			newLocation, newEnd, err := client.PushBlobChunk(
				repository, digest, blobSize,
				io.NopCloser(bytes.NewReader(chunk)),
				start, end, location,
			)
			require.NoError(t, err, "chunk upload %d-%d failed", start, end)

			location = newLocation
			end = newEnd

			t.Logf("  OK - new location length: %d", len(location))

			if end == endRange {
				break
			}
		}

		// Verify blob exists
		exist, err := client.BlobExist(repository, digest)
		require.NoError(t, err, "failed to check blob existence")
		assert.True(t, exist, "blob should exist after chunked upload")

		t.Log("=== SUCCESS: Chunked upload to Azure ACR works! ===")
	})
}
