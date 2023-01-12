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
	"strconv"
	"strings"
)

// Scheme define the authentication method.
type Scheme byte

const (
	// SchemeUnknown represents unknown or unsupported schemes
	SchemeUnknown Scheme = iota

	// SchemeBasic represents the "Basic" HTTP authentication scheme.
	// Reference: https://tools.ietf.org/html/rfc7617
	SchemeBasic

	// SchemeBearer represents the Bearer token in OAuth 2.0.
	// Reference: https://tools.ietf.org/html/rfc6750
	SchemeBearer
)

// parseScheme parse the authentication scheme from the given string
// case-insensitively.
func parseScheme(scheme string) Scheme {
	switch {
	case strings.EqualFold(scheme, "basic"):
		return SchemeBasic
	case strings.EqualFold(scheme, "bearer"):
		return SchemeBearer
	}
	return SchemeUnknown
}

// String return the string for the scheme.
func (s Scheme) String() string {
	switch s {
	case SchemeBasic:
		return "Basic"
	case SchemeBearer:
		return "Bearer"
	}
	return "Unknown"
}

// parseChallenge parses the "WWW-Authenticate" header returned by the remote
// registry, and extracts parameters if scheme is Bearer.
// References:
// - https://docs.docker.com/registry/spec/auth/token/#how-to-authenticate
// - https://tools.ietf.org/html/rfc7235#section-2.1
func parseChallenge(header string) (scheme Scheme, params map[string]string) {
	// as defined in RFC 7235 section 2.1, we have
	//     challenge   = auth-scheme [ 1*SP ( token68 / #auth-param ) ]
	//     auth-scheme = token
	//     auth-param  = token BWS "=" BWS ( token / quoted-string )
	//
	// since we focus parameters only on Bearer, we have
	//     challenge   = auth-scheme [ 1*SP #auth-param ]
	schemeString, rest := parseToken(header)
	scheme = parseScheme(schemeString)

	// fast path for non bearer challenge
	if scheme != SchemeBearer {
		return
	}

	// parse params for bearer auth.
	// combining RFC 7235 section 2.1 with RFC 7230 section 7, we have
	//     #auth-param => auth-param *( OWS "," OWS auth-param )
	var key, value string
	for {
		key, rest = parseToken(skipSpace(rest))
		if key == "" {
			return
		}

		rest = skipSpace(rest)
		if rest == "" || rest[0] != '=' {
			return
		}
		rest = skipSpace(rest[1:])
		if rest == "" {
			return
		}

		if rest[0] == '"' {
			prefix, err := strconv.QuotedPrefix(rest)
			if err != nil {
				return
			}
			value, err = strconv.Unquote(prefix)
			if err != nil {
				return
			}
			rest = rest[len(prefix):]
		} else {
			value, rest = parseToken(rest)
			if value == "" {
				return
			}
		}
		if params == nil {
			params = map[string]string{
				key: value,
			}
		} else {
			params[key] = value
		}

		rest = skipSpace(rest)
		if rest == "" || rest[0] != ',' {
			return
		}
		rest = rest[1:]
	}
}

// isNotTokenChar reports whether rune is not a `tchar` defined in RFC 7230
// section 3.2.6.
func isNotTokenChar(r rune) bool {
	// tchar = "!" / "#" / "$" / "%" / "&" / "'" / "*"
	//       / "+" / "-" / "." / "^" / "_" / "`" / "|" / "~"
	//       / DIGIT / ALPHA
	//       ; any VCHAR, except delimiters
	return (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') &&
		(r < '0' || r > '9') && !strings.ContainsRune("!#$%&'*+-.^_`|~", r)
}

// parseToken finds the next token from the given string. If no token found,
// an empty token is returned and the whole of the input is returned in rest.
// Note: Since token = 1*tchar, empty string is not a valid token.
func parseToken(s string) (token, rest string) {
	if i := strings.IndexFunc(s, isNotTokenChar); i != -1 {
		return s[:i], s[i:]
	}
	return s, ""
}

// skipSpace skips "bad" whitespace (BWS) defined in RFC 7230 section 3.2.3.
func skipSpace(s string) string {
	// OWS = *( SP / HTAB )
	//     ; optional whitespace
	// BWS = OWS
	//     ; "bad" whitespace
	if i := strings.IndexFunc(s, func(r rune) bool {
		return r != ' ' && r != '\t'
	}); i != -1 {
		return s[i:]
	}
	return s
}
