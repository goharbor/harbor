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
	"net/http"
)

type (
	// LoginOption allows specifying various settings on login.
	LoginOption func(*LoginSettings)

	// LoginSettings represent all the various settings on login.
	LoginSettings struct {
		Context   context.Context
		Hostname  string
		Username  string
		Secret    string
		Insecure  bool
		UserAgent string
	}
)

// WithLoginContext returns a function that sets the Context setting on login.
func WithLoginContext(context context.Context) LoginOption {
	return func(settings *LoginSettings) {
		settings.Context = context
	}
}

// WithLoginHostname returns a function that sets the Hostname setting on login.
func WithLoginHostname(hostname string) LoginOption {
	return func(settings *LoginSettings) {
		settings.Hostname = hostname
	}
}

// WithLoginUsername returns a function that sets the Username setting on login.
func WithLoginUsername(username string) LoginOption {
	return func(settings *LoginSettings) {
		settings.Username = username
	}
}

// WithLoginSecret returns a function that sets the Secret setting on login.
func WithLoginSecret(secret string) LoginOption {
	return func(settings *LoginSettings) {
		settings.Secret = secret
	}
}

// WithLoginInsecure returns a function that sets the Insecure setting to true on login.
func WithLoginInsecure() LoginOption {
	return func(settings *LoginSettings) {
		settings.Insecure = true
	}
}

// WithLoginUserAgent returns a function that sets the UserAgent setting on login.
func WithLoginUserAgent(userAgent string) LoginOption {
	return func(settings *LoginSettings) {
		settings.UserAgent = userAgent
	}
}

type (
	// ResolverOption allows specifying various settings on the resolver.
	ResolverOption func(*ResolverSettings)

	// ResolverSettings represent all the various settings on a resolver.
	ResolverSettings struct {
		Client    *http.Client
		PlainHTTP bool
		Headers   http.Header
	}
)

// WithResolverClient returns a function that sets the Client setting on resolver.
func WithResolverClient(client *http.Client) ResolverOption {
	return func(settings *ResolverSettings) {
		settings.Client = client
	}
}

// WithResolverPlainHTTP returns a function that sets the PlainHTTP setting to true on resolver.
func WithResolverPlainHTTP() ResolverOption {
	return func(settings *ResolverSettings) {
		settings.PlainHTTP = true
	}
}

// WithResolverHeaders returns a function that sets the Headers setting on resolver.
func WithResolverHeaders(headers http.Header) ResolverOption {
	return func(settings *ResolverSettings) {
		settings.Headers = headers
	}
}
