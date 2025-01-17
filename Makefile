# Makefile for Harbor project
#
# Targets:
#
# all:			prepare env, compile binaries, build images and install images
# prepare: 		prepare env
# compile: 		compile core and jobservice code
#
# compile_golangimage:
#			compile from golang image
#			for example: make compile_golangimage -e GOBUILDIMAGE= \
#							goharbor/golang:1.23.4
# compile_core, compile_jobservice: compile specific binary
#
# build:	build Harbor docker images from photon baseimage
#
# install:		include compile binarys, build images, prepare specific \
#				version composefile and startup Harbor instance
#
# start:		startup Harbor instance
#
# down:			shutdown Harbor instance
#
# package_online:
#				prepare online install package
#			for example: make package_online -e DEVFLAG=false\
#							REGISTRYSERVER=reg-bj.goharbor.io \
#							REGISTRYPROJECTNAME=harborrelease
#
# package_offline:
#				prepare offline install package
#
# pushimage:	push Harbor images to specific registry server
#			for example: make pushimage -e DEVFLAG=false REGISTRYUSER=admin \
#							REGISTRYPASSWORD=***** \
#							REGISTRYSERVER=reg-bj.goharbor.io/ \
#							REGISTRYPROJECTNAME=harborrelease
#				note**: need add "/" on end of REGISTRYSERVER. If not setting \
#						this value will push images directly to dockerhub.
#						 make pushimage -e DEVFLAG=false REGISTRYUSER=goharbor \
#							REGISTRYPASSWORD=***** \
#							REGISTRYPROJECTNAME=goharbor
#
# clean:        remove binary, Harbor images, specific version docker-compose \
#               file, specific version tag and online/offline install package
# cleanbinary:	remove core and jobservice binary
# cleanimage: 	remove Harbor images
# cleandockercomposefile:
#				remove specific version docker-compose
# cleanversiontag:
#				cleanpackageremove specific version tag
# cleanpackage: remove online/offline install package
#
# other example:
#	clean specific version binarys and images:
#				make clean -e VERSIONTAG=[TAG]
#				note**: If commit new code to github, the git commit TAG will \
#				change. Better use this commond clean previous images and \
#				files with specific TAG.
#   By default DEVFLAG=true, if you want to release new version of Harbor, \
#		should setting the flag to false.
#				make XXXX -e DEVFLAG=false

SHELL := /bin/bash
BUILDPATH=$(CURDIR)
MAKEPATH=$(BUILDPATH)/make
MAKEDEVPATH=$(MAKEPATH)/dev
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
NOTARYFLAG=false
CLAIRFLAG=false
HTTPPROXY=
BUILDBIN=true
MIGRATORFLAG=false
NPM_REGISTRY=https://registry.npmjs.org
# enable/disable chart repo supporting
CHARTFLAG=false

# version prepare
# for docker image tag
VERSIONTAG=dev
# for base docker image tag
BASEIMAGETAG=dev
# for harbor package name
PKGVERSIONTAG=dev
# input true/false only
PULL_BASE_FROM_DOCKERHUB=true

PREPARE_VERSION_NAME=versions

#versions
REGISTRYVERSION=v2.7.1-patch-2819-2553
NOTARYVERSION=v0.6.1
CLAIRVERSION=v2.1.6
NOTARYMIGRATEVERSION=v3.5.4
CLAIRADAPTERVERSION=v1.0.1

# version of chartmuseum
CHARTMUSEUMVERSION=v0.12.0

# version of registry for pulling the source code
REGISTRY_SRC_TAG=v2.7.1

define VERSIONS_FOR_PREPARE
VERSION_TAG: $(VERSIONTAG)
REGISTRY_VERSION: $(REGISTRYVERSION)
NOTARY_VERSION: $(NOTARYVERSION)
CLAIR_VERSION: $(CLAIRVERSION)
CLAIR_ADAPTER_VERSION: $(CLAIRADAPTERVERSION)
CHARTMUSEUM_VERSION: $(CHARTMUSEUMVERSION)
endef

# docker parameters
DOCKERCMD=$(shell which docker)
DOCKERBUILD=$(DOCKERCMD) build
DOCKERRMIMAGE=$(DOCKERCMD) rmi
DOCKERPULL=$(DOCKERCMD) pull
DOCKERIMASES=$(DOCKERCMD) images
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
GOVERSION=1.23.4
GOBUILDIMAGE=goharbor/golang:$(GOVERSION)
PUSHGOIMAGE=false
GOBUILDPATH=/harbor

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
GOIMAGEBUILDCMD=/usr/local/go/bin/go build -mod vendor
GOIMAGEBUILD_COMMON=$(GOIMAGEBUILDCMD) $(GOFLAGS) ${GOTAGS} ${GOLDFLAGS}
GOIMAGEBUILD_CORE=$(GOIMAGEBUILDCMD) $(GOFLAGS) ${GOTAGS} --ldflags "-w -s $(CORE_LDFLAGS)"

GOBUILDPATH_CORE=$(GOBUILDPATH)/src/core
GOBUILDPATH_JOBSERVICE=$(GOBUILDPATH)/src/jobservice
GOBUILDPATH_REGISTRYCTL=$(GOBUILDPATH)/src/registryctl
GOBUILDPATH_MIGRATEPATCH=$(GOBUILDPATH)/src/cmd/migrate-patch
GOBUILDMAKEPATH=$(GOBUILDPATH)/make
GOBUILDMAKEPATH_CORE=$(GOBUILDMAKEPATH)/photon/core
GOBUILDMAKEPATH_JOBSERVICE=$(GOBUILDMAKEPATH)/photon/jobservice
GOBUILDMAKEPATH_REGISTRYCTL=$(GOBUILDMAKEPATH)/photon/registryctl
GOBUILDMAKEPATH_NOTARY=$(GOBUILDMAKEPATH)/photon/notary

# binary
CORE_BINARYPATH=$(MAKEDEVPATH)/core
CORE_BINARYNAME=harbor_core
JOBSERVICEBINARYPATH=$(MAKEDEVPATH)/jobservice
JOBSERVICEBINARYNAME=harbor_jobservice
REGISTRYCTLBINARYPATH=$(MAKEDEVPATH)/registryctl
REGISTRYCTLBINARYNAME=harbor_registryctl
MIGRATEPATCHBINARYNAME=migrate-patch

# configfile
CONFIGPATH=$(MAKEPATH)
INSIDE_CONFIGPATH=/compose_location
CONFIGFILE=harbor.yml

# prepare parameters
PREPAREPATH=$(TOOLSPATH)
PREPARECMD=prepare
PREPARECMD_PARA=--conf $(INSIDE_CONFIGPATH)/$(CONFIGFILE)
ifeq ($(NOTARYFLAG), true)
	PREPARECMD_PARA+= --with-notary
endif
ifeq ($(CLAIRFLAG), true)
	PREPARECMD_PARA+= --with-clair
endif
# append chartmuseum parameters if set
ifeq ($(CHARTFLAG), true)
    PREPARECMD_PARA+= --with-chartmuseum
endif

# makefile
MAKEFILEPATH_PHOTON=$(MAKEPATH)/photon

# common dockerfile
DOCKERFILEPATH_COMMON=$(MAKEPATH)/common

# docker image name
DOCKER_IMAGE_NAME_PREPARE=goharbor/prepare
DOCKERIMAGENAME_PORTAL=goharbor/harbor-portal
DOCKERIMAGENAME_CORE=goharbor/harbor-core
DOCKERIMAGENAME_JOBSERVICE=goharbor/harbor-jobservice
DOCKERIMAGENAME_LOG=goharbor/harbor-log
DOCKERIMAGENAME_DB=goharbor/harbor-db
DOCKERIMAGENAME_CHART_SERVER=goharbor/chartmuseum-photon
DOCKERIMAGENAME_REGCTL=goharbor/harbor-registryctl

# docker-compose files
DOCKERCOMPOSEFILEPATH=$(MAKEPATH)
DOCKERCOMPOSETPLFILENAME=docker-compose.tpl
DOCKERCOMPOSEFILENAME=docker-compose.yml
DOCKERCOMPOSENOTARYTPLFILENAME=docker-compose.notary.tpl
DOCKERCOMPOSENOTARYFILENAME=docker-compose.notary.yml
DOCKERCOMPOSECLAIRTPLFILENAME=docker-compose.clair.tpl
DOCKERCOMPOSECLAIRFILENAME=docker-compose.clair.yml
DOCKERCOMPOSECHARTMUSEUMTPLFILENAME=docker-compose.chartmuseum.tpl
DOCKERCOMPOSECHARTMUSEUMFILENAME=docker-compose.chartmuseum.yml

SEDCMD=$(shell which sed)

# package
TARCMD=$(shell which tar)
ZIPCMD=$(shell which gzip)
DOCKERIMGFILE=harbor
HARBORPKG=harbor

# pushimage
PUSHSCRIPTPATH=$(MAKEPATH)
PUSHSCRIPTNAME=pushimage.sh
REGISTRYUSER=user
REGISTRYPASSWORD=default

# cmds
DOCKERSAVE_PARA=$(DOCKER_IMAGE_NAME_PREPARE):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_CORE):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_LOG):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_DB):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_REGCTL):$(VERSIONTAG) \
		goharbor/redis-photon:$(VERSIONTAG) \
		goharbor/nginx-photon:$(VERSIONTAG) \
		goharbor/registry-photon:$(VERSIONTAG)

PACKAGE_OFFLINE_PARA=-zcvf harbor-offline-installer-$(PKGVERSIONTAG).tgz \
					$(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tar.gz \
					$(HARBORPKG)/prepare \
					$(HARBORPKG)/LICENSE $(HARBORPKG)/install.sh \
					$(HARBORPKG)/common.sh \
					$(HARBORPKG)/harbor.yml

PACKAGE_ONLINE_PARA=-zcvf harbor-online-installer-$(PKGVERSIONTAG).tgz \
					$(HARBORPKG)/prepare \
					$(HARBORPKG)/LICENSE \
					$(HARBORPKG)/install.sh \
					$(HARBORPKG)/common.sh \
					$(HARBORPKG)/harbor.yml

DOCKERCOMPOSE_FILE_OPT=-f $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)

ifeq ($(NOTARYFLAG), true)
	DOCKERSAVE_PARA+= goharbor/notary-server-photon:$(VERSIONTAG) goharbor/notary-signer-photon:$(VERSIONTAG)
endif
ifeq ($(CLAIRFLAG), true)
	DOCKERSAVE_PARA+= goharbor/clair-photon:$(VERSIONTAG) goharbor/clair-adapter-photon:$(VERSIONTAG)
endif
ifeq ($(MIGRATORFLAG), true)
	DOCKERSAVE_PARA+= goharbor/harbor-migrator:$(VERSIONTAG)
endif
# append chartmuseum parameters if set
ifeq ($(CHARTFLAG), true)
	DOCKERSAVE_PARA+= $(DOCKERIMAGENAME_CHART_SERVER):$(VERSIONTAG)
endif

export VERSIONS_FOR_PREPARE
versions_prepare:
	@echo "$$VERSIONS_FOR_PREPARE" > $(MAKE_PREPARE_PATH)/$(PREPARE_VERSION_NAME)

check_environment:
	@$(MAKEPATH)/$(CHECKENVCMD)

compile_core:
	@echo "compiling binary for core (golang image)..."
	@echo $(GOBUILDPATH)
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATH) -w $(GOBUILDPATH_CORE) $(GOBUILDIMAGE) $(GOIMAGEBUILD_CORE) -o $(GOBUILDMAKEPATH_CORE)/$(CORE_BINARYNAME)
	@echo "Done."

compile_jobservice:
	@echo "compiling binary for jobservice (golang image)..."
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATH) -w $(GOBUILDPATH_JOBSERVICE) $(GOBUILDIMAGE) $(GOIMAGEBUILD_COMMON) -o $(GOBUILDMAKEPATH_JOBSERVICE)/$(JOBSERVICEBINARYNAME)
	@echo "Done."

compile_registryctl:
	@echo "compiling binary for harbor registry controller (golang image)..."
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATH) -w $(GOBUILDPATH_REGISTRYCTL) $(GOBUILDIMAGE) $(GOIMAGEBUILD_COMMON) -o $(GOBUILDMAKEPATH_REGISTRYCTL)/$(REGISTRYCTLBINARYNAME)
	@echo "Done."

compile_notary_migrate_patch:
	@echo "compiling binary for migrate patch (golang image)..."
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATH) -w $(GOBUILDPATH_MIGRATEPATCH) $(GOBUILDIMAGE) $(GOIMAGEBUILD_COMMON) -o $(GOBUILDMAKEPATH_NOTARY)/$(MIGRATEPATCHBINARYNAME)
	@echo "Done."

compile: check_environment versions_prepare compile_core compile_jobservice compile_registryctl compile_notary_migrate_patch

update_prepare_version:
	@echo "substitute the prepare version tag in prepare file..."
	@$(SEDCMD) -i -e 's/goharbor\/prepare:.*[[:space:]]\+/goharbor\/prepare:$(VERSIONTAG) /' $(MAKEPATH)/prepare ;

prepare: update_prepare_version
	@echo "preparing..."
	@$(MAKEPATH)/$(PREPARECMD) $(PREPARECMD_PARA)

build:
# PUSHBASEIMAGE should not be true if BUILD_BASE is not true
	@if [ "$(PULL_BASE_FROM_DOCKERHUB)" != "true" ] && [ "$(PULL_BASE_FROM_DOCKERHUB)" != "false" ] ; then \
		echo set PULL_BASE_FROM_DOCKERHUB to true or false.; exit 1; \
	fi
	make -f $(MAKEFILEPATH_PHOTON)/Makefile build -e DEVFLAG=$(DEVFLAG) -e GOBUILDIMAGE=$(GOBUILDIMAGE) \
	 -e REGISTRYVERSION=$(REGISTRYVERSION) -e REGISTRY_SRC_TAG=$(REGISTRY_SRC_TAG) -e NOTARYVERSION=$(NOTARYVERSION) -e NOTARYMIGRATEVERSION=$(NOTARYMIGRATEVERSION) \
	 -e CLAIRVERSION=$(CLAIRVERSION) -e CLAIRADAPTERVERSION=$(CLAIRADAPTERVERSION) -e VERSIONTAG=$(VERSIONTAG) \
	 -e BUILDBIN=$(BUILDBIN) \
	 -e CHARTMUSEUMVERSION=$(CHARTMUSEUMVERSION) -e DOCKERIMAGENAME_CHART_SERVER=$(DOCKERIMAGENAME_CHART_SERVER) \
	 -e NPM_REGISTRY=$(NPM_REGISTRY) -e BASEIMAGETAG=$(BASEIMAGETAG) -e PULL_BASE_FROM_DOCKERHUB=$(PULL_BASE_FROM_DOCKERHUB)

build_base_docker:
	if [ -n "$(REGISTRYUSER)" ] && [ -n "$(REGISTRYPASSWORD)" ] ; then \
		docker login -u $(REGISTRYUSER) -p $(REGISTRYPASSWORD) ; \
	else \
		echo "No docker credentials provided, please make sure enough priviledges to access docker hub!" ; \
	fi
	@for name in chartserver clair clair-adapter core db jobservice log nginx notary-server notary-signer portal prepare redis registry registryctl; do \
		echo $$name ; \
		if [ $$name == "db" ] ; then \
			cd $(MAKEFILEPATH_PHOTON)/$$name && $(MAKEFILEPATH_PHOTON)/$$name/rpm_builder.sh && cd - ; \
		fi; \
		$(DOCKERBUILD) --pull -f $(MAKEFILEPATH_PHOTON)/$$name/Dockerfile.base -t goharbor/harbor-$$name-base:$(BASEIMAGETAG) . && \
		$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) goharbor/harbor-$$name-base:$(BASEIMAGETAG) $(REGISTRYUSER) $(REGISTRYPASSWORD) || exit 1 ; \
	done
	@echo build_base_docker done

build_base_images:
	@echo build_base_images start
	@for name in chartserver clair clair-adapter core db jobservice log nginx notary-server notary-signer portal prepare redis registry registryctl; do \
		if [ $$name == "db" ] ; then \
			cd $(MAKEFILEPATH_PHOTON)/$$name && $(MAKEFILEPATH_PHOTON)/$$name/rpm_builder.sh && cd - ; \
		fi; \
		$(DOCKERBUILD) --pull -f $(MAKEFILEPATH_PHOTON)/$$name/Dockerfile.base -t goharbor/harbor-$$name-base:$(BASEIMAGETAG) .; \
	done
	@echo build_base_images done

install: compile build prepare start

package_online: update_prepare_version
	@echo "packing online package ..."
	@cp -r make $(HARBORPKG)
	@if [ -n "$(REGISTRYSERVER)" ] ; then \
		$(SEDCMD) -i -e 's/image\: goharbor/image\: $(REGISTRYSERVER)\/$(REGISTRYPROJECTNAME)/' \
		$(HARBORPKG)/docker-compose.yml ; \
	fi
	@cp LICENSE $(HARBORPKG)/LICENSE

	@$(TARCMD) $(PACKAGE_ONLINE_PARA)
	@rm -rf $(HARBORPKG)
	@echo "Done."

package_offline: update_prepare_version compile build

	@echo "packing offline package ..."
	@cp -r make $(HARBORPKG)
	@cp LICENSE $(HARBORPKG)/LICENSE

	@echo "saving harbor docker image"
	@$(DOCKERSAVE) $(DOCKERSAVE_PARA) > $(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tar
	@gzip $(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tar

	@$(TARCMD) $(PACKAGE_OFFLINE_PARA)
	@rm -rf $(HARBORPKG)
	@echo "Done."

build_golang:
	@echo "build goharbor/golang image"
	$(DOCKERBUILD) --build-arg GOVERSION=$(GOVERSION) -f $(MAKEPATH)/photon/golang/Dockerfile -t $(GOBUILDIMAGE) .
	@if [ "$(PUSHGOIMAGE)" = "true" ] ; then \
	echo "push goharbor/golang image"; \
	docker login -u $(REGISTRYUSER) -p $(REGISTRYPASSWORD) ; \
	docker push $(GOBUILDIMAGE); \
	fi; \

gosec:
	#go get github.com/securego/gosec/cmd/gosec
	#go get github.com/dghubble/sling
	@echo "run secure go scan ..."
	@if [ "$(GOSECRESULTS)" != "" ] ; then \
		$(GOPATH)/bin/gosec -fmt=json -out=$(GOSECRESULTS) -quiet ./... | true ; \
	else \
		$(GOPATH)/bin/gosec -fmt=json -out=harbor_gas_output.json -quiet ./... | true ; \
	fi

go_check: misspell golint govet gofmt commentfmt

gofmt:
	@echo checking gofmt...
	@res=$$(gofmt -d -e -s $$(find . -type d \( -path ./src/vendor -o -path ./tests \) -prune -o -name '*.go' -print)); \
	if [ -n "$${res}" ]; then \
		echo checking gofmt fail... ; \
		echo "$${res}"; \
		exit 1; \
	fi

commentfmt:
	@echo checking comment format...
	@res=$$(find . -type d \( -path ./src/vendor -o -path ./tests \) -prune -o -name '*.go' -print | xargs egrep '(^|\s)\/\/(\S)'); \
	if [ -n "$${res}" ]; then \
		echo checking comment format fail.. ; \
		echo missing whitespace between // and comment body;\
		echo "$${res}"; \
		exit 1; \
	fi

misspell:
	@echo checking misspell...
	@find . -type d \( -path ./src/vendor -o -path ./tests \) -prune -o -name '*.go' -print | xargs misspell -error

golint:
	@echo checking golint...
	@cd ./src;go list ./... | grep -v -E 'vendor|test' | xargs -L1 golint

govet:
	@echo checking govet...
	@cd ./src;go list ./... | grep -v -E 'vendor|test' | xargs -L1 go vet

pushimage:
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

start:
	@echo "loading harbor images..."
	@$(DOCKERCOMPOSECMD) $(DOCKERCOMPOSE_FILE_OPT) up -d
	@echo "Start complete. You can visit harbor now."

down:
	@while [ -z "$$CONTINUE" ]; do \
        read -r -p "Type anything but Y or y to exit. [Y/N]: " CONTINUE; \
    done ; \
    [ $$CONTINUE = "y" ] || [ $$CONTINUE = "Y" ] || (echo "Exiting."; exit 1;)
	@echo "stoping harbor instance..."
	@$(DOCKERCOMPOSECMD) $(DOCKERCOMPOSE_FILE_OPT) down -v
	@echo "Done."

swagger_client:
	@echo "Generate swagger client"
	wget https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli/4.3.1/openapi-generator-cli-4.3.1.jar -O openapi-generator-cli.jar
	rm -rf harborclient
	mkdir harborclient
	java -jar openapi-generator-cli.jar generate -i api/harbor/swagger.yaml -g python -o harborclient --package-name swagger_client
	cd harborclient; python ./setup.py install
	pip install docker -q
	pip freeze

cleanbinary:
	@echo "cleaning binary..."
	@if [ -f $(CORE_BINARYPATH)/$(CORE_BINARYNAME) ] ; then rm $(CORE_BINARYPATH)/$(CORE_BINARYNAME) ; fi
	@if [ -f $(JOBSERVICEBINARYPATH)/$(JOBSERVICEBINARYNAME) ] ; then rm $(JOBSERVICEBINARYPATH)/$(JOBSERVICEBINARYNAME) ; fi

cleanimage:
	@echo "cleaning image for photon..."
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_CORE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_DB):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_LOG):$(VERSIONTAG)

cleandockercomposefile:
	@echo "cleaning docker-compose files in $(DOCKERCOMPOSEFILEPATH)"
	@find $(DOCKERCOMPOSEFILEPATH) -maxdepth 1 -name "docker-compose*.yml" -exec rm -f {} \;
	@find $(DOCKERCOMPOSEFILEPATH) -maxdepth 1 -name "docker-compose*.yml-e" -exec rm -f {} \;

cleanpackage:
	@echo "cleaning harbor install package"
	@if [ -d $(BUILDPATH)/harbor ] ; then rm -rf $(BUILDPATH)/harbor ; fi
	@if [ -f $(BUILDPATH)/harbor-online-installer-$(VERSIONTAG).tgz ] ; \
	then rm $(BUILDPATH)/harbor-online-installer-$(VERSIONTAG).tgz ; fi
	@if [ -f $(BUILDPATH)/harbor-offline-installer-$(VERSIONTAG).tgz ] ; \
	then rm $(BUILDPATH)/harbor-offline-installer-$(VERSIONTAG).tgz ; fi

.PHONY: cleanall
cleanall: cleanbinary cleanimage cleandockercomposefile cleanpackage

clean:
	@echo "  make cleanall:		remove binary, Harbor images, specific version docker-compose"
	@echo "		file, specific version tag, online and offline install package"
	@echo "  make cleanbinary:		remove core and jobservice binary"
	@echo "  make cleanimage:		remove Harbor images"
	@echo "  make cleandockercomposefile:	remove specific version docker-compose"
	@echo "  make cleanpackage:		remove online and offline install package"

all: install
