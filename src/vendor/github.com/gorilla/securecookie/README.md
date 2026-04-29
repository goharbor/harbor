# gorilla/securecookie

![testing](https://github.com/gorilla/securecookie/actions/workflows/test.yml/badge.svg)
[![codecov](https://codecov.io/github/gorilla/securecookie/branch/main/graph/badge.svg)](https://codecov.io/github/gorilla/securecookie)
[![godoc](https://godoc.org/github.com/gorilla/securecookie?status.svg)](https://godoc.org/github.com/gorilla/securecookie)
[![sourcegraph](https://sourcegraph.com/github.com/gorilla/securecookie/-/badge.svg)](https://sourcegraph.com/github.com/gorilla/securecookie?badge)

![Gorilla Logo](https://github.com/gorilla/.github/assets/53367916/d92caabf-98e0-473e-bfbf-ab554ba435e5)

securecookie encodes and decodes authenticated and optionally encrypted
cookie values.

Secure cookies can't be forged, because their values are validated using HMAC.
When encrypted, the content is also inaccessible to malicious eyes. It is still
recommended that sensitive data not be stored in cookies, and that HTTPS be used
to prevent cookie [replay attacks](https://en.wikipedia.org/wiki/Replay_attack).

## Examples

To use it, first create a new SecureCookie instance:

```go
// Hash keys should be at least 32 bytes long
var hashKey = []byte("very-secret")
// Block keys should be 16 bytes (AES-128) or 32 bytes (AES-256) long.
// Shorter keys may weaken the encryption used.
var blockKey = []byte("a-lot-secret")
var s = securecookie.New(hashKey, blockKey)
```

The hashKey is required, used to authenticate the cookie value using HMAC.
It is recommended to use a key with 32 or 64 bytes.

The blockKey is optional, used to encrypt the cookie value -- set it to nil
to not use encryption. If set, the length must correspond to the block size
of the encryption algorithm. For AES, used by default, valid lengths are
16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.

Strong keys can be created using the convenience function
`GenerateRandomKey()`. Note that keys created using `GenerateRandomKey()` are not
automatically persisted. New keys will be created when the application is
restarted, and previously issued cookies will not be able to be decoded.

Once a SecureCookie instance is set, use it to encode a cookie value:

```go
func SetCookieHandler(w http.ResponseWriter, r *http.Request) {
	value := map[string]string{
		"foo": "bar",
	}
	if encoded, err := s.Encode("cookie-name", value); err == nil {
		cookie := &http.Cookie{
			Name:  "cookie-name",
			Value: encoded,
			Path:  "/",
			Secure: true,
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)
	}
}
```

Later, use the same SecureCookie instance to decode and validate a cookie
value:

```go
func ReadCookieHandler(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("cookie-name"); err == nil {
		value := make(map[string]string)
		if err = s2.Decode("cookie-name", cookie.Value, &value); err == nil {
			fmt.Fprintf(w, "The value of foo is %q", value["foo"])
		}
	}
}
```

We stored a map[string]string, but secure cookies can hold any value that
can be encoded using `encoding/gob`. To store custom types, they must be
registered first using gob.Register(). For basic types this is not needed;
it works out of the box. An optional JSON encoder that uses `encoding/json` is
available for types compatible with JSON.

### Key Rotation
Rotating keys is an important part of any security strategy. The `EncodeMulti` and
`DecodeMulti` functions allow for multiple keys to be rotated in and out.
For example, let's take a system that stores keys in a map:

```go
// keys stored in a map will not be persisted between restarts
// a more persistent storage should be considered for production applications.
var cookies = map[string]*securecookie.SecureCookie{
	"previous": securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32),
	),
	"current": securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32),
	),
}
```

Using the current key to encode new cookies:
```go
func SetCookieHandler(w http.ResponseWriter, r *http.Request) {
	value := map[string]string{
		"foo": "bar",
	}
	if encoded, err := securecookie.EncodeMulti("cookie-name", value, cookies["current"]); err == nil {
		cookie := &http.Cookie{
			Name:  "cookie-name",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(w, cookie)
	}
}
```

Later, decode cookies. Check against all valid keys:
```go
func ReadCookieHandler(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("cookie-name"); err == nil {
		value := make(map[string]string)
		err = securecookie.DecodeMulti("cookie-name", cookie.Value, &value, cookies["current"], cookies["previous"])
		if err == nil {
			fmt.Fprintf(w, "The value of foo is %q", value["foo"])
		}
	}
}
```

Rotate the keys. This strategy allows previously issued cookies to be valid until the next rotation:
```go
func Rotate(newCookie *securecookie.SecureCookie) {
	cookies["previous"] = cookies["current"]
	cookies["current"] = newCookie
}
```

## License

BSD licensed. See the LICENSE file for details.
