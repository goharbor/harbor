

# Development Guide

This document is proviede for users development and building `harbor`.

## GitHub workflow

To check out code to work on, please refer to [this guide](https://github.com/kubernetes/community/blob/master/contributors/guide/github-workflow.md) from kubernetes.

## Development Step

1. Pull code `git clone https://github.com/goharbor/harbor.git`.

> when you open it for the first time, there will be a dependency error, because some modules in the project are dynamically generated.

2. Make modules `make gen_apis`

> If the task fails due to network problems during the execution process, you can configure the proxy to solve the problem.
>
> eg:
>
> Set proxy to local network, then docker build use host netowkr.
>
> export http_proxy="http://127.0.0.1:8001";
>
> export HTTP_PROXY="http://127.0.0.1:8001";
>
>  export https_proxy="http://127.0.0.1:8001";
>
> export HTTPS_PROXY="http://127.0.0.1:8001"
>
> change Makefile `DOCKERBUILD=$(DOCKERCMD) build --network host`

3. Developer func
4. Push Code

## Building Harbor

Building harbor on a local OS/shell environment.

1. ### Requirements

- Docker
- Go

2. build：

- `make compile`: compile `core` and `jobservice` code

- `make build`: build Harbor docker images from photon baseimage

## Dependency management

Harbor uses [go modules](https://github.com/golang/go/wiki/Modules) to manage dependencies.



