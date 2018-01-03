
## Example API Client with Sling

Try the example Github API Client.

    cd examples
    go get .

List the public issues on the [github.com/golang/go](https://github.com/golang/go) repository.
    
    go run github.go

To list your public and private Github issues, pass your [Github Access Token](https://github.com/settings/tokens)

    go run github.go -access-token=xxx

or set the `GITHUB_ACCESS_TOKEN` environment variable.

For a complete Github API, see the excellent [google/go-github](https://github.com/google/go-github) package.