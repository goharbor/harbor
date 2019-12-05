# Makefile for a harbor project
#
# Targets:
#
# build: 	build harbor photon images
# clean:	clean core and jobservice harbor images

# common
SHELL := /bin/bash
BUILDPATH=$(CURDIR)
MAKEPATH=$(BUILDPATH)/make
SRCPATH=./src
TOOLSPATH=$(CURDIR)/tools
SEDCMD=$(shell which sed)
WGET=$(shell which wget)

# docker parameters
DOCKERCMD=$(shell which docker)
DOCKERBUILD=$(DOCKERCMD) build
DOCKERRMIMAGE=$(DOCKERCMD) rmi
DOCKERIMASES=$(DOCKERCMD) images

# binary
CORE_SOURCECODE=$(SRCPATH)/core
CORE_BINARYNAME=harbor_core
JOBSERVICESOURCECODE=$(SRCPATH)/jobservice
JOBSERVICEBINARYNAME=harbor_jobservice

# photon dockerfile
DOCKERFILEPATH=$(MAKEPATH)/photon

DOCKERFILEPATH_PREPARE=$(DOCKERFILEPATH)/prepare
DOCKERFILENAME_PREPARE=Dockerfile
DOCKERIMAGENAME_PREPARE=goharbor/prepare

DOCKERFILEPATH_PORTAL=$(DOCKERFILEPATH)/portal
DOCKERFILENAME_PORTAL=Dockerfile
DOCKERIMAGENAME_PORTAL=goharbor/harbor-portal

DOCKERFILEPATH_CORE=$(DOCKERFILEPATH)/core
DOCKERFILENAME_CORE=Dockerfile
DOCKERIMAGENAME_CORE=goharbor/harbor-core

DOCKERFILEPATH_JOBSERVICE=$(DOCKERFILEPATH)/jobservice
DOCKERFILENAME_JOBSERVICE=Dockerfile
DOCKERIMAGENAME_JOBSERVICE=goharbor/harbor-jobservice

DOCKERFILEPATH_LOG=$(DOCKERFILEPATH)/log
DOCKERFILENAME_LOG=Dockerfile
DOCKERIMAGENAME_LOG=goharbor/harbor-log

DOCKERFILEPATH_DB=$(DOCKERFILEPATH)/db
DOCKERFILENAME_DB=Dockerfile
DOCKERIMAGENAME_DB=goharbor/harbor-db

DOCKERFILEPATH_POSTGRESQL=$(DOCKERFILEPATH)/postgresql
DOCKERFILENAME_POSTGRESQL=Dockerfile
DOCKERIMAGENAME_POSTGRESQL=goharbor/postgresql-photon

DOCKERFILEPATH_CLAIR=$(DOCKERFILEPATH)/clair
DOCKERFILENAME_CLAIR=Dockerfile
DOCKERIMAGENAME_CLAIR=goharbor/clair-photon

DOCKERFILEPATH_CLAIR_ADAPTER=$(DOCKERFILEPATH)/clair-adapter
DOCKERFILENAME_CLAIR_ADAPTER=Dockerfile
DOCKERIMAGENAME_CLAIR_ADAPTER=goharbor/clair-adapter-photon

DOCKERFILEPATH_NGINX=$(DOCKERFILEPATH)/nginx
DOCKERFILENAME_NGINX=Dockerfile
DOCKERIMAGENAME_NGINX=goharbor/nginx-photon

DOCKERFILEPATH_REG=$(DOCKERFILEPATH)/registry
DOCKERFILENAME_REG=Dockerfile
DOCKERIMAGENAME_REG=goharbor/registry-photon

DOCKERFILEPATH_REGISTRYCTL=$(DOCKERFILEPATH)/registryctl
DOCKERFILENAME_REGISTRYCTL=Dockerfile
DOCKERIMAGENAME_REGISTRYCTL=goharbor/harbor-registryctl

DOCKERFILEPATH_NOTARY=$(DOCKERFILEPATH)/notary
DOCKERFILEPATH_NOTARYSERVER=$(DOCKERFILEPATH)/notary-server
DOCKERFILENAME_NOTARYSIGNER=Dockerfile
DOCKERIMAGENAME_NOTARYSIGNER=goharbor/notary-signer-photon
DOCKERFILEPATH_NOTARYSIGNER=$(DOCKERFILEPATH)/notary-signer
DOCKERFILENAME_NOTARYSERVER=Dockerfile
DOCKERIMAGENAME_NOTARYSERVER=goharbor/notary-server-photon

DOCKERFILEPATH_REDIS=$(DOCKERFILEPATH)/redis
DOCKERFILENAME_REDIS=Dockerfile
DOCKERIMAGENAME_REDIS=goharbor/redis-photon

DOCKERFILEPATH_MIGRATOR=$(TOOLSPATH)/migration
DOCKERFILENAME_MIGRATOR=Dockerfile
DOCKERIMAGENAME_MIGRATOR=goharbor/harbor-migrator

# for chart server (chartmuseum)
DOCKERFILEPATH_CHART_SERVER=$(DOCKERFILEPATH)/chartserver
DOCKERFILENAME_CHART_SERVER=Dockerfile
CHART_SERVER_CODE_BASE=https://github.com/helm/chartmuseum.git
CHART_SERVER_MAIN_PATH=cmd/chartmuseum
CHART_SERVER_BIN_NAME=chartm

_build_prepare:
	@echo "building prepare container for photon..."
	@$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) -f $(DOCKERFILEPATH_PREPARE)/$(DOCKERFILENAME_PREPARE) -t $(DOCKERIMAGENAME_PREPARE):$(VERSIONTAG) .
	@echo "Done."

_build_db:
	@echo "building db container for photon..."
	@$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) -f $(DOCKERFILEPATH_DB)/$(DOCKERFILENAME_DB) -t $(DOCKERIMAGENAME_DB):$(VERSIONTAG) .
	@echo "Done."

_build_portal:
	@echo "building portal container for photon..."
	$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg npm_registry=$(NPM_REGISTRY) -f $(DOCKERFILEPATH_PORTAL)/$(DOCKERFILENAME_PORTAL) -t $(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG) .
	@echo "Done."

_build_core:
	@echo "building core container for photon..."
	@$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) -f $(DOCKERFILEPATH_CORE)/$(DOCKERFILENAME_CORE) -t $(DOCKERIMAGENAME_CORE):$(VERSIONTAG) .
	@echo "Done."

_build_jobservice:
	@echo "building jobservice container for photon..."
	@$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) -f $(DOCKERFILEPATH_JOBSERVICE)/$(DOCKERFILENAME_JOBSERVICE) -t $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) .
	@echo "Done."

_build_log:
	@echo "building log container for photon..."
	$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) -f $(DOCKERFILEPATH_LOG)/$(DOCKERFILENAME_LOG) -t $(DOCKERIMAGENAME_LOG):$(VERSIONTAG) .
	@echo "Done."

_build_clair:
	@if [ "$(CLAIRFLAG)" = "true" ] ; then \
		if [ "$(BUILDBIN)" != "true" ] ; then \
			rm -rf $(DOCKERFILEPATH_CLAIR)/binary && mkdir -p $(DOCKERFILEPATH_CLAIR)/binary && \
			$(call _get_binary, https://storage.googleapis.com/harbor-builds/bin/clair/release2.0-$(CLAIRVERSION)/clair, $(DOCKERFILEPATH_CLAIR)/binary/clair); \
		else \
			cd $(DOCKERFILEPATH_CLAIR) && $(DOCKERFILEPATH_CLAIR)/builder $(CLAIRVERSION) && cd - ; \
		fi ; \
		echo "building clair container for photon..." ; \
		$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) -f $(DOCKERFILEPATH_CLAIR)/$(DOCKERFILENAME_CLAIR) -t $(DOCKERIMAGENAME_CLAIR):$(CLAIRVERSION)-$(VERSIONTAG) . ; \
		rm -rf $(DOCKERFILEPATH_CLAIR)/binary; \
		echo "Done." ; \
	fi

_build_clair_adapter:
	@if [ "$(CLAIRFLAG)" = "true" ] ; then \
		if [ "$(BUILDBIN)" != "true" ] ; then \
			rm -rf $(DOCKERFILEPATH_CLAIR_ADAPTER)/binary && mkdir -p $(DOCKERFILEPATH_CLAIR_ADAPTER)/binary && \
			$(call _extract_archive, https://github.com/goharbor/harbor-scanner-clair/releases/download/$(CLAIRADAPTERVERSION)/harbor-scanner-clair_$(CLAIRADAPTERVERSION:v%=%)_Linux_x86_64.tar.gz, $(DOCKERFILEPATH_CLAIR_ADAPTER)/binary/) && \
			mv $(DOCKERFILEPATH_CLAIR_ADAPTER)/binary/scanner-clair $(DOCKERFILEPATH_CLAIR_ADAPTER)/binary/harbor-scanner-clair; \
		else \
			cd $(DOCKERFILEPATH_CLAIR_ADAPTER) && $(DOCKERFILEPATH_CLAIR_ADAPTER)/builder $(CLAIRADAPTERVERSION) && cd - ; \
		fi ; \
		echo "building clair adapter container for photon..." ; \
		$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) -f $(DOCKERFILEPATH_CLAIR_ADAPTER)/$(DOCKERFILENAME_CLAIR_ADAPTER) -t $(DOCKERIMAGENAME_CLAIR_ADAPTER):$(CLAIRADAPTERVERSION)-$(VERSIONTAG) . ; \
		rm -rf $(DOCKERFILEPATH_CLAIR_ADAPTER)/binary; \
		echo "Done." ; \
	fi

_build_chart_server:
	@if [ "$(CHARTFLAG)" = "true" ] ; then \
		if [ "$(BUILDBIN)" != "true" ] ; then \
			rm -rf $(DOCKERFILEPATH_CHART_SERVER)/binary && mkdir -p $(DOCKERFILEPATH_CHART_SERVER)/binary && \
			$(call _get_binary, https://storage.googleapis.com/harbor-builds/bin/chartmuseum/release-$(CHARTMUSEUMVERSION)/chartm, $(DOCKERFILEPATH_CHART_SERVER)/binary/chartm); \
		else \
			cd $(DOCKERFILEPATH_CHART_SERVER) && $(DOCKERFILEPATH_CHART_SERVER)/builder $(GOBUILDIMAGE) $(CHART_SERVER_CODE_BASE) $(CHARTMUSEUMVERSION) $(CHART_SERVER_MAIN_PATH) $(CHART_SERVER_BIN_NAME) && cd - ; \
		fi ; \
		echo "building chartmuseum container for photon..." ; \
		$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) -f $(DOCKERFILEPATH_CHART_SERVER)/$(DOCKERFILENAME_CHART_SERVER) -t $(DOCKERIMAGENAME_CHART_SERVER):$(CHARTMUSEUMVERSION)-$(VERSIONTAG) . ; \
		rm -rf $(DOCKERFILEPATH_CHART_SERVER)/binary; \
		echo "Done." ; \
	fi

_build_nginx:
	@echo "building nginx container for photon..."
	@$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) -f $(DOCKERFILEPATH_NGINX)/$(DOCKERFILENAME_NGINX) -t $(DOCKERIMAGENAME_NGINX):$(NGINXVERSION) .
	@echo "Done."

_build_notary:
	@if [ "$(NOTARYFLAG)" = "true" ] ; then \
		if [ "$(BUILDBIN)" != "true" ] ; then \
			rm -rf $(DOCKERFILEPATH_NOTARY)/binary && mkdir -p $(DOCKERFILEPATH_NOTARY)/binary && \
			$(call _get_binary, https://storage.googleapis.com/harbor-builds/bin/notary/release-$(NOTARYVERSION)/binary-bundle.tgz, $(DOCKERFILEPATH_NOTARY)/binary-bundle.tgz); \
			cd $(DOCKERFILEPATH_NOTARY) && tar -zvxf binary-bundle.tgz && cd - ; \
		else \
			cd $(DOCKERFILEPATH_NOTARY) && $(DOCKERFILEPATH_NOTARY)/builder $(NOTARYVERSION) $(NOTARYMIGRATEVERSION) && cd - ; \
		fi ; \
		echo "building notary container for photon..."; \
		chmod 655 $(DOCKERFILEPATH_NOTARY)/binary/notary-signer && $(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) -f $(DOCKERFILEPATH_NOTARYSIGNER)/$(DOCKERFILENAME_NOTARYSIGNER) -t $(DOCKERIMAGENAME_NOTARYSIGNER):$(NOTARYVERSION)-$(VERSIONTAG) . ; \
		chmod 655 $(DOCKERFILEPATH_NOTARY)/binary/notary-server && $(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) -f $(DOCKERFILEPATH_NOTARYSERVER)/$(DOCKERFILENAME_NOTARYSERVER) -t $(DOCKERIMAGENAME_NOTARYSERVER):$(NOTARYVERSION)-$(VERSIONTAG) . ; \
		rm -rf $(DOCKERFILEPATH_NOTARY)/binary; \
		echo "Done."; \
	fi

_build_registry:
	@if [ "$(BUILDBIN)" != "true" ] ; then \
		rm -rf $(DOCKERFILEPATH_REG)/binary && mkdir -p $(DOCKERFILEPATH_REG)/binary && \
		$(call _get_binary, https://storage.googleapis.com/harbor-builds/bin/registry/release-$(REGISTRYVERSION)/registry, $(DOCKERFILEPATH_REG)/binary/registry); \
	else \
		cd $(DOCKERFILEPATH_REG) && $(DOCKERFILEPATH_REG)/builder $(REGISTRY_SRC_TAG) && cd - ; \
	fi
	@echo "building registry container for photon..."
	@chmod 655 $(DOCKERFILEPATH_REG)/binary/registry && $(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) -f $(DOCKERFILEPATH_REG)/$(DOCKERFILENAME_REG) -t $(DOCKERIMAGENAME_REG):$(REGISTRYVERSION)-$(VERSIONTAG) .
	@echo "Done."

_build_registryctl:
	@echo "building registry controller for photon..."
	@$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) -f $(DOCKERFILEPATH_REGISTRYCTL)/$(DOCKERFILENAME_REGISTRYCTL) -t $(DOCKERIMAGENAME_REGISTRYCTL):$(VERSIONTAG) .
	@rm -rf $(DOCKERFILEPATH_REG)/binary
	@echo "Done."

_build_redis:
	@echo "building redis container for photon..."
	@$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) -f $(DOCKERFILEPATH_REDIS)/$(DOCKERFILENAME_REDIS) -t $(DOCKERIMAGENAME_REDIS):$(REDISVERSION) .
	@echo "Done."

_build_migrator:
	@if [ "$(MIGRATORFLAG)" = "true" ] ; then \
		echo "building db migrator container for photon..."; \
		cd $(DOCKERFILEPATH_MIGRATOR) && $(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) -f $(DOCKERFILEPATH_MIGRATOR)/$(DOCKERFILENAME_MIGRATOR) -t $(DOCKERIMAGENAME_MIGRATOR):$(MIGRATORVERSION) . ; \
		echo "Done."; \
	fi

define _extract_archive
	$(WGET) --timeout 30 --no-check-certificate -O- $1 | tar xvz -C $2
endef

define _get_binary
	$(WGET) --timeout 30 --no-check-certificate $1 -O $2
endef

build: _build_prepare _build_db _build_portal _build_core _build_jobservice _build_log _build_nginx _build_registry _build_registryctl _build_notary _build_clair _build_clair_adapter _build_redis _build_migrator _build_chart_server

cleanimage:
	@echo "cleaning image for photon..."
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_CORE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_LOG):$(VERSIONTAG)

.PHONY: clean
clean: cleanimage

