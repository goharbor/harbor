## Web UI

```
cd cmd/workwebui
go run main.go
open "http://localhost:5040/"
```

## Assets

Web UI frontend is written in [react](https://facebook.github.io/react/). [Webpack](https://webpack.github.io/) is used to transpile and bundle es7 and jsx to run on modern browsers.
Finally bundled js is embedded in a go file.

All NPM commands can be found in `package.json`.

- fetch dependency: `npm install`
- test: `npm test`
- generate test coverage: `npm run cover`
- lint: `npm run lint`
- bundle for production: `npm run build`
- bundle for testing: `npm run dev`

To embed bundled js, do

```
go get -u github.com/jteeuwen/go-bindata/...
cd webui/internal/assets
go generate
```
