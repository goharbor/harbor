// Copyright 2020 Ant Group. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"crypto/tls"
	"encoding/base64"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/containerd/containerd/remotes/docker"
	dockerconfig "github.com/docker/cli/cli/config"
	"github.com/pkg/errors"

	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/remote"
)

func newDefaultClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          10,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 5 * time.Second,
			DisableKeepAlives:     true,
			TLSNextProto:          make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		},
	}
}

// withCredentialFunc accepts host url parameter and returns with
// username, password and error.
type withCredentialFunc = func(string) (string, string, error)

// withRemote creates an remote instance, it uses the implemention of containerd
// docker remote to access image from remote registry.
func withRemote(ref string, insecure bool, credFunc withCredentialFunc) (*remote.Remote, error) {
	registryHosts := docker.ConfigureDefaultRegistries(
		docker.WithAuthorizer(docker.NewAuthorizer(
			newDefaultClient(),
			credFunc,
		)),
		docker.WithClient(newDefaultClient()),
		docker.WithPlainHTTP(func(host string) (bool, error) {
			_insecure, err := docker.MatchLocalhost(host)
			if err != nil {
				return false, err
			}
			if _insecure {
				return true, nil
			}
			return insecure, nil
		}),
	)

	resolver := docker.NewResolver(docker.ResolverOptions{
		Hosts: registryHosts,
	})

	remote, err := remote.New(ref, resolver)
	if err != nil {
		return nil, err
	}

	return remote, nil
}

// DefaultRemote creates an remote instance, it attempts to read docker auth config
// file `$DOCKER_CONFIG/config.json` to communicate with remote registry, `$DOCKER_CONFIG`
// defaults to `~/.docker`.
func DefaultRemote(ref string, insecure bool) (*remote.Remote, error) {
	return withRemote(ref, insecure, func(host string) (string, string, error) {
		// The host of docker hub image will be converted to `registry-1.docker.io` in:
		// github.com/containerd/containerd/remotes/docker/registry.go
		// But we need use the key `https://index.docker.io/v1/` to find auth from docker config.
		if host == "registry-1.docker.io" {
			host = "https://index.docker.io/v1/"
		}

		config := dockerconfig.LoadDefaultConfigFile(os.Stderr)
		authConfig, err := config.GetAuthConfig(host)
		if err != nil {
			return "", "", err
		}

		return authConfig.Username, authConfig.Password, nil
	})
}

// DefaultRemoteWithAuth creates an remote instance, it parses base64 encoded auth string
// to communicate with remote registry.
func DefaultRemoteWithAuth(ref string, insecure bool, auth string) (*remote.Remote, error) {
	return withRemote(ref, insecure, func(host string) (string, string, error) {
		decoded, err := base64.StdEncoding.DecodeString(auth)
		if err != nil {
			return "", "", errors.Wrap(err, "Decode base64 encoded auth string")
		}
		ary := strings.Split(string(decoded), ":")
		if len(ary) != 2 {
			return "", "", errors.New("Invalid base64 encoded auth string")
		}
		return ary[0], ary[1], nil
	})
}
