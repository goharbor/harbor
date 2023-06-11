//
// Copyright 2022 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tuf

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sigstore/sigstore/pkg/cryptoutils"
	"github.com/sigstore/sigstore/pkg/signature"
	"github.com/sigstore/sigstore/pkg/signature/options"
	"github.com/theupdateframework/go-tuf"
)

type TestSigstoreRoot struct {
	Rekor             signature.Verifier
	FulcioCertificate *x509.Certificate
	// TODO: Include a CTFE key if/when cosign verifies SCT.
}

// This creates a new sigstore TUF repo whose signers can be used to create dynamic
// signed Rekor entries.
func NewSigstoreTufRepo(t *testing.T, root TestSigstoreRoot) (tuf.LocalStore, *tuf.Repo) {
	td := t.TempDir()
	ctx := context.Background()
	remote := tuf.FileSystemStore(td, nil)
	r, err := tuf.NewRepo(remote)
	if err != nil {
		t.Error(err)
	}
	if err := r.Init(false); err != nil {
		t.Error(err)
	}

	for _, role := range []string{"root", "targets", "snapshot", "timestamp"} {
		if _, err := r.GenKey(role); err != nil {
			t.Error(err)
		}
	}
	targetsPath := filepath.Join(td, "staged", "targets")
	if err := os.MkdirAll(filepath.Dir(targetsPath), 0o755); err != nil {
		t.Error(err)
	}
	// Add the rekor key target
	pk, err := root.Rekor.PublicKey(options.WithContext(ctx))
	if err != nil {
		t.Error(err)
	}
	b, err := x509.MarshalPKIXPublicKey(pk)
	if err != nil {
		t.Error(err)
	}
	rekorPath := "rekor.pub"
	rekorData := cryptoutils.PEMEncode(cryptoutils.PublicKeyPEMType, b)
	if err := os.WriteFile(filepath.Join(targetsPath, rekorPath), rekorData, 0o600); err != nil {
		t.Error(err)
	}
	scmRekor, err := json.Marshal(&sigstoreCustomMetadata{Sigstore: customMetadata{Usage: Rekor, Status: Active}})
	if err != nil {
		t.Error(err)
	}
	if err := r.AddTarget("rekor.pub", scmRekor); err != nil {
		t.Error(err)
	}
	// Add Fulcio Certificate information.
	fulcioPath := "fulcio.crt.pem"
	fulcioData := cryptoutils.PEMEncode(cryptoutils.CertificatePEMType, root.FulcioCertificate.Raw)
	if err := os.WriteFile(filepath.Join(targetsPath, fulcioPath), fulcioData, 0o600); err != nil {
		t.Error(err)
	}
	scmFulcio, err := json.Marshal(&sigstoreCustomMetadata{Sigstore: customMetadata{Usage: Fulcio, Status: Active}})
	if err != nil {
		t.Error(err)
	}
	if err := r.AddTarget(fulcioPath, scmFulcio); err != nil {
		t.Error(err)
	}
	if err := r.Snapshot(); err != nil {
		t.Error(err)
	}
	if err := r.Timestamp(); err != nil {
		t.Error(err)
	}
	if err := r.Commit(); err != nil {
		t.Error(err)
	}
	// Serve remote repository.
	s := httptest.NewServer(http.FileServer(http.Dir(filepath.Join(td, "repository"))))
	defer s.Close()

	// Initialize with custom root.
	tufRoot := t.TempDir()
	t.Setenv("TUF_ROOT", tufRoot)
	meta, err := remote.GetMeta()
	if err != nil {
		t.Error(err)
	}
	rootBytes, ok := meta["root.json"]
	if !ok {
		t.Error(err)
	}
	resetForTests()
	if err := Initialize(ctx, s.URL, rootBytes); err != nil {
		t.Error(err)
	}
	t.Cleanup(func() {
		resetForTests()
	})
	return remote, r
}
