
## v1.0.16 / 2020-11-23
  
  * Fix regexp matching issues in headers (#59)

## v1.0.15 / 2019-07-03
  
  * NewMatcher() will now return objects that completely separate one another. (#55)
  * add request Options (#49)
  * fix typo: function -> func (#52)
  * feat(docs): change note
  * feat(docs): add net/http support
  * Add Basic Auth (#47)
  * Update LICENSE (#46)

## v1.0.14 / 2019-01-31

  * feat(version): bump to v1.0.14
  * feat: add go.mod

## v1.0.13 / 2019-01-30

  * Add PathParam matcher (#42)

## v1.0.12 / 2018-11-13

  * Fix possible data race. (#41)

## v1.0.11 / 2018-10-29

  * Do not reset response body (#40)
  * refactor(travis): remove unsupported versions for golint based on Go release policy support
  * feat(gock): add gock.Observe to support inspection of the outgoing intercepted HTTP traffic (#38)

## v1.0.10 / 2018-09-09

  * Support multiple response headers with same name #35 (#36)

## v1.0.9 / 2018-06-14

  * fix(url-encoding) add exact match test in MatchPath (#34)
  * fix(travis): use string notation for Go versions

## v1.0.8 / 2018-02-28

  * chore(LICENSE): update year ;)
  * feat(docs): add additional tips and examples
  * feat(gock): ignore already intercepted http.Client

## v1.0.7 / 2017-12-21

  * Make MatchHost case insensitive. (#31)
  * refactor(docs): remove codesponsor :(
  * add example when request reply with error (#28)
  * feat(docs): add sponsor ad
  * Add example networking partially enabled (#23)

## v1.0.6 / 2017-07-27

  * fix(#23): mock transport deadlock

## v1.0.5 / 2017-07-26

  * feat(#25, #24): use content type only if missing while matching JSON/XML
  * feat(#24): add CleanUnmatchedRequests() and OffAll() public functions
  * feat(version): bump to v1.0.5
  * fix(store): use proper indent style
  * fix(mutex): use different mutex for store
  * feat(travis): add Go 1.8 CI support

## v1.0.4 / 2017-02-14

  * Update README to include most up to date version (#17)
  * Update MatchBody() to compare if key + value pairs of JSON match regardless of order they are in. (#16)
  * feat(examples): add new example for unmatch case
  * refactor(docs): add pook reference

## 1.0.3 / 14-11-2016

- feat(#13): adds `GetUnmatchedRequests()` and `HasUnmatchedRequests()` API functions.

## 1.0.2 / 10-11-2016

- fix(#11): adds `Compression()` method for output HTTP traffic body compression processing and matching.

## 1.0.1 / 07-09-2016

- fix(#9): missing URL query param matcher.

## 1.0.0 / 19-04-2016

- feat(version): first major version release.

## 0.1.6 / 19-04-2016

- fix(#7): if error configured, RoundTripper should reply with `nil` response.

## 0.1.5 / 09-04-2016

- feat(#5): support `ReplyFunc` for convenience.

## 0.1.4 / 16-03-2016

- feat(api): add `IsDone()` method.
- fix(responder): return mock error if present.
- feat(#4): support define request/response body from file disk.

## 0.1.3 / 09-03-2016

- feat(matcher): add content type matcher helper method supporting aliases.
- feat(interceptor): add function to restore HTTP client transport.
- feat(matcher): add URL scheme matcher function.
- fix(request): ignore base slash path.
- feat(api): add Off() method for easier restore and clean up.
- feat(store): add public API for pending mocks.

## 0.1.2 / 04-03-2016

- fix(matcher): body matchers no used by default.
- feat(matcher): add matcher factories for multiple cases.

## 0.1.1 / 04-03-2016

- fix(params): persist query params accordingly.

## 0.1.0 / 02-03-2016

- First release.
