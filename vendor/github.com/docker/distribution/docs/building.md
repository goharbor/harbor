<!--[metadata]>
+++
title = "Build instructions"
description = "Explains how to build & hack on the registry"
keywords = ["registry, on-prem, images, tags, repository, distribution, build, recipe, advanced"]
+++
<![end-metadata]-->

# Building the registry source

## Use-case

This is useful if you intend to actively work on the registry.

### Alternatives

Most people should use the [official Registry docker image](https://hub.docker.com/r/library/registry/).

People looking for advanced operational use cases might consider rolling their own image with a custom Dockerfile inheriting `FROM registry:2`.

OS X users who want to run natively can do so following [the instructions here](osx-setup-guide.md).

### Gotchas

You are expected to know your way around with go & git.

If you are a casual user with no development experience, and no preliminary knowledge of go, building from source is probably not a good solution for you.

## Build the development environment

The first prerequisite of properly building distribution targets is to have a Go
development environment setup. Please follow [How to Write Go Code](https://golang.org/doc/code.html)
for proper setup. If done correctly, you should have a GOROOT and GOPATH set in the
environment.

If a Go development environment is setup, one can use `go get` to install the
`registry` command from the current latest:

    go get github.com/docker/distribution/cmd/registry

The above will install the source repository into the `GOPATH`.

Now create the directory for the registry data (this might require you to set permissions properly)

    mkdir -p /var/lib/registry

... or alternatively `export REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=/somewhere` if you want to store data into another location.

The `registry`
binary can then be run with the following:

    $ $GOPATH/bin/registry --version
    $GOPATH/bin/registry github.com/docker/distribution v2.0.0-alpha.1+unknown

> __NOTE:__ While you do not need to use `go get` to checkout the distribution
> project, for these build instructions to work, the project must be checked
> out in the correct location in the `GOPATH`. This should almost always be
> `$GOPATH/src/github.com/docker/distribution`.

The registry can be run with the default config using the following
incantation:

    $ $GOPATH/bin/registry $GOPATH/src/github.com/docker/distribution/cmd/registry/config-example.yml
    INFO[0000] endpoint local-5003 disabled, skipping        app.id=34bbec38-a91a-494a-9a3f-b72f9010081f version=v2.0.0-alpha.1+unknown
    INFO[0000] endpoint local-8083 disabled, skipping        app.id=34bbec38-a91a-494a-9a3f-b72f9010081f version=v2.0.0-alpha.1+unknown
    INFO[0000] listening on :5000                            app.id=34bbec38-a91a-494a-9a3f-b72f9010081f version=v2.0.0-alpha.1+unknown
    INFO[0000] debug server listening localhost:5001

If it is working, one should see the above log messages.

### Repeatable Builds

For the full development experience, one should `cd` into
`$GOPATH/src/github.com/docker/distribution`. From there, the regular `go`
commands, such as `go test`, should work per package (please see
[Developing](#developing) if they don't work).

A `Makefile` has been provided as a convenience to support repeatable builds.
Please install the following into `GOPATH` for it to work:

    go get github.com/tools/godep github.com/golang/lint/golint

**TODO(stevvooe):** Add a `make setup` command to Makefile to run this. Have to think about how to interact with Godeps properly.

Once these commands are available in the `GOPATH`, run `make` to get a full
build:

    $ GOPATH=`godep path`:$GOPATH make
    + clean
    + fmt
    + vet
    + lint
    + build
    github.com/docker/docker/vendor/src/code.google.com/p/go/src/pkg/archive/tar
    github.com/Sirupsen/logrus
    github.com/docker/libtrust
    ...
    github.com/yvasiyarov/gorelic
    github.com/docker/distribution/registry/handlers
    github.com/docker/distribution/cmd/registry
    + test
    ...
    ok    github.com/docker/distribution/digest 7.875s
    ok    github.com/docker/distribution/manifest 0.028s
    ok    github.com/docker/distribution/notifications  17.322s
    ?     github.com/docker/distribution/registry [no test files]
    ok    github.com/docker/distribution/registry/api/v2  0.101s
    ?     github.com/docker/distribution/registry/auth  [no test files]
    ok    github.com/docker/distribution/registry/auth/silly  0.011s
    ...
    + /Users/sday/go/src/github.com/docker/distribution/bin/registry
    + /Users/sday/go/src/github.com/docker/distribution/bin/registry-api-descriptor-template
    + binaries

The above provides a repeatable build using the contents of the vendored
Godeps directory. This includes formatting, vetting, linting, building,
testing and generating tagged binaries. We can verify this worked by running
the registry binary generated in the "./bin" directory:

    $ ./bin/registry -version
    ./bin/registry github.com/docker/distribution v2.0.0-alpha.2-80-g16d8b2c.m

### Developing

The above approaches are helpful for small experimentation. If more complex
tasks are at hand, it is recommended to employ the full power of `godep`.

The Makefile is designed to have its `GOPATH` defined externally. This allows
one to experiment with various development environment setups. This is
primarily useful when testing upstream bugfixes, by modifying local code. This
can be demonstrated using `godep` to migrate the `GOPATH` to use the specified
dependencies. The `GOPATH` can be migrated to the current package versions
declared in `Godeps` with the following command:

    godep restore

> **WARNING:** This command will checkout versions of the code specified in
> Godeps/Godeps.json, modifying the contents of `GOPATH`. If this is
> undesired, it is recommended to create a workspace devoted to work on the
> _Distribution_ project.

With a successful run of the above command, one can now use `make` without
specifying the `GOPATH`:

    make

If that is successful, standard `go` commands, such as `go test` should work,
per package, without issue.

### Optional build tags

Optional [build tags](http://golang.org/pkg/go/build/) can be provided using
the environment variable `DOCKER_BUILDTAGS`.

To enable the [Ceph RADOS storage driver](storage-drivers/rados.md)
(librados-dev and librbd-dev will be required to build the bindings):

    export DOCKER_BUILDTAGS='include_rados'
