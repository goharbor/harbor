# Sling Changelog

Notable changes between releases.

## Latest

## v1.4.2

* Update Go module dependencies

## v1.4.1

* Update minimum Go version to v1.18 ([#76](https://github.com/dghubble/sling/pull/76))

## v1.4.0

* `Do` reads Body to reuse HTTP/1.x "keep-alive" TCP connections ([#59](https://github.com/dghubble/sling/pull/59))
* `Receive` skips decoding if status is 204 (no content) ([#63](https://github.com/dghubble/sling/pull/63))

## v1.3.0

* Add Sling `ResponseDecoder` setter for receiving responses with a custom `ResponseDecoder` ([#49](https://github.com/dghubble/sling/pull/49))
* Add Go module support (i.e. `go.mod`). Exclude `examples` (multi-module). ([#52](https://github.com/dghubble/sling/pull/52))

## v1.2.0

* Add `Connect`, `Options`, and `Trace` HTTP methods ([c51967](https://github.com/dghubble/sling/commit/c519674860ff275e0ceb12caf5d87b31765c4e71))
* Skip receiving (i.e. decoding) `204 No Content` responses ([#31](https://github.com/dghubble/sling/pull/31))

## v1.1.0

* Allow JSON decoding, regardless of response Content-Type (#26)
* Add `BodyProvider` interface and setter so request Body encoding can be customized (#23)
* Add `Doer` interface and setter so request sending behavior can be customized (#21)
* Add `SetBasicAuth` setter for Authorization headers (#16)
* Add Sling `Body` setter to set an `io.Reader` on the Request (#9)

## v1.0.0

* Added support for receiving and decoding error JSON structs
* Renamed Sling `JsonBody` setter to `BodyJSON` (breaking)
* Renamed Sling `BodyStruct` setter to `BodyForm` (breaking)
* Renamed Sling fields `httpClient`, `method`, `rawURL`, and `header` to be internal (breaking)
* Changed `Do` and `Receive` to skip response JSON decoding if "application/json" Content-Type is missing
* Changed `Sling.Receive(v interface{})` to `Sling.Receive(successV, failureV interface{})` (breaking)
    * Previously `Receive` attempted to decode the response Body in all cases
    * Updated `Receive` will decode the response Body into successV for 2XX responses or decode the Body into failureV for other status codes. Pass a nil `successV` or `failureV` to skip JSON decoding into that value.
    * To upgrade, pass nil for the `failureV` argument or consider defining a JSON tagged struct appropriate for the API endpoint. (e.g. `s.Receive(&issue, nil)`, `s.Receive(&issue, &githubError)`)
    * To retain the old behavior, duplicate the first argument (e.g. s.Receive(&tweet, &tweet))
* Changed `Sling.Do(http.Request, v interface{})` to `Sling.Do(http.Request, successV, failureV interface{})` (breaking)
    * See the changelog entry about `Receive`, the upgrade path is the same.
* Removed HEAD, GET, POST, PUT, PATCH, DELETE constants, no reason to export them (breaking)

## v0.4.0

* Improved golint compliance
* Fixed typos and test printouts

## v0.3.0

* Added BodyStruct method for setting a url encoded form body on the Request
* Added Add and Set methods for adding or setting Request Headers
* Added JsonBody method for setting JSON Request Body
* Improved examples and documentation

## v0.2.0

* Added http.Client setter
* Added Sling.New() method to return a copy of a Sling
* Added Base setter and Path extension support
* Added method setters (Get, Post, Put, Patch, Delete, Head)
* Added support for encoding URL Query parameters
* Added example tiny Github API
* Changed v0.1.0 method signatures and names (breaking)
* Removed Go 1.0 support

## v0.1.0

* Support decoding JSON responses.


