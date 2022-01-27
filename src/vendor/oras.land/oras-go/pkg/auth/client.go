/*
Copyright The ORAS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/containerd/containerd/remotes"
)

// Common errors
var (
	ErrNotLoggedIn = errors.New("not logged in")
)

// Client provides authentication operations for remotes.
type Client interface {
	// Login logs in to a remote server identified by the hostname.
	// Deprecated: use LoginWithOpts
	Login(ctx context.Context, hostname, username, secret string, insecure bool) error
	// LoginWithOpts logs in to a remote server identified by the hostname with custom options
	LoginWithOpts(options ...LoginOption) error
	// Logout logs out from a remote server identified by the hostname.
	Logout(ctx context.Context, hostname string) error
	// Resolver returns a new authenticated resolver.
	// Deprecated: use ResolverWithOpts
	Resolver(ctx context.Context, client *http.Client, plainHTTP bool) (remotes.Resolver, error)
	// ResolverWithOpts returns a new authenticated resolver with custom options.
	ResolverWithOpts(options ...ResolverOption) (remotes.Resolver, error)
}
