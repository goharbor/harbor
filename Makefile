# Makefile for Harbor project
#
# Targets:
#
#
# other example:
#	clean specific version binaries and images:
#				make clean -e VERSIONTAG=[TAG]
#				note**: If commit new code to github, the git commit TAG will \
#				change. Better use this command clean previous images and \
#				files with specific TAG.
#   By default DEVFLAG=true, if you want to release new version of Harbor, \
#		should setting the flag to false.
#				make XXXX -e DEVFLAG=false

#!/usr/bin/env bash

#
# Makefile with some common workflow for dev, build and test
#

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

# The following help command is Licensed under the Apache License, Version 2.0 (the "License")
# Copyright 2023 The Kubernetes Authors.
.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: all
all: install ## prepare env, compile binaries, build images and install images

SHELL := /bin/bash
BUILDPATH=$(CURDIR)
MAKEPATH=$(BUILDPATH)/make
MAKE_PREPARE_PATH=$(MAKEPATH)/photon/prepare
SRCPATH=./src
TOOLSPATH=$(BUILDPATH)/tools
CORE_PATH=$(BUILDPATH)/src/core
PORTAL_PATH=$(BUILDPATH)/src/portal
CHECKENVCMD=checkenv.sh

# parameters
REGISTRYSERVER=
REGISTRYPROJECTNAME=goharbor
DEVFLAG=true
TRIVYFLAG=false
HTTPPROXY=
BUILDREG=true
BUILDTRIVYADP=true
NPM_REGISTRY=https://registry.npmjs.org
BUILDTARGET=build
GEN_TLS=

# version prepare
# for docker image tag
VERSIONTAG=dev
# for base docker image tag
BUILD_BASE=true
PUSHBASEIMAGE=false
BASEIMAGETAG=dev
BUILDBASETARGET=trivy-adapter core db jobservice log nginx portal prepare redis registry registryctl exporter
IMAGENAMESPACE=goharbor
BASEIMAGENAMESPACE=goharbor
# #input true/false only
PULL_BASE_FROM_DOCKERHUB=true

# for harbor package name
PKGVERSIONTAG=dev

PREPARE_VERSION_NAME=versions

#versions
REGISTRYVERSION=v2.8.3-patch-redis
TRIVYVERSION=v0.58.2
TRIVYADAPTERVERSION=v0.32.3
NODEBUILDIMAGE=node:16.18.0

# version of registry for pulling the source code
REGISTRY_SRC_TAG=v2.8.3
# source of upstream distribution code
DISTRIBUTION_SRC=https://github.com/distribution/distribution.git

# dependency binaries
REGISTRYURL=https://storage.googleapis.com/harbor-builds/bin/registry/release-${REGISTRYVERSION}/registry
TRIVY_DOWNLOAD_URL=https://github.com/aquasecurity/trivy/releases/download/$(TRIVYVERSION)/trivy_$(TRIVYVERSION:v%=%)_Linux-64bit.tar.gz
TRIVY_ADAPTER_DOWNLOAD_URL=https://github.com/goharbor/harbor-scanner-trivy/archive/refs/tags/$(TRIVYADAPTERVERSION).tar.gz

define VERSIONS_FOR_PREPARE
VERSION_TAG: $(VERSIONTAG)
REGISTRY_VERSION: $(REGISTRYVERSION)
TRIVY_VERSION: $(TRIVYVERSION)
TRIVY_ADAPTER_VERSION: $(TRIVYADAPTERVERSION)
endef

# docker parameters
DOCKERCMD=$(shell which docker)
DOCKERBUILD=$(DOCKERCMD) build
DOCKERRMIMAGE=$(DOCKERCMD) rmi
DOCKERPULL=$(DOCKERCMD) pull
DOCKERIMAGES=$(DOCKERCMD) images
DOCKERSAVE=$(DOCKERCMD) save
DOCKERCOMPOSECMD=$(shell which docker-compose)
DOCKERTAG=$(DOCKERCMD) tag

# go parameters
GOCMD=$(shell which go)
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOINSTALL=$(GOCMD) install
GOTEST=$(GOCMD) test
GODEP=$(GOTEST) -i
GOFMT=gofmt -w
GOBUILDIMAGE=golang:1.23.2
GOBUILDPATHINCONTAINER=/harbor

# go build
PKG_PATH=github.com/goharbor/harbor/src/pkg
GITCOMMIT := $(shell git rev-parse --short=8 HEAD)
RELEASEVERSION := $(shell cat VERSION)
GOFLAGS="-buildvcs=false"
GOTAGS=$(if $(GOBUILDTAGS),-tags "$(GOBUILDTAGS)",)
GOLDFLAGS=$(if $(GOBUILDLDFLAGS),--ldflags "-w -s $(GOBUILDLDFLAGS)",)
CORE_LDFLAGS=-X $(PKG_PATH)/version.GitCommit=$(GITCOMMIT) -X $(PKG_PATH)/version.ReleaseVersion=$(RELEASEVERSION)
ifneq ($(GOBUILDLDFLAGS),)
	CORE_LDFLAGS += $(GOBUILDLDFLAGS)
endif

# go build command
GOIMAGEBUILDCMD=/usr/local/go/bin/go build
GOIMAGEBUILD_COMMON=$(GOIMAGEBUILDCMD) $(GOFLAGS) ${GOTAGS} ${GOLDFLAGS}
GOIMAGEBUILD_CORE=$(GOIMAGEBUILDCMD) $(GOFLAGS) ${GOTAGS} --ldflags "-w -s $(CORE_LDFLAGS)"

GOBUILDPATH_CORE=$(GOBUILDPATHINCONTAINER)/src/core
GOBUILDPATH_JOBSERVICE=$(GOBUILDPATHINCONTAINER)/src/jobservice
GOBUILDPATH_REGISTRYCTL=$(GOBUILDPATHINCONTAINER)/src/registryctl
GOBUILDPATH_STANDALONE_DB_MIGRATOR=$(GOBUILDPATHINCONTAINER)/src/cmd/standalone-db-migrator
GOBUILDPATH_EXPORTER=$(GOBUILDPATHINCONTAINER)/src/cmd/exporter
GOBUILDMAKEPATH=make
GOBUILDMAKEPATH_CORE=$(GOBUILDMAKEPATH)/photon/core
GOBUILDMAKEPATH_JOBSERVICE=$(GOBUILDMAKEPATH)/photon/jobservice
GOBUILDMAKEPATH_REGISTRYCTL=$(GOBUILDMAKEPATH)/photon/registryctl
GOBUILDMAKEPATH_STANDALONE_DB_MIGRATOR=$(GOBUILDMAKEPATH)/photon/standalone-db-migrator
GOBUILDMAKEPATH_EXPORTER=$(GOBUILDMAKEPATH)/photon/exporter

# binary
CORE_BINARYPATH=$(BUILDPATH)/$(GOBUILDMAKEPATH_CORE)
CORE_BINARYNAME=harbor_core
JOBSERVICEBINARYPATH=$(BUILDPATH)/$(GOBUILDMAKEPATH_JOBSERVICE)
JOBSERVICEBINARYNAME=harbor_jobservice
REGISTRYCTLBINARYPATH=$(BUILDPATH)/$(GOBUILDMAKEPATH_REGISTRYCTL)
REGISTRYCTLBINARYNAME=harbor_registryctl
STANDALONE_DB_MIGRATOR_BINARYPATH=$(BUILDPATH)/$(GOBUILDMAKEPATH_STANDALONE_DB_MIGRATOR)
STANDALONE_DB_MIGRATOR_BINARYNAME=migrate

# configfile
CONFIGPATH=$(MAKEPATH)
INSIDE_CONFIGPATH=/compose_location
CONFIGFILE=harbor.yml

# prepare parameters
PREPAREPATH=$(TOOLSPATH)
PREPARECMD=prepare
PREPARECMD_PARA=--conf $(INSIDE_CONFIGPATH)/$(CONFIGFILE)
ifeq ($(TRIVYFLAG), true)
	PREPARECMD_PARA+= --with-trivy
endif

# makefile
MAKEFILEPATH_PHOTON=$(MAKEPATH)/photon

# common dockerfile
DOCKERFILEPATH_COMMON=$(MAKEPATH)/common

# docker image name
DOCKER_IMAGE_NAME_PREPARE=$(IMAGENAMESPACE)/prepare
DOCKERIMAGENAME_PORTAL=$(IMAGENAMESPACE)/harbor-portal
DOCKERIMAGENAME_CORE=$(IMAGENAMESPACE)/harbor-core
DOCKERIMAGENAME_JOBSERVICE=$(IMAGENAMESPACE)/harbor-jobservice
DOCKERIMAGENAME_LOG=$(IMAGENAMESPACE)/harbor-log
DOCKERIMAGENAME_DB=$(IMAGENAMESPACE)/harbor-db
DOCKERIMAGENAME_REGCTL=$(IMAGENAMESPACE)/harbor-registryctl
DOCKERIMAGENAME_EXPORTER=$(IMAGENAMESPACE)/harbor-exporter

# docker-compose files
DOCKERCOMPOSEFILEPATH=$(MAKEPATH)
DOCKERCOMPOSEFILENAME=docker-compose.yml

SEDCMD=$(shell which sed)
SEDCMDI=$(SEDCMD) -i
ifeq ($(shell uname),Darwin)
    SEDCMDI=$(SEDCMD) -i ''
endif

# cmds
DOCKERSAVE_PARA=$(DOCKER_IMAGE_NAME_PREPARE):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_CORE):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_LOG):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_DB):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_REGCTL):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_EXPORTER):$(VERSIONTAG) \
		$(IMAGENAMESPACE)/redis-photon:$(VERSIONTAG) \
		$(IMAGENAMESPACE)/nginx-photon:$(VERSIONTAG) \
		$(IMAGENAMESPACE)/registry-photon:$(VERSIONTAG)

DOCKERCOMPOSE_FILE_OPT=-f $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)

ifeq ($(TRIVYFLAG), true)
	DOCKERSAVE_PARA+= $(IMAGENAMESPACE)/trivy-adapter-photon:$(VERSIONTAG)
endif


RUNCONTAINER=$(DOCKERCMD) run --rm -u $(shell id -u):$(shell id -g) -v $(BUILDPATH):$(BUILDPATH) -w $(BUILDPATH)
##@ Build

.PHONY: install
install: compile build prepare start ## Compile binaries, build images, prepare env and start Harbor instance

.PHONY: compile
compile: check_environment versions_prepare compile_core compile_jobservice compile_registryctl ## Compile core and jobservice code

.PHONY: build
build:	## Build Harbor docker images from photon baseimage
# PUSHBASEIMAGE should not be true if BUILD_BASE is not true
	@if [ "$(PULL_BASE_FROM_DOCKERHUB)" != "true" ] && [ "$(PULL_BASE_FROM_DOCKERHUB)" != "false" ] ; then \
		echo set PULL_BASE_FROM_DOCKERHUB to true or false.; exit 1; \
	fi
	@if [ "$(BUILD_BASE)" != "true" ]  && [ "$(PUSHBASEIMAGE)" = "true" ] ; then \
		echo Do not push base images since no base images built. ; \
		exit 1; \
	fi
# PULL_BASE_FROM_DOCKERHUB should be true if BUILD_BASE is not true
	@if [ "$(BUILD_BASE)" != "true" ]  && [ "$(PULL_BASE_FROM_DOCKERHUB)" = "false" ] ; then \
		echo Should pull base images from registry in docker configuration since no base images built. ; \
		exit 1; \
	fi
	make -f $(MAKEFILEPATH_PHOTON)/Makefile $(BUILDTARGET) -e DEVFLAG=$(DEVFLAG) -e GOBUILDIMAGE=$(GOBUILDIMAGE) -e NODEBUILDIMAGE=$(NODEBUILDIMAGE) \
	 -e REGISTRYVERSION=$(REGISTRYVERSION) -e REGISTRY_SRC_TAG=$(REGISTRY_SRC_TAG)  -e DISTRIBUTION_SRC=$(DISTRIBUTION_SRC)\
	 -e TRIVYVERSION=$(TRIVYVERSION) -e TRIVYADAPTERVERSION=$(TRIVYADAPTERVERSION) \
	 -e VERSIONTAG=$(VERSIONTAG) \
	 -e BUILDREG=$(BUILDREG) -e BUILDTRIVYADP=$(BUILDTRIVYADP) \
	 -e NPM_REGISTRY=$(NPM_REGISTRY) -e BASEIMAGETAG=$(BASEIMAGETAG) -e IMAGENAMESPACE=$(IMAGENAMESPACE) -e BASEIMAGENAMESPACE=$(BASEIMAGENAMESPACE) \
	 -e REGISTRYURL=$(REGISTRYURL) \
	 -e TRIVY_DOWNLOAD_URL=$(TRIVY_DOWNLOAD_URL) -e TRIVY_ADAPTER_DOWNLOAD_URL=$(TRIVY_ADAPTER_DOWNLOAD_URL) \
	 -e PULL_BASE_FROM_DOCKERHUB=$(PULL_BASE_FROM_DOCKERHUB) -e BUILD_BASE=$(BUILD_BASE) \
	 -e REGISTRYUSER=$(REGISTRYUSER) -e REGISTRYPASSWORD=$(REGISTRYPASSWORD) \
	 -e PUSHBASEIMAGE=$(PUSHBASEIMAGE)

.PHONY: build_standalone_db_migrator
build_standalone_db_migrator: compile_standalone_db_migrator ## Build only the db migrator
	make -f $(MAKEFILEPATH_PHOTON)/Makefile _build_standalone_db_migrator -e BASEIMAGETAG=$(BASEIMAGETAG) -e VERSIONTAG=$(VERSIONTAG)

.PHONY: build_base_docker
build_base_docker: ## Build only the docker base image
	if [ -n "$(REGISTRYUSER)" ] && [ -n "$(REGISTRYPASSWORD)" ] ; then \
		docker login -u $(REGISTRYUSER) -p $(REGISTRYPASSWORD) ; \
	else \
		echo "No docker credentials provided, please make sure enough privileges to access docker hub!" ; \
	fi
	@for name in $(BUILDBASETARGET); do \
		echo $$name ; \
		sleep 30 ; \
		$(DOCKERBUILD) --pull --no-cache -f $(MAKEFILEPATH_PHOTON)/$$name/Dockerfile.base -t $(BASEIMAGENAMESPACE)/harbor-$$name-base:$(BASEIMAGETAG) --label base-build-date=$(date +"%Y%m%d") . ; \
		if [ "$(PUSHBASEIMAGE)" != "false" ] ; then \
			$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(BASEIMAGENAMESPACE)/harbor-$$name-base:$(BASEIMAGETAG) $(REGISTRYUSER) $(REGISTRYPASSWORD) || exit 1; \
		fi ; \
	done

.PHONY: compile_core
compile_core: gen_apis ## Compile the core binary
	@echo "compiling binary for core (golang image)..."
	@echo $(GOBUILDPATHINCONTAINER)
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATHINCONTAINER) -w $(GOBUILDPATH_CORE) $(GOBUILDIMAGE) $(GOIMAGEBUILD_CORE) -o $(GOBUILDPATHINCONTAINER)/$(GOBUILDMAKEPATH_CORE)/$(CORE_BINARYNAME)
	@echo "Done."

.PHONY: compile_jobservice
compile_jobservice: ## Compile jobservice binary
	@echo "compiling binary for jobservice (golang image)..."
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATHINCONTAINER) -w $(GOBUILDPATH_JOBSERVICE) $(GOBUILDIMAGE) $(GOIMAGEBUILD_COMMON) -o $(GOBUILDPATHINCONTAINER)/$(GOBUILDMAKEPATH_JOBSERVICE)/$(JOBSERVICEBINARYNAME)
	@echo "Done."

.PHONY: compile_registryctl
compile_registryctl: ## Compile registryctl binary
	@echo "compiling binary for harbor registry controller (golang image)..."
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATHINCONTAINER) -w $(GOBUILDPATH_REGISTRYCTL) $(GOBUILDIMAGE) $(GOIMAGEBUILD_COMMON) -o $(GOBUILDPATHINCONTAINER)/$(GOBUILDMAKEPATH_REGISTRYCTL)/$(REGISTRYCTLBINARYNAME)
	@echo "Done."

.PHONY: compile_standalone_db_migrator
compile_standalone_db_migrator: ## Compile standalone db-migrator binary
	@echo "compiling binary for standalone db migrator (golang image)..."
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATHINCONTAINER) -w $(GOBUILDPATH_STANDALONE_DB_MIGRATOR) $(GOBUILDIMAGE) $(GOIMAGEBUILD_COMMON) -o $(GOBUILDPATHINCONTAINER)/$(GOBUILDMAKEPATH_STANDALONE_DB_MIGRATOR)/$(STANDALONE_DB_MIGRATOR_BINARYNAME)
	@echo "Done."

##@ Package

TARCMD=$(shell which tar)
ZIPCMD=$(shell which gzip)
DOCKERIMGFILE=harbor
HARBORPKG=harbor

PACKAGE_OFFLINE_PARA=-zcvf harbor-offline-installer-$(PKGVERSIONTAG).tgz \
					$(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tar.gz \
					$(HARBORPKG)/prepare \
					$(HARBORPKG)/LICENSE $(HARBORPKG)/install.sh \
					$(HARBORPKG)/common.sh \
					$(HARBORPKG)/harbor.yml.tmpl

PACKAGE_ONLINE_PARA=-zcvf harbor-online-installer-$(PKGVERSIONTAG).tgz \
					$(HARBORPKG)/prepare \
					$(HARBORPKG)/LICENSE \
					$(HARBORPKG)/install.sh \
					$(HARBORPKG)/common.sh \
					$(HARBORPKG)/harbor.yml.tmpl

.PHONY: package_online
package_online: update_prepare_version ## Prepare online install package
	# For example: make package_online -e DEVFLAG=false
	#                                     REGISTRYSERVER=reg-bj.goharbor.io
	#                                     REGISTRYPROJECTNAME=harborrelease
	@echo "packing online package ..."
	@cp -r make $(HARBORPKG)
	@if [ -n "$(REGISTRYSERVER)" ] ; then \
		$(SEDCMDI) -e 's/image\: $(IMAGENAMESPACE)/image\: $(REGISTRYSERVER)\/$(REGISTRYPROJECTNAME)/' \
		$(HARBORPKG)/docker-compose.yml ; \
	fi
	@cp LICENSE $(HARBORPKG)/LICENSE

	@$(TARCMD) $(PACKAGE_ONLINE_PARA)
	@rm -rf $(HARBORPKG)
	@echo "Done."

.PHONY: package_offline
package_offline: update_prepare_version compile build ## Prepare offline install package

	@echo "packing offline package ..."
	@cp -r make $(HARBORPKG)
	@cp LICENSE $(HARBORPKG)/LICENSE

	@echo "saving harbor docker image"
	@$(DOCKERSAVE) $(DOCKERSAVE_PARA) > $(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tar
	@gzip $(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tar

	@$(TARCMD) $(PACKAGE_OFFLINE_PARA)
	@rm -rf $(HARBORPKG)
	@echo "Done."

##@ Publish

PUSHSCRIPTPATH=$(MAKEPATH)
PUSHSCRIPTNAME=pushimage.sh
REGISTRYUSER=
REGISTRYPASSWORD=

.PHONY: pushimage
pushimage: ## Push Harbor images to specific registry server
	# For example: make pushimage -e DEVFLAG=false REGISTRYUSER=admin \
	#  				REGISTRYPASSWORD=***** \
	#  				REGISTRYSERVER=reg-bj.goharbor.io/ \
	#  				REGISTRYPROJECTNAME=harborrelease
	#  	note**: need add "/" on end of REGISTRYSERVER. If not setting \
	#  			this value will push images directly to dockerhub.
	#  			 make pushimage -e DEVFLAG=false REGISTRYUSER=goharbor \
	#  				REGISTRYPASSWORD=***** \
	#  				REGISTRYPROJECTNAME=goharbor
	@echo "pushing harbor images ..."
	@$(DOCKERTAG) $(DOCKER_IMAGE_NAME_PREPARE):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKER_IMAGE_NAME_PREPARE):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKER_IMAGE_NAME_PREPARE):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKER_IMAGE_NAME_PREPARE):$(VERSIONTAG)

	@$(DOCKERTAG) $(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG)

	@$(DOCKERTAG) $(DOCKERIMAGENAME_CORE):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKERIMAGENAME_CORE):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKERIMAGENAME_CORE):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKERIMAGENAME_CORE):$(VERSIONTAG)

	@$(DOCKERTAG) $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG)

	@$(DOCKERTAG) $(DOCKERIMAGENAME_LOG):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKERIMAGENAME_LOG):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKERIMAGENAME_LOG):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKERIMAGENAME_LOG):$(VERSIONTAG)

	@$(DOCKERTAG) $(DOCKERIMAGENAME_DB):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKERIMAGENAME_DB):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKERIMAGENAME_DB):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKERIMAGENAME_DB):$(VERSIONTAG)


##@ Utility

.PHONY: prepare
prepare: update_prepare_version ## Prepare environment for starting a Harbor instance
	@echo "preparing..."
	@if [ -n "$(GEN_TLS)" ] ; then \
		$(DOCKERCMD) run --rm -v /:/hostfs:z $(IMAGENAMESPACE)/prepare:$(VERSIONTAG) gencert -p /etc/harbor/tls/internal; \
	fi
	@$(MAKEPATH)/$(PREPARECMD) $(PREPARECMD_PARA)

.PHONY: update_prepare_version
update_prepare_version:
	@echo "substitute the prepare version tag in prepare file..."
	@$(SEDCMDI) -e 's/goharbor\/prepare:.*[[:space:]]\+/goharbor\/prepare:$(VERSIONTAG) prepare /' $(MAKEPATH)/prepare ;

# $1 the name of the docker image
# $2 the tag of the docker image
# $3 the command to build the docker image
define prepare_docker_image
	@if [ "$(shell ${DOCKERIMAGES} -q $(1):$(2) 2> /dev/null)" == "" ]; then \
		$(3) && echo "build $(1):$(2) successfully" || (echo "build $(1):$(2) failed" && exit 1) ; \
	fi
endef

.PHONY: pull_base_docker
pull_base_docker: ## Pull base docker image
	@for name in $(BUILDBASETARGET); do \
		echo $$name ; \
		$(DOCKERPULL) $(BASEIMAGENAMESPACE)/harbor-$$name-base:$(BASEIMAGETAG) ; \
	done

.PHONY: swagger_client
swagger_client: ## Generate swagger client
	@echo "Generate swagger client"
	wget https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli/4.3.1/openapi-generator-cli-4.3.1.jar -O openapi-generator-cli.jar
	rm -rf harborclient
	mkdir  -p harborclient/harbor_v2_swagger_client
	java -jar openapi-generator-cli.jar generate -i api/v2.0/swagger.yaml -g python -o harborclient/harbor_v2_swagger_client --package-name v2_swagger_client
	cd harborclient/harbor_v2_swagger_client; python ./setup.py install
	pip install docker -q
	pip freeze


# lint swagger doc
SPECTRAL_IMAGENAME=$(IMAGENAMESPACE)/spectral
SPECTRAL_VERSION=v6.14.2
SPECTRAL_IMAGE_BUILD_CMD=${DOCKERBUILD} -f ${TOOLSPATH}/spectral/Dockerfile --build-arg NODE=${NODEBUILDIMAGE} --build-arg SPECTRAL_VERSION=${SPECTRAL_VERSION} -t ${SPECTRAL_IMAGENAME}:$(SPECTRAL_VERSION) .
SPECTRAL=$(RUNCONTAINER) $(SPECTRAL_IMAGENAME):$(SPECTRAL_VERSION)

.PHONY: lint_apis
lint_apis: ## Lint Swagger API
	$(call prepare_docker_image,${SPECTRAL_IMAGENAME},${SPECTRAL_VERSION},${SPECTRAL_IMAGE_BUILD_CMD})
	$(SPECTRAL) lint ./api/v2.0/swagger.yaml

SWAGGER_IMAGENAME=$(IMAGENAMESPACE)/swagger
SWAGGER_VERSION=v0.31.0
SWAGGER=$(RUNCONTAINER) ${SWAGGER_IMAGENAME}:${SWAGGER_VERSION}
SWAGGER_GENERATE_SERVER=${SWAGGER} generate server --template-dir=$(TOOLSPATH)/swagger/templates --exclude-main --additional-initialism=CVE --additional-initialism=GC --additional-initialism=OIDC
SWAGGER_IMAGE_BUILD_CMD=${DOCKERBUILD} -f ${TOOLSPATH}/swagger/Dockerfile --build-arg GOLANG=${GOBUILDIMAGE} --build-arg SWAGGER_VERSION=${SWAGGER_VERSION} -t ${SWAGGER_IMAGENAME}:$(SWAGGER_VERSION) .

# $1 the path of swagger spec
# $2 the path of base directory for generating the files
# $3 the name of the application
define swagger_generate_server
	@echo "generate all the files for API from $(1)"
	@rm -rf $(2)/{models,restapi}
	@mkdir -p $(2)
	@$(SWAGGER_GENERATE_SERVER) -f $(1) -A $(3) --target $(2)
endef

.PHONY: gen_apis
gen_apis: lint_apis ## Generate Swagger API
	$(call prepare_docker_image,${SWAGGER_IMAGENAME},${SWAGGER_VERSION},${SWAGGER_IMAGE_BUILD_CMD})
	$(call swagger_generate_server,api/v2.0/swagger.yaml,src/server/v2.0,harbor)


MOCKERY_IMAGENAME=$(IMAGENAMESPACE)/mockery
MOCKERY_VERSION=v2.51.0
MOCKERY=$(RUNCONTAINER)/src ${MOCKERY_IMAGENAME}:${MOCKERY_VERSION}
MOCKERY_IMAGE_BUILD_CMD=${DOCKERBUILD} -f ${TOOLSPATH}/mockery/Dockerfile --build-arg GOLANG=${GOBUILDIMAGE} --build-arg MOCKERY_VERSION=${MOCKERY_VERSION} -t ${MOCKERY_IMAGENAME}:$(MOCKERY_VERSION) .

.PHONY: gen_mocks
gen_mocks: ## Generate Mocks
	$(call prepare_docker_image,${MOCKERY_IMAGENAME},${MOCKERY_VERSION},${MOCKERY_IMAGE_BUILD_CMD})
	${MOCKERY} mockery

.PHONY: mocks_check
mocks_check: gen_mocks ## Check generated mocks
	@echo checking mocks...
	@res=$$(git status -s src/ | awk '{ printf("%s\n", $$2) }' | egrep .*.go); \
	if [ -n "$${res}" ]; then \
		echo mocks of the interface are out of date... ; \
		echo "$${res}"; \
		exit 1; \
	fi

export VERSIONS_FOR_PREPARE
versions_prepare:
	@echo "$$VERSIONS_FOR_PREPARE" > $(MAKE_PREPARE_PATH)/$(PREPARE_VERSION_NAME)

check_environment:
	@$(MAKEPATH)/$(CHECKENVCMD)

gen_tls:
	@$(DOCKERCMD) run --rm -v /:/hostfs:z $(IMAGENAMESPACE)/prepare:$(VERSIONTAG) gencert -p /etc/harbor/tls/internal

##@ Instance controls
start: ## Startup Harbor instance
	@echo "loading harbor images..."
	@$(DOCKERCOMPOSECMD) $(DOCKERCOMPOSE_FILE_OPT) up -d
	@echo "Start complete. You can visit harbor now."

down: ## Shutdown Harbor instance
	@while [ -z "$$CONTINUE" ]; do \
        read -r -p "Type anything but Y or y to exit. [Y/N]: " CONTINUE; \
    done ; \
    [ $$CONTINUE = "y" ] || [ $$CONTINUE = "Y" ] || (echo "Exiting."; exit 1;)
	@echo "stoping harbor instance..."
	@$(DOCKERCOMPOSECMD) $(DOCKERCOMPOSE_FILE_OPT) down -v
	@echo "Done."

restart: down prepare start ## Restart Harbor instance

##@ Code Quality

.PHONY: go_check
go_check: gen_apis mocks_check misspell commentfmt lint ## Run all of the following golang checks

.PHONY: commentfmt
commentfmt: ## Check comment formatting
	@echo checking comment format...
	@res=$$(find . -type d \( -path ./tests \) -prune -o -name '*.go' -print | xargs egrep '(^|\s)\/\/(\S)'|grep -v '//go:generate'); \
	if [ -n "$${res}" ]; then \
		echo checking comment format fail.. ; \
		echo missing whitespace between // and comment body;\
		echo "$${res}"; \
		exit 1; \
	fi

.PHONY: misspell
misspell: # Check for misspellings using misspell utility
	@echo checking misspell...
	@find . -type d \( -path ./tests \) -prune -o -name '*.go' -print | xargs misspell -error

# golangci-lint binary installation or refer to https://golangci-lint.run/usage/install/#local-installation
# curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
GOLANGCI_LINT := $(shell go env GOPATH)/bin/golangci-lint
.PHONY: lint
lint: ## Lint all files using golangci-lint
	@echo checking lint
	@echo $(GOLANGCI_LINT)
	@cd ./src/; $(GOLANGCI_LINT) cache clean; $(GOLANGCI_LINT) -v run ./... --timeout=10m;

# go install golang.org/x/vuln/cmd/govulncheck@latest
GOVULNCHECK := $(shell go env GOPATH)/bin/govulncheck
.PHONY: govulncheck
govulncheck: ## Check for golang vulnerabilities using govulncheck
	@echo golang vulnerability check
	@cd ./src/; $(GOVULNCHECK) ./...;

##@ Clean Up

.PHONY: cleanall
cleanall: cleanbinary cleanimage cleanbaseimage cleandockercomposefile cleanconfig cleanpackage ## Remove binary, Harbor images, specific version docker-compose file, specific version tag and online/offline install package

.PHONY: cleanbinary
cleanbinary: ## Remove core and jobservice binary
	@echo "cleaning binary..."
	if [ -f $(CORE_BINARYPATH)/$(CORE_BINARYNAME) ] ; then rm $(CORE_BINARYPATH)/$(CORE_BINARYNAME) ; fi
	if [ -f $(JOBSERVICEBINARYPATH)/$(JOBSERVICEBINARYNAME) ] ; then rm $(JOBSERVICEBINARYPATH)/$(JOBSERVICEBINARYNAME) ; fi
	if [ -f $(REGISTRYCTLBINARYPATH)/$(REGISTRYCTLBINARYNAME) ] ; then rm $(REGISTRYCTLBINARYPATH)/$(REGISTRYCTLBINARYNAME) ; fi
	rm -rf make/photon/*/binary/

.PHONY: cleanimage
cleanimage: ## Remove Harbor images
	@echo "cleaning image for photon..."
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_CORE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_DB):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_LOG):$(VERSIONTAG)

.PHONY: cleanbaseimage
cleanbaseimage: ## Remove base image of Harbor images
	@echo "cleaning base image for photon..."
	@for name in $(BUILDBASETARGET); do \
		$(DOCKERRMIMAGE) -f $(BASEIMAGENAMESPACE)/harbor-$$name-base:$(BASEIMAGETAG) ; \
	done

.PHONY: cleandockercomposefile
cleandockercomposefile: ## Remove specific version docker-compose
	@echo "cleaning docker-compose files in $(DOCKERCOMPOSEFILEPATH)"
	@find $(DOCKERCOMPOSEFILEPATH) -maxdepth 1 -name "docker-compose*.yml" -exec rm -f {} \;
	@find $(DOCKERCOMPOSEFILEPATH) -maxdepth 1 -name "docker-compose*.yml-e" -exec rm -f {} \;

.PHONY: cleanpackage
cleanpackage: ## Remove online and offline install package
	@echo "cleaning harbor install package"
	@if [ -d $(BUILDPATH)/harbor ] ; then rm -rf $(BUILDPATH)/harbor ; fi
	@if [ -f $(BUILDPATH)/harbor-online-installer-$(VERSIONTAG).tgz ] ; \
	then rm $(BUILDPATH)/harbor-online-installer-$(VERSIONTAG).tgz ; fi
	@if [ -f $(BUILDPATH)/harbor-offline-installer-$(VERSIONTAG).tgz ] ; \
	then rm $(BUILDPATH)/harbor-offline-installer-$(VERSIONTAG).tgz ; fi

.PHONY: cleanconfig
cleanconfig: ## Remove temporary config files
	@echo "clean generated config files"
	rm -f $(BUILDPATH)/make/photon/prepare/versions
	rm -f $(BUILDPATH)/UIVERSION
	rm -rf $(BUILDPATH)/make/common
	rm -rf $(BUILDPATH)/harborclient
	rm -rf $(BUILDPATH)/src/portal/dist
	rm -rf $(BUILDPATH)/src/portal/lib/dist
	rm -f $(BUILDPATH)/src/portal/proxy.config.json

