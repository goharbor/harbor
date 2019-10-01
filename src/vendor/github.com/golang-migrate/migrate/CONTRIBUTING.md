# Development, Testing and Contributing

  1. Make sure you have a running Docker daemon
     (Install for [MacOS](https://docs.docker.com/docker-for-mac/))
  1. Fork this repo and `git clone` somewhere to `$GOPATH/src/github.com/golang-migrate/migrate`
  1. Install [dep](https://github.com/golang/dep) and run `dep ensure` to pull dependencies
  1. Confirm tests are working: `make test-short`
  1. Write awesome code ...
  1. `make test` to run all tests against all database versions
  1. Push code and open Pull Request

Some more helpful commands:

  * You can specify which database/ source tests to run:
    `make test-short SOURCE='file go_bindata' DATABASE='postgres cassandra'`
  * After `make test`, run `make html-coverage` which opens a shiny test coverage overview.
  * `make build-cli` builds the CLI in directory `cli/build/`.
  * `make list-external-deps` lists all external dependencies for each package
  * `make docs && make open-docs` opens godoc in your browser, `make kill-docs` kills the godoc server.
    Repeatedly call `make docs` to refresh the server.
  * Set the `DOCKER_API_VERSION` environment variable to the latest supported version if you get errors regarding the docker client API version being too new.
