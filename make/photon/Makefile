# Makefile for a harbor project
#	
# Targets:
#
# build: 	build harbor photon images
# clean:	clean adminserver, ui and jobservice harbor images

# common
SHELL := /bin/bash
BUILDPATH=$(CURDIR)
MAKEPATH=$(BUILDPATH)/make
MAKEDEVPATH=$(MAKEPATH)/dev
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
ADMINSERVERSOURCECODE=$(SRCPATH)/adminserver
ADMINSERVERBINARYPATH=$(MAKEDEVPATH)/adminserver
ADMINSERVERBINARYNAME=harbor_adminserver
UISOURCECODE=$(SRCPATH)/ui
UIBINARYPATH=$(MAKEDEVPATH)/ui
UIBINARYNAME=harbor_ui
JOBSERVICESOURCECODE=$(SRCPATH)/jobservice
JOBSERVICEBINARYPATH=$(MAKEDEVPATH)/jobservice
JOBSERVICEBINARYNAME=harbor_jobservice

# photon dockerfile
DOCKERFILEPATH=$(MAKEPATH)/photon

DOCKERFILEPATH_ADMINSERVER=$(DOCKERFILEPATH)/adminserver
DOCKERFILENAME_ADMINSERVER=Dockerfile
DOCKERIMAGENAME_ADMINSERVER=vmware/harbor-adminserver

DOCKERFILEPATH_UI=$(DOCKERFILEPATH)/ui
DOCKERFILENAME_UI=Dockerfile
DOCKERIMAGENAME_UI=vmware/harbor-ui

DOCKERFILEPATH_JOBSERVICE=$(DOCKERFILEPATH)/jobservice
DOCKERFILENAME_JOBSERVICE=Dockerfile
DOCKERIMAGENAME_JOBSERVICE=vmware/harbor-jobservice

DOCKERFILEPATH_LOG=$(DOCKERFILEPATH)/log
DOCKERFILENAME_LOG=Dockerfile
DOCKERIMAGENAME_LOG=vmware/harbor-log

DOCKERFILEPATH_DB=$(DOCKERFILEPATH)/db
DOCKERFILENAME_DB=Dockerfile
DOCKERIMAGENAME_DB=vmware/harbor-db

DOCKERFILEPATH_POSTGRESQL=$(DOCKERFILEPATH)/postgresql
DOCKERFILENAME_POSTGRESQL=Dockerfile
DOCKERIMAGENAME_POSTGRESQL=vmware/postgresql-photon

DOCKERFILEPATH_CLAIR=$(DOCKERFILEPATH)/clair
DOCKERFILENAME_CLAIR=Dockerfile
DOCKERIMAGENAME_CLAIR=vmware/clair-photon

DOCKERFILEPATH_NGINX=$(DOCKERFILEPATH)/nginx
DOCKERFILENAME_NGINX=Dockerfile
DOCKERIMAGENAME_NGINX=vmware/nginx-photon

DOCKERFILEPATH_REG=$(DOCKERFILEPATH)/registry
DOCKERFILENAME_REG=Dockerfile
DOCKERIMAGENAME_REG=vmware/registry-photon

DOCKERFILEPATH_NOTARY=$(DOCKERFILEPATH)/notary
DOCKERFILENAME_NOTARYSIGNER=signer.Dockerfile
DOCKERIMAGENAME_NOTARYSIGNER=vmware/notary-signer-photon
DOCKERFILENAME_NOTARYSERVER=server.Dockerfile
DOCKERIMAGENAME_NOTARYSERVER=vmware/notary-server-photon

DOCKERFILEPATH_REDIS=$(DOCKERFILEPATH)/redis
DOCKERFILENAME_REDIS=Dockerfile
DOCKERIMAGENAME_REDIS=vmware/redis-photon

DOCKERFILEPATH_MIGRATOR=$(TOOLSPATH)/migration
DOCKERFILENAME_MIGRATOR=Dockerfile
DOCKERIMAGENAME_MIGRATOR=vmware/harbor-migrator

_build_db:
	@echo "building db container for photon..."
	@cd $(DOCKERFILEPATH_DB) && $(DOCKERBUILD) -f $(DOCKERFILEPATH_DB)/$(DOCKERFILENAME_DB) -t $(DOCKERIMAGENAME_DB):$(VERSIONTAG) .
	@echo "Done."

_build_adminiserver:		
	@echo "building adminserver container for photon..."
	$(DOCKERBUILD) -f $(DOCKERFILEPATH_ADMINSERVER)/$(DOCKERFILENAME_ADMINSERVER) -t $(DOCKERIMAGENAME_ADMINSERVER):$(VERSIONTAG) .
	@echo "Done."

_build_ui:	
	@echo "building ui container for photon..."
	$(DOCKERBUILD) -f $(DOCKERFILEPATH_UI)/$(DOCKERFILENAME_UI) -t $(DOCKERIMAGENAME_UI):$(VERSIONTAG) .
	@echo "Done."
	
_build_jobservice:		
	@echo "building jobservice container for photon..."
	$(DOCKERBUILD) -f $(DOCKERFILEPATH_JOBSERVICE)/$(DOCKERFILENAME_JOBSERVICE) -t $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) .
	@echo "Done."

_build_log:	
	@echo "building log container for photon..."
	$(DOCKERBUILD) -f $(DOCKERFILEPATH_LOG)/$(DOCKERFILENAME_LOG) -t $(DOCKERIMAGENAME_LOG):$(VERSIONTAG) $(DOCKERFILEPATH_LOG)
	@echo "Done."

_build_clair:
	@if [ "$(CLAIRFLAG)" = "true" ] ; then \
		if [ "$(BUILDBIN)" != "true" ] ; then \
			rm -rf $(DOCKERFILEPATH_CLAIR)/binary && mkdir -p $(DOCKERFILEPATH_CLAIR)/binary && \
			$(call _get_binary, https://storage.googleapis.com/harbor-builds/bin/clair, $(DOCKERFILEPATH_CLAIR)/binary/clair); \
		else \
			cd $(DOCKERFILEPATH_CLAIR) && $(DOCKERFILEPATH_CLAIR)/builder $(CLAIRVERSION); \
		fi ; \
		echo "building clair container for photon..." ; \
		cd $(DOCKERFILEPATH_CLAIR) && $(DOCKERBUILD) -f $(DOCKERFILEPATH_CLAIR)/$(DOCKERFILENAME_CLAIR) -t $(DOCKERIMAGENAME_CLAIR):$(CLAIRVERSION)-$(VERSIONTAG) . ; \
		rm -rf $(DOCKERFILEPATH_CLAIR)/binary; \
		echo "Done." ; \
	fi
		
_build_nginx:
	@echo "building nginx container for photon..."
	@$(DOCKERBUILD) -f $(DOCKERFILEPATH_NGINX)/$(DOCKERFILENAME_NGINX) -t $(DOCKERIMAGENAME_NGINX):$(NGINXVERSION) .
	@echo "Done."

_build_notary:
	@if [ "$(NOTARYFLAG)" = "true" ] ; then \
		if [ "$(BUILDBIN)" != "true" ] ; then \
			rm -rf $(DOCKERFILEPATH_NOTARY)/binary && mkdir -p $(DOCKERFILEPATH_NOTARY)/binary && \
			$(call _get_binary, https://storage.googleapis.com/harbor-builds/bin/notary/notary-signer, $(DOCKERFILEPATH_NOTARY)/binary/notary-signer) && \
			$(call _get_binary, https://storage.googleapis.com/harbor-builds/bin/notary/notary-server, $(DOCKERFILEPATH_NOTARY)/binary/notary-server) && \
			$(call _get_binary, https://storage.googleapis.com/harbor-builds/bin/notary/notary-migrate-postgresql.tgz, $(DOCKERFILEPATH_NOTARY)/binary/notary-migrate.tgz); \
			cd $(DOCKERFILEPATH_NOTARY)/binary && tar -zvxf notary-migrate.tgz; \
		else \
			cd $(DOCKERFILEPATH_NOTARY) && $(DOCKERFILEPATH_NOTARY)/builder_public $(NOTARYVERSION); \
		fi ; \
		echo "building notary container for photon..."; \
		cd $(DOCKERFILEPATH_NOTARY) && chmod 655 $(DOCKERFILEPATH_NOTARY)/binary/notary-signer && $(DOCKERBUILD) -f $(DOCKERFILEPATH_NOTARY)/$(DOCKERFILENAME_NOTARYSIGNER) -t $(DOCKERIMAGENAME_NOTARYSIGNER):$(NOTARYVERSION)-$(VERSIONTAG) . ; \
		cd $(DOCKERFILEPATH_NOTARY) && chmod 655 $(DOCKERFILEPATH_NOTARY)/binary/notary-server && $(DOCKERBUILD) -f $(DOCKERFILEPATH_NOTARY)/$(DOCKERFILENAME_NOTARYSERVER) -t $(DOCKERIMAGENAME_NOTARYSERVER):$(NOTARYVERSION)-$(VERSIONTAG) . ; \
		rm -rf $(DOCKERFILEPATH_NOTARY)/binary; \
		echo "Done."; \
	fi
		
_build_registry:
	@if [ "$(BUILDBIN)" != "true" ] ; then \
		rm -rf $(DOCKERFILEPATH_REG)/binary && mkdir -p $(DOCKERFILEPATH_REG)/binary && \
		$(call _get_binary, https://storage.googleapis.com/harbor-builds/bin/registry, $(DOCKERFILEPATH_REG)/binary/registry); \
	else \
		cd $(DOCKERFILEPATH_REG) && $(DOCKERFILEPATH_REG)/builder $(REGISTRYVERSION); \
	fi
	@echo "building registry container for photon..."
	@cd $(DOCKERFILEPATH_REG) && chmod 655 $(DOCKERFILEPATH_REG)/binary/registry && $(DOCKERBUILD) -f $(DOCKERFILEPATH_REG)/$(DOCKERFILENAME_REG) -t $(DOCKERIMAGENAME_REG):$(REGISTRYVERSION)-$(VERSIONTAG) .
	@rm -rf $(DOCKERFILEPATH_REG)/binary
	@echo "Done."

_build_redis:
	@echo "building redis container for photon..."
	@cd $(DOCKERFILEPATH_REDIS) && $(DOCKERBUILD) -f $(DOCKERFILEPATH_REDIS)/$(DOCKERFILENAME_REDIS) -t $(DOCKERIMAGENAME_REDIS):$(REDISVERSION) .
	@echo "Done."

_build_migrator:
	@echo "building db migrator container for photon..."
	@cd $(DOCKERFILEPATH_MIGRATOR) && $(DOCKERBUILD) -f $(DOCKERFILEPATH_MIGRATOR)/$(DOCKERFILENAME_MIGRATOR) -t $(DOCKERIMAGENAME_MIGRATOR):$(VERSIONTAG) .
	@echo "Done."

define _get_binary
	$(WGET) --timeout 30 --no-check-certificate $1 -O $2
endef

build: _build_db _build_adminiserver _build_ui _build_jobservice _build_log _build_nginx _build_registry _build_notary _build_clair _build_redis _build_migrator

cleanimage:
	@echo "cleaning image for photon..."
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_ADMINSERVER):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_UI):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_LOG):$(VERSIONTAG)
	
.PHONY: clean
clean: cleanimage

