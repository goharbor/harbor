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
TIMESTAMP=$(shell date +"%Y%m%d")

# docker parameters
DOCKERCMD=$(shell which docker)
DOCKERBUILD=$(DOCKERCMD) build --no-cache --network=$(DOCKERNETWORK)
DOCKERBUILD_WITH_PULL_PARA=$(DOCKERBUILD) --pull=$(PULL_BASE_FROM_DOCKERHUB)
DOCKERRMIMAGE=$(DOCKERCMD) rmi
DOCKERIMAGES=$(DOCKERCMD) images
IMAGENAMESPACE=goharbor
BASEIMAGENAMESPACE=goharbor

# pushimage
PUSHSCRIPTPATH=$(MAKEPATH)
PUSHSCRIPTNAME=pushimage.sh

# binary
CORE_SOURCECODE=$(SRCPATH)/core
CORE_BINARYNAME=harbor_core
JOBSERVICESOURCECODE=$(SRCPATH)/jobservice
JOBSERVICEBINARYNAME=harbor_jobservice

# photon dockerfile
DOCKERFILEPATH=$(MAKEPATH)/photon

PREPARE=prepare
DOCKERFILEPATH_PREPARE=$(DOCKERFILEPATH)/$(PREPARE)
DOCKERFILENAME_PREPARE=Dockerfile
DOCKERIMAGENAME_PREPARE=$(IMAGENAMESPACE)/$(PREPARE)

PORTAL=portal
DOCKERFILEPATH_PORTAL=$(DOCKERFILEPATH)/$(PORTAL)
DOCKERFILENAME_PORTAL=Dockerfile
DOCKERIMAGENAME_PORTAL=$(IMAGENAMESPACE)/harbor-$(PORTAL)

CORE=core
DOCKERFILEPATH_CORE=$(DOCKERFILEPATH)/$(CORE)
DOCKERFILENAME_CORE=Dockerfile
DOCKERIMAGENAME_CORE=$(IMAGENAMESPACE)/harbor-$(CORE)

JOBSERVICE=jobservice
DOCKERFILEPATH_JOBSERVICE=$(DOCKERFILEPATH)/$(JOBSERVICE)
DOCKERFILENAME_JOBSERVICE=Dockerfile
DOCKERIMAGENAME_JOBSERVICE=$(IMAGENAMESPACE)/harbor-$(JOBSERVICE)

LOG=log
DOCKERFILEPATH_LOG=$(DOCKERFILEPATH)/$(LOG)
DOCKERFILENAME_LOG=Dockerfile
DOCKERIMAGENAME_LOG=$(IMAGENAMESPACE)/harbor-$(LOG)

DB=db
DOCKERFILEPATH_DB=$(DOCKERFILEPATH)/$(DB)
DOCKERFILENAME_DB=Dockerfile
DOCKERIMAGENAME_DB=$(IMAGENAMESPACE)/harbor-$(DB)

POSTGRESQL=postgresql
DOCKERFILEPATH_POSTGRESQL=$(DOCKERFILEPATH)/$(POSTGRESQL)
DOCKERFILENAME_POSTGRESQL=Dockerfile
DOCKERIMAGENAME_POSTGRESQL=$(IMAGENAMESPACE)/$(POSTGRESQL)-photon

TRIVY_ADAPTER=trivy-adapter
DOCKERFILEPATH_TRIVY_ADAPTER=$(DOCKERFILEPATH)/$(TRIVY_ADAPTER)
DOCKERFILENAME_TRIVY_ADAPTER=Dockerfile
DOCKERIMAGENAME_TRIVY_ADAPTER=$(IMAGENAMESPACE)/$(TRIVY_ADAPTER)-photon

NGINX=nginx
DOCKERFILEPATH_NGINX=$(DOCKERFILEPATH)/$(NGINX)
DOCKERFILENAME_NGINX=Dockerfile
DOCKERIMAGENAME_NGINX=$(IMAGENAMESPACE)/$(NGINX)-photon

REGISTRY=registry
DOCKERFILEPATH_REG=$(DOCKERFILEPATH)/$(REGISTRY)
DOCKERFILENAME_REG=Dockerfile
DOCKERIMAGENAME_REG=$(IMAGENAMESPACE)/$(REGISTRY)-photon

REGISTRYCTL=registryctl
DOCKERFILEPATH_REGISTRYCTL=$(DOCKERFILEPATH)/$(REGISTRYCTL)
DOCKERFILENAME_REGISTRYCTL=Dockerfile
DOCKERIMAGENAME_REGISTRYCTL=$(IMAGENAMESPACE)/harbor-$(REGISTRYCTL)

REDIS=redis
DOCKERFILEPATH_REDIS=$(DOCKERFILEPATH)/$(REDIS)
DOCKERFILENAME_REDIS=Dockerfile
DOCKERIMAGENAME_REDIS=$(IMAGENAMESPACE)/$(REDIS)-photon

DOCKERFILEPATH_STANDALONE_DB_MIGRATOR=$(DOCKERFILEPATH)/standalone-db-migrator
DOCKERFILENAME_STANDALONE_DB_MIGRATOR=Dockerfile
DOCKERIMAGENAME_STANDALONE_DB_MIGRATOR=$(IMAGENAMESPACE)/standalone-db-migrator

EXPORTER=exporter
DOCKERFILEPATH_EXPORTER=$(DOCKERFILEPATH)/$(EXPORTER)
DOCKERFILENAME_EXPORTER=Dockerfile
DOCKERIMAGENAME_EXPORTER=$(IMAGENAMESPACE)/harbor-$(EXPORTER)

_build_prepare:
	@$(call _build_base,$(PREPARE),$(DOCKERFILEPATH_PREPARE))
	@echo "building prepare container for photon..."
	@$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_PREPARE)/$(DOCKERFILENAME_PREPARE) -t $(DOCKERIMAGENAME_PREPARE):$(VERSIONTAG) .
	@echo "Done."

_build_db:
	@$(call _build_base,$(DB),$(DOCKERFILEPATH_DB))
	@echo "building db container for photon..."
	@$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_DB)/$(DOCKERFILENAME_DB) -t $(DOCKERIMAGENAME_DB):$(VERSIONTAG) .
	@echo "Done."

_build_portal:
	@$(call _build_base,$(PORTAL),$(DOCKERFILEPATH_PORTAL))
	@echo "building portal container for photon..."
	$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) --build-arg NODE=${NODEBUILDIMAGE} --build-arg npm_registry=$(NPM_REGISTRY) -f $(DOCKERFILEPATH_PORTAL)/$(DOCKERFILENAME_PORTAL) -t $(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG) .
	@echo "Done."

_build_core:
	@$(call _build_base,$(CORE),$(DOCKERFILEPATH_CORE))
	@echo "building core container for photon..."
	@$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_CORE)/$(DOCKERFILENAME_CORE) -t $(DOCKERIMAGENAME_CORE):$(VERSIONTAG) .
	@echo "Done."

_build_jobservice:
	@$(call _build_base,$(JOBSERVICE),$(DOCKERFILEPATH_JOBSERVICE))
	@echo "building jobservice container for photon..."
	@$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_JOBSERVICE)/$(DOCKERFILENAME_JOBSERVICE) -t $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) .
	@echo "Done."

_build_log:
	@$(call _build_base,$(LOG),$(DOCKERFILEPATH_LOG))
	@echo "building log container for photon..."
	$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_LOG)/$(DOCKERFILENAME_LOG) -t $(DOCKERIMAGENAME_LOG):$(VERSIONTAG) .
	@echo "Done."

_build_trivy_adapter:
	@if [ "$(TRIVYFLAG)" = "true" ] ; then \
		$(call _build_base,$(TRIVY_ADAPTER),$(DOCKERFILEPATH_TRIVY_ADAPTER)) ; \
		rm -rf $(DOCKERFILEPATH_TRIVY_ADAPTER)/binary && mkdir -p $(DOCKERFILEPATH_TRIVY_ADAPTER)/binary ; \
		echo "Downloading Trivy scanner $(TRIVYVERSION)..." ; \
		$(call _extract_archive, $(TRIVY_DOWNLOAD_URL), $(DOCKERFILEPATH_TRIVY_ADAPTER)/binary/) ; \
		if [ "$(BUILDTRIVYADP)" != "true" ] ; then \
			echo "Downloading Trivy adapter $(TRIVYADAPTERVERSION)..." ; \
			$(call _extract_archive, $(TRIVY_ADAPTER_DOWNLOAD_URL), $(DOCKERFILEPATH_TRIVY_ADAPTER)/binary/) ; \
		else \
			echo "Building Trivy adapter $(TRIVYADAPTERVERSION) from sources..." ; \
			cd $(DOCKERFILEPATH_TRIVY_ADAPTER) && $(DOCKERFILEPATH_TRIVY_ADAPTER)/builder.sh $(TRIVYADAPTERVERSION) $(GOBUILDIMAGE) $(DOCKERNETWORK) && cd - ; \
		fi ; \
		echo "Building Trivy adapter container for photon..." ; \
		$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) \
			--build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) \
			--build-arg trivy_version=$(TRIVYVERSION) \
			-f $(DOCKERFILEPATH_TRIVY_ADAPTER)/$(DOCKERFILENAME_TRIVY_ADAPTER) \
			-t $(DOCKERIMAGENAME_TRIVY_ADAPTER):$(VERSIONTAG) . ; \
		rm -rf $(DOCKERFILEPATH_TRIVY_ADAPTER)/binary; \
		echo "Done." ; \
	fi

_build_nginx:
	@$(call _build_base,$(NGINX),$(DOCKERFILEPATH_NGINX))
	@echo "building nginx container for photon..."
	@$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_NGINX)/$(DOCKERFILENAME_NGINX) -t $(DOCKERIMAGENAME_NGINX):$(VERSIONTAG) .
	@echo "Done."

_build_registry:
	@$(call _build_base,$(REGISTRY),$(DOCKERFILEPATH_REG))
	@if [ "$(BUILDREG)" != "true" ] ; then \
		rm -rf $(DOCKERFILEPATH_REG)/binary && mkdir -p $(DOCKERFILEPATH_REG)/binary && \
		$(call _get_binary, $(REGISTRYURL), $(DOCKERFILEPATH_REG)/binary/registry); \
	else \
		cd $(DOCKERFILEPATH_REG) && $(DOCKERFILEPATH_REG)/builder $(REGISTRY_SRC_TAG) $(DISTRIBUTION_SRC) $(GOBUILDIMAGE) $(DOCKERNETWORK) && cd - ; \
	fi
	@echo "building registry container for photon..."
	@chmod 655 $(DOCKERFILEPATH_REG)/binary/registry && $(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_REG)/$(DOCKERFILENAME_REG) -t $(DOCKERIMAGENAME_REG):$(VERSIONTAG) .
	@echo "Done."

_build_registryctl:
	@$(call _build_base,$(REGISTRYCTL),$(DOCKERFILEPATH_REGISTRYCTL))
	@echo "building registry controller for photon..."
	@$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_REGISTRYCTL)/$(DOCKERFILENAME_REGISTRYCTL) -t $(DOCKERIMAGENAME_REGISTRYCTL):$(VERSIONTAG) .
	@rm -rf $(DOCKERFILEPATH_REG)/binary
	@echo "Done."

_build_redis:
	@$(call _build_base,$(REDIS),$(DOCKERFILEPATH_REDIS))
	@echo "building redis container for photon..."
	@$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_REDIS)/$(DOCKERFILENAME_REDIS) -t $(DOCKERIMAGENAME_REDIS):$(VERSIONTAG) .
	@echo "Done."

_build_standalone_db_migrator:
	@echo "building standalone db migrator image for photon..."
	$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_STANDALONE_DB_MIGRATOR)/$(DOCKERFILENAME_STANDALONE_DB_MIGRATOR) -t $(DOCKERIMAGENAME_STANDALONE_DB_MIGRATOR):$(VERSIONTAG) .
	@echo "Done."

_compile_and_build_exporter:
	@$(call _build_base,$(EXPORTER),$(DOCKERFILEPATH_EXPORTER))
	@echo "compiling and building image for exporter..."
	@$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) --build-arg build_image=$(GOBUILDIMAGE) --build-arg harbor_version=$(RELEASEVERSION) --build-arg git_commit=$(GITCOMMIT) -f ${DOCKERFILEPATH_EXPORTER}/${DOCKERFILENAME_EXPORTER} -t $(DOCKERIMAGENAME_EXPORTER):$(VERSIONTAG) .
	@echo "Done."

define _extract_archive
	echo "download $1";\
	$(CURL) --connect-timeout 30 -f -k -L $1 | tar xvz -C $2 || exit 1
endef

define _get_binary
	echo "download $1";\
	$(CURL) --connect-timeout 30 -f -k -L $1 -o $2 || exit 1
endef

define _build_base
	if [ "$(BUILD_BASE)" = "true" ] ; then \
		echo "building base image for $(1)...";\
		if [ -n "$(REGISTRYUSER)"  ] && [ -n "$(REGISTRYPASSWORD)" ] ; then \
			docker login -u $(REGISTRYUSER) -p $(REGISTRYPASSWORD) ; \
		else \
			echo "No docker credentials provided, please be aware of privileges to access docker hub!" ; \
		fi ;\
		$(DOCKERBUILD) -f $(2)/Dockerfile.base -t $(BASEIMAGENAMESPACE)/harbor-$(1)-base:$(BASEIMAGETAG) --label base-build-date=$(TIMESTAMP) . ;\
		if [ "$(PUSHBASEIMAGE)" = "true" ] ; then \
			$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(BASEIMAGENAMESPACE)/harbor-$(1)-base:$(BASEIMAGETAG) $(REGISTRYUSER) $(REGISTRYPASSWORD) docker.io $(PULL_BASE_FROM_DOCKERHUB) || exit 1; \
		fi ; \
	fi
endef

ifeq ($(BUILD_INSTALLER), true)
buildcompt: _build_prepare _build_db _build_portal _build_core _build_jobservice _build_log _build_nginx _build_registry _build_registryctl _build_trivy_adapter _build_redis _compile_and_build_exporter
else
buildcompt: _build_db _build_portal _build_core _build_jobservice _build_nginx _build_registry _build_registryctl _build_trivy_adapter _build_redis _compile_and_build_exporter
endif

build: buildcompt
	@if [ -n "$(REGISTRYUSER)"  ] && [ -n "$(REGISTRYPASSWORD)" ] ; then \
		docker logout ; \
	fi

cleanimage:
	@echo "cleaning image for photon..."
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_CORE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_LOG):$(VERSIONTAG)

.PHONY: clean
clean: cleanimage
