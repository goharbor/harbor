//go:build go1.18
// +build go1.18

// Copyright Project Harbor Authors
// SPDX-License-Identifier: Apache-2.0

package harbor_test

import (
	"testing"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/pkg/distribution"
)

// FuzzValidateHTTPURL tests URL validation with arbitrary
// attacker-controlled URL strings.
//
// Harbor is a cloud-native container registry with 28K+ stars
// and 23 GitHub Security Advisories.
func FuzzValidateHTTPURL(f *testing.F) {
	f.Add("https://registry.example.com")
	f.Add("http://localhost")
	f.Add("")
	f.Add("not-a-url")
	f.Add(string(make([]byte, 1000)))

	f.Fuzz(func(t *testing.T, s string) {
		if len(s) > 1<<16 {
			return
		}
		_, _ = lib.ValidateHTTPURL(s)
	})
}

// FuzzParseRef tests container image reference parsing with
// arbitrary attacker-controlled reference strings.
//
// Image references are the core input to a container registry.
// Parsing bugs affect image pulling, pushing, and replication.
func FuzzParseRef(f *testing.F) {
	f.Add("library/nginx:latest")
	f.Add("docker.io/ubuntu@sha256:abc123")
	f.Add("")
	f.Add(":")
	f.Add(string(make([]byte, 1000)))

	f.Fuzz(func(t *testing.T, s string) {
		if len(s) > 1<<16 {
			return
		}
		_, _, _ = distribution.ParseRef(s)
	})
}
