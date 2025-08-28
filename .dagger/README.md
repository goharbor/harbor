# Dagger Harbor Setup and Usage Guide

## Introduction

This guide explains how to use Dagger to build, run, and test a Harbor setup locally.

## Prerequisites

- Dagger installed
- Harbor source code available
- Generate `consts.go`

## Getting Started

### Generate `consts.go` file
```bash
go run ./.dagger/scripts/parseMakefile.go
```

### Step 1: Build and Spin Up Harbor Components Locally

Run the following command to build and start the necessary Harbor components inside Dagger:

```bash
dagger call run-dev up -v
```

This command will:

- Build the Harbor components.
- Spin up the Harbor backend inside Dagger, similar to running the entire Harbor server.
- Bind Harbor to port `8080`.

### Step 2: Start the Portal

After running the Harbor backend, you can move to the portal directory to spin up the Harbor portal:

```bash
cd ./src/portal
```

> **Note:** The current portal image has some issues.
So follow the instructions in the `./src/portal/README.md` for setup.

Make sure to configure the portal to connect to the Harbor backend by setting the `target-harbor-server` in the portal's configuration to:

```bash
http://localhost:8080
```

This ensures the portal connects to the Harbor running inside Dagger.

### Step 3: Running Harbor Locally

Once both Harbor backend and the portal are set up, you will have a working Harbor setup. You can now use this setup for building, running, and testing Harbor locally.

## Available Functions in Dagger

You can list all available functions in Dagger by running:

```bash
dagger functions
```

This will display a list of functions you can use. Currently, we have the following functions:

### 1. **publish-all-images**

Publishes all images in the registry.

Example usage:

```bash
dagger call publish-all-images --registry-username=harbor-cli --registry=demo.goharbor.io --registry-password=env:REGPASS --image-tags v3.0.0 --version v3.0.0 --debugbin=false --project-name=library/dagger-test -vvv
```

- `-vvv` flag is used for highly verbose output. You can remove this flag for a less verbose output.
- Feel free to change the version and registry flags according to your needs.

### 2. **publish-image**

Publishes a specific image package.

Example usage:

```bash
dagger call publish-image --registry-username=admin --registry=ttl.sh --registry-password=env:REGPASS --image-tags v3.2.2 --version v3.0 --pkg registryctl --debugbin=false --project-name=library/dagger-test -vvv -i
```

This will publish the `registryctl` package.

### 3. **build-binary**

Builds specific Harbor binaries for a given platform.

Example usage:

```bash
dagger call build-binary --pkg core --platform "linux/amd64" --version v2.12.2 --debugbin=false export --path=bin/harbor_core
```

This command will build the `core` package for the `linux/amd64` platform and export the binary to `harbor_core`.

#### Extras
Use these to pull the `dagger.json` and `.dagger` folder
In case if you want to use dagger in other branches this might be helpful. 
Also while using this don't forget to Generate consts file based on the branch 
```
oras pull bupd/harbor-dagger-dir:latest
oras pull bupd/harbor-dagger-json:latest
```

## Conclusion

By following the above steps, you can have a fully functional Harbor setup running inside Dagger. You can use this setup for local development and testing. The available Dagger functions like `publish-all-images`, `publish-image`, and `build-binary` make it easy to manage Harbor images and binaries.
