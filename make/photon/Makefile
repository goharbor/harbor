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
CURL=$(shell which curl)

# docker parameters
DOCKERCMD=$(shell which docker)
DOCKERBUILD=$(DOCKERCMD) build
DOCKERRMIMAGE=$(DOCKERCMD) rmi
DOCKERIMAGES=$(DOCKERCMD) images


TARGETARCHS=amd64 arm64

REGISTRYPROJECTNAME=goharbor
BASEIMAGENAMESPACE={REGISTRYPROJECTNAME}

# binary
CORE_SOURCECODE=$(SRCPATH)/core
CORE_BINARYNAME=harbor_core
JOBSERVICESOURCECODE=$(SRCPATH)/jobservice
JOBSERVICEBINARYNAME=harbor_jobservice

# photon dockerfile
DOCKERFILEPATH=$(MAKEPATH)/photon

DOCKERFILEPATH_PREPARE=$(DOCKERFILEPATH)/prepare
DOCKERFILENAME_PREPARE=Dockerfile
DOCKERIMAGENAME_PREPARE=$(REGISTRYPROJECTNAME)/prepare

DOCKERFILEPATH_PORTAL=$(DOCKERFILEPATH)/portal
DOCKERFILENAME_PORTAL=Dockerfile
DOCKERIMAGENAME_PORTAL=$(REGISTRYPROJECTNAME)/harbor-portal

DOCKERFILEPATH_CORE=$(DOCKERFILEPATH)/core
DOCKERFILENAME_CORE=Dockerfile
DOCKERIMAGENAME_CORE=$(REGISTRYPROJECTNAME)/harbor-core

DOCKERFILEPATH_JOBSERVICE=$(DOCKERFILEPATH)/jobservice
DOCKERFILENAME_JOBSERVICE=Dockerfile
DOCKERIMAGENAME_JOBSERVICE=$(REGISTRYPROJECTNAME)/harbor-jobservice

DOCKERFILEPATH_LOG=$(DOCKERFILEPATH)/log
DOCKERFILENAME_LOG=Dockerfile
DOCKERIMAGENAME_LOG=$(REGISTRYPROJECTNAME)/harbor-log

DOCKERFILEPATH_DB=$(DOCKERFILEPATH)/db
DOCKERFILENAME_DB=Dockerfile
DOCKERIMAGENAME_DB=$(REGISTRYPROJECTNAME)/harbor-db

DOCKERFILEPATH_POSTGRESQL=$(DOCKERFILEPATH)/postgresql
DOCKERFILENAME_POSTGRESQL=Dockerfile
DOCKERIMAGENAME_POSTGRESQL=$(REGISTRYPROJECTNAME)/postgresql-photon

DOCKERFILEPATH_TRIVY_ADAPTER=$(DOCKERFILEPATH)/trivy-adapter
DOCKERFILENAME_TRIVY_ADAPTER=Dockerfile
DOCKERIMAGENAME_TRIVY_ADAPTER=$(REGISTRYPROJECTNAME)/trivy-adapter-photon

DOCKERFILEPATH_NGINX=$(DOCKERFILEPATH)/nginx
DOCKERFILENAME_NGINX=Dockerfile
DOCKERIMAGENAME_NGINX=$(REGISTRYPROJECTNAME)/nginx-photon

DOCKERFILEPATH_REG=$(DOCKERFILEPATH)/registry
DOCKERFILENAME_REG=Dockerfile
DOCKERIMAGENAME_REG=$(REGISTRYPROJECTNAME)/registry-photon

DOCKERFILEPATH_REGISTRYCTL=$(DOCKERFILEPATH)/registryctl
DOCKERFILENAME_REGISTRYCTL=Dockerfile
DOCKERIMAGENAME_REGISTRYCTL=$(REGISTRYPROJECTNAME)/harbor-registryctl

DOCKERFILEPATH_NOTARY=$(DOCKERFILEPATH)/notary
DOCKERFILEPATH_NOTARYSERVER=$(DOCKERFILEPATH)/notary-server
DOCKERFILENAME_NOTARYSIGNER=Dockerfile
DOCKERIMAGENAME_NOTARYSIGNER=$(REGISTRYPROJECTNAME)/notary-signer-photon
DOCKERFILEPATH_NOTARYSIGNER=$(DOCKERFILEPATH)/notary-signer
DOCKERFILENAME_NOTARYSERVER=Dockerfile
DOCKERIMAGENAME_NOTARYSERVER=$(REGISTRYPROJECTNAME)/notary-server-photon

DOCKERFILEPATH_REDIS=$(DOCKERFILEPATH)/redis
DOCKERFILENAME_REDIS=Dockerfile
DOCKERIMAGENAME_REDIS=$(REGISTRYPROJECTNAME)/redis-photon

DOCKERFILEPATH_STANDALONE_DB_MIGRATOR=$(DOCKERFILEPATH)/standalone-db-migrator
DOCKERFILENAME_STANDALONE_DB_MIGRATOR=Dockerfile
DOCKERIMAGENAME_STANDALONE_DB_MIGRATOR=$(REGISTRYPROJECTNAME)/standalone-db-migrator

DOCKERFILEPATH_EXPORTER=$(DOCKERFILEPATH)/exporter
DOCKERFILENAME_EXPORTER=Dockerfile
DOCKERIMAGENAME_EXPORTER=goharbor/harbor-exporter

# for chart server (chartmuseum)
DOCKERFILEPATH_CHART_SERVER=$(DOCKERFILEPATH)/chartserver
DOCKERFILENAME_CHART_SERVER=Dockerfile
CHART_SERVER_CODE_BASE=https://github.com/helm/chartmuseum.git
CHART_SERVER_MAIN_PATH=cmd/chartmuseum
CHART_SERVER_BIN_NAME=chartm

_build_prepare:
	@echo "building prepare container for photon..."
	@$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_PREPARE)/$(DOCKERFILENAME_PREPARE) -t $(DOCKERIMAGENAME_PREPARE):$(VERSIONTAG) .
	@echo "Done."

_build_db:
	@echo "building db container for photon..."
	@$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_DB)/$(DOCKERFILENAME_DB) -t $(DOCKERIMAGENAME_DB):$(VERSIONTAG) .
	@echo "Done."

_build_portal:
	@echo "building portal container for photon..."
	$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) --build-arg npm_registry=$(NPM_REGISTRY) -f $(DOCKERFILEPATH_PORTAL)/$(DOCKERFILENAME_PORTAL) -t $(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG) .
	@echo "Done."

_build_core:
	@echo "building core container for photon..."
	@$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_CORE)/$(DOCKERFILENAME_CORE) -t $(DOCKERIMAGENAME_CORE):$(VERSIONTAG) .
	@echo "Done."

_build_jobservice:
	@echo "building jobservice container for photon..."
	@$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_JOBSERVICE)/$(DOCKERFILENAME_JOBSERVICE) -t $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) .
	@echo "Done."

_build_log:
	@echo "building log container for photon..."
	$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_LOG)/$(DOCKERFILENAME_LOG) -t $(DOCKERIMAGENAME_LOG):$(VERSIONTAG) .
	@echo "Done."

_build_trivy_adapter:
	@if [ "$(TRIVYFLAG)" = "true" ] ; then \
		if [ "$(BUILDBIN)" != "true" ] ; then \
			rm -rf $(DOCKERFILEPATH_TRIVY_ADAPTER)/binary && mkdir -p $(DOCKERFILEPATH_TRIVY_ADAPTER)/binary ; \
			echo "Downloading Trivy adapter $(TRIVYADAPTERVERSION)..." ; \
			$(call _extract_archive, $(TRIVY_ADAPTER_DOWNLOAD_URL), $(DOCKERFILEPATH_TRIVY_ADAPTER)/binary/) ; \
			mv $(DOCKERFILEPATH_TRIVY_ADAPTER)/binary/scanner-trivy $(DOCKERFILEPATH_TRIVY_ADAPTER)/binary/scanner-trivy-linux-amd64; \
		else \
			echo "Building Trivy adapter $(TRIVYADAPTERVERSION) from sources..." ; \
			cd $(DOCKERFILEPATH_TRIVY_ADAPTER) && TARGETARCHS="$(TARGETARCHS)" $(DOCKERFILEPATH_TRIVY_ADAPTER)/builder.sh $(TRIVYADAPTERVERSION) && cd - ; \
		fi ; \
		echo "Building Trivy adapter container for photon..." ; \
		$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) \
			--build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) \
			--build-arg trivy_version=$(TRIVYVERSION) \
			-f $(DOCKERFILEPATH_TRIVY_ADAPTER)/$(DOCKERFILENAME_TRIVY_ADAPTER) \
			-t $(DOCKERIMAGENAME_TRIVY_ADAPTER):$(VERSIONTAG) . ; \
		echo "Done." ; \
	fi

_build_chart_server:
	@if [ "$(CHARTFLAG)" = "true" ] ; then \
		if [ "$(BUILDBIN)" != "true" ] ; then \
			rm -rf $(DOCKERFILEPATH_CHART_SERVER)/binary && mkdir -p $(DOCKERFILEPATH_CHART_SERVER)/binary && \
			$(call _get_binary, $(CHARTURL), $(DOCKERFILEPATH_CHART_SERVER)/binary/chartm-linux-amd64); \
		else \
			cd $(DOCKERFILEPATH_CHART_SERVER) && TARGETARCHS="$(TARGETARCHS)" $(DOCKERFILEPATH_CHART_SERVER)/builder $(GOBUILDIMAGE) $(CHART_SERVER_CODE_BASE) $(CHARTMUSEUM_SRC_TAG) $(CHART_SERVER_MAIN_PATH) $(CHART_SERVER_BIN_NAME) && cd - ; \
		fi ; \
		echo "building chartmuseum container for photon..." ; \
		$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_CHART_SERVER)/$(DOCKERFILENAME_CHART_SERVER) -t $(DOCKERIMAGENAME_CHART_SERVER):$(VERSIONTAG) . ; \
		echo "Done." ; \
	fi

_build_nginx:
	@echo "building nginx container for photon..."
	@$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_NGINX)/$(DOCKERFILENAME_NGINX) -t $(DOCKERIMAGENAME_NGINX):$(VERSIONTAG) .
	@echo "Done."

_build_notary_binary:
	@if [ "$(NOTARYFLAG)" = "true" ] ; then \
		if [ "$(BUILDBIN)" != "true" ] ; then \
			mkdir -p $(DOCKERFILEPATH_NOTARY)/binary && rm -r $(DOCKERFILEPATH_NOTARY)/binary/notary-* && rm -r $(DOCKERFILEPATH_NOTARY)/binary/migrate-linux-*; \
			$(call _get_binary, $(NORARYURL), $(DOCKERFILEPATH_NOTARY)/binary-bundle.tgz); \
			cd $(DOCKERFILEPATH_NOTARY) && tar -zvxf binary-bundle.tgz && cd - ; \
			mv $(DOCKERFILEPATH_NOTARY)/binary/migrate $(DOCKERFILEPATH_NOTARY)/binary/migrate-linux-amd64; \
			mv $(DOCKERFILEPATH_NOTARY)/binary/notary-signer $(DOCKERFILEPATH_NOTARY)/binary/notary-signer-linux-amd64; \
			mv $(DOCKERFILEPATH_NOTARY)/binary/notary-server $(DOCKERFILEPATH_NOTARY)/binary/notary-server-linux-amd64; \
		else \
			cd $(DOCKERFILEPATH_NOTARY) && $(DOCKERFILEPATH_NOTARY)/builder $(NOTARYVERSION) $(NOTARYMIGRATEVERSION) && cd - ; \
		fi ; \
		chmod 655 $(DOCKERFILEPATH_NOTARY)/binary/notary-signer*; \
		chmod 655 $(DOCKERFILEPATH_NOTARY)/binary/notary-server*; \
	fi

_build_notary_signer:
	@if [ "$(NOTARYFLAG)" = "true" ] ; then \
		echo "building notary signer container for photon..."; \
		$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_NOTARYSIGNER)/$(DOCKERFILENAME_NOTARYSIGNER) -t $(DOCKERIMAGENAME_NOTARYSIGNER):$(VERSIONTAG) . ; \
		echo "Done."; \
	fi

_build_notary_server:
	@if [ "$(NOTARYFLAG)" = "true" ] ; then \
		echo "building notary server container for photon..."; \
		$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_NOTARYSERVER)/$(DOCKERFILENAME_NOTARYSERVER) -t $(DOCKERIMAGENAME_NOTARYSERVER):$(VERSIONTAG) . ; \
		echo "Done."; \
	fi

_build_notary: _build_notary_binary _build_notary_signer _build_notary_server

_build_registry:
	@if [ "$(BUILDBIN)" != "true" ] ; then \
		rm -rf $(DOCKERFILEPATH_REG)/binary && mkdir -p $(DOCKERFILEPATH_REG)/binary && \
		$(call _get_binary, $(REGISTRYURL), $(DOCKERFILEPATH_REG)/binary/registry); \
		mv $(DOCKERFILEPATH_REG)/binary/registry $(DOCKERFILEPATH_REG)/binary/registry-linux-amd64; \
	else \
		cd $(DOCKERFILEPATH_REG) && TARGETARCHS="$(TARGETARCHS)" $(DOCKERFILEPATH_REG)/builder $(REGISTRY_SRC_TAG) && cd - ; \
	fi
	@echo "building registry container for photon..."
	@chmod 655 $(DOCKERFILEPATH_REG)/binary/registry* && $(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_REG)/$(DOCKERFILENAME_REG) -t $(DOCKERIMAGENAME_REG):$(VERSIONTAG) .
	@echo "Done."

_build_registryctl:
	@echo "building registry controller for photon..."
	@$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_REGISTRYCTL)/$(DOCKERFILENAME_REGISTRYCTL) -t $(DOCKERIMAGENAME_REGISTRYCTL):$(VERSIONTAG) .
	@rm -rf $(DOCKERFILEPATH_REG)/binary
	@echo "Done."

_build_redis:
	@echo "building redis container for photon..."
	@$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_REDIS)/$(DOCKERFILENAME_REDIS) -t $(DOCKERIMAGENAME_REDIS):$(VERSIONTAG) .
	@echo "Done."

_build_standalone_db_migrator:
	@echo "building standalone db migrator image for photon..."
	@$(DOCKERBUILD) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_STANDALONE_DB_MIGRATOR)/$(DOCKERFILENAME_STANDALONE_DB_MIGRATOR) -t $(DOCKERIMAGENAME_STANDALONE_DB_MIGRATOR):$(VERSIONTAG) .
	@echo "Done."

_compile_and_build_exporter:
	@echo "compiling and building image for exporter..."
	@$(DOCKERCMD) build --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) --build-arg build_image=$(GOBUILDIMAGE) -f ${DOCKERFILEPATH_EXPORTER}/${DOCKERFILENAME_EXPORTER} -t $(DOCKERIMAGENAME_EXPORTER):$(VERSIONTAG) .
	@echo "Done."

define _extract_archive
	echo "download $1";\
	$(CURL) --connect-timeout 30 -f -k -L $1 | tar xvz -C $2 || exit 1
endef

define _get_binary
	echo "download $1";\
	$(CURL) --connect-timeout 30 -f -k -L $1 -o $2 || exit 1
endef



build: _build_prepare _build_db _build_portal _build_core _build_jobservice _build_log _build_nginx _build_registry _build_registryctl _build_notary _build_trivy_adapter _build_redis _build_chart_server _compile_and_build_exporter

cleanimage:
	@echo "cleaning image for photon..."
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_CORE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_LOG):$(VERSIONTAG)

.PHONY: clean
clean: cleanimage

