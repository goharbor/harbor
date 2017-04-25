# Makefile for Harbor project
#	
# Targets:
#
# all:			prepare env, compile binarys, build images and install images 
# prepare: 		prepare env
# compile: 		compile adminserver, ui and jobservice code
#
# compile_golangimage:
#			compile from golang image
#			for example: make compile_golangimage -e GOBUILDIMAGE= \
#							golang:1.7.3
# compile_adminserver, compile_ui, compile_jobservice: compile specific binary
#
# build: 		build Harbor docker images (defuault: build_photon)
#			for example: make build -e BASEIMAGE=photon
# build_photon:	build Harbor docker images from photon baseimage
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
#							REGISTRYSERVER=reg-bj.eng.vmware.com \
#							REGISTRYPROJECTNAME=harborrelease
#						
# package_offline:
#				prepare offline install package
# 
# pushimage:	push Harbor images to specific registry server
#			for example: make pushimage -e DEVFLAG=false REGISTRYUSER=admin \
#							REGISTRYPASSWORD=***** \
#							REGISTRYSERVER=reg-bj.eng.vmware.com/ \
#							REGISTRYPROJECTNAME=harborrelease
#				note**: need add "/" on end of REGISTRYSERVER. If not setting \
#						this value will push images directly to dockerhub.
#						 make pushimage -e DEVFLAG=false REGISTRYUSER=vmware \
#							REGISTRYPASSWORD=***** \
#							REGISTRYPROJECTNAME=vmware
#
# clean:        remove binary, Harbor images, specific version docker-compose \
#               file, specific version tag and online/offline install package
# cleanbinary:	remove adminserver, ui and jobservice binary
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
SRCPATH=./src
TOOLSPATH=$(BUILDPATH)/tools
UIPATH=$(BUILDPATH)/src/ui
UINGPATH=$(BUILDPATH)/src/ui_ng
GOBASEPATH=/go/src/github.com/vmware
CHECKENVCMD=checkenv.sh
BASEIMAGE=photon
COMPILETAG=compile_normal
REGISTRYSERVER=
REGISTRYPROJECTNAME=vmware
DEVFLAG=true
NOTARYFLAG=false
REGISTRYVERSION=photon-2.6.0
NGINXVERSION=1.11.5-patched
PHOTONVERSION=1.0
NOTARYVERSION=server-0.5.0
NOTARYSIGNERVERSION=signer-0.5.0
MARIADBVERSION=mariadb-10.1.10
HTTPPROXY=

#clarity parameters
CLARITYIMAGE=vmware/harbor-clarity-ui-builder[:tag]
CLARITYSEEDPATH=/clarity-seed
CLARITYBUILDSCRIPT=/entrypoint.sh

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
GOBUILDIMAGE=reg.mydomain.com/library/harborgo[:tag]
GOBUILDPATH=$(GOBASEPATH)/harbor
GOIMAGEBUILDCMD=/usr/local/go/bin/go
GOIMAGEBUILD=$(GOIMAGEBUILDCMD) build
GOBUILDPATH_ADMINSERVER=$(GOBUILDPATH)/src/adminserver
GOBUILDPATH_UI=$(GOBUILDPATH)/src/ui
GOBUILDPATH_JOBSERVICE=$(GOBUILDPATH)/src/jobservice
GOBUILDMAKEPATH=$(GOBUILDPATH)/make
GOBUILDMAKEPATH_ADMINSERVER=$(GOBUILDMAKEPATH)/dev/adminserver
GOBUILDMAKEPATH_UI=$(GOBUILDMAKEPATH)/dev/ui
GOBUILDMAKEPATH_JOBSERVICE=$(GOBUILDMAKEPATH)/dev/jobservice
GOLANGDOCKERFILENAME=Dockerfile.golang

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

# prepare parameters
PREPAREPATH=$(TOOLSPATH)
PREPARECMD=prepare

# configfile
CONFIGPATH=$(MAKEPATH)
CONFIGFILE=harbor.cfg

# makefile
MAKEFILEPATH_PHOTON=$(MAKEPATH)/photon

# common dockerfile
DOCKERFILEPATH_COMMON=$(MAKEPATH)/common
DOCKERFILEPATH_DB=$(DOCKERFILEPATH_COMMON)/db
DOCKERFILENAME_DB=Dockerfile

# docker image name
DOCKERIMAGENAME_ADMINSERVER=vmware/harbor-adminserver
DOCKERIMAGENAME_UI=vmware/harbor-ui
DOCKERIMAGENAME_JOBSERVICE=vmware/harbor-jobservice
DOCKERIMAGENAME_LOG=vmware/harbor-log
DOCKERIMAGENAME_DB=vmware/harbor-db

# docker-compose files
DOCKERCOMPOSEFILEPATH=$(MAKEPATH)
DOCKERCOMPOSETPLFILENAME=docker-compose.tpl
DOCKERCOMPOSEFILENAME=docker-compose.yml
DOCKERCOMPOSENOTARYFILENAME=docker-compose.notary.yml

# version prepare
VERSIONFILEPATH=$(CURDIR)
VERSIONFILENAME=VERSION
GITCMD=$(shell which git)
GITTAG=$(GITCMD) describe --tags
GITTAGVERSION=$(shell git describe --tags || echo UNKNOWN)
ifeq ($(DEVFLAG), true)        
	VERSIONTAG=dev
else        
	VERSIONTAG=$(GITTAGVERSION)
endif

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

version:
	@printf $(GITTAGVERSION) > $(VERSIONFILEPATH)/$(VERSIONFILENAME);
	
check_environment:
	@$(MAKEPATH)/$(CHECKENVCMD)

compile_adminserver:
	@echo "compiling binary for adminserver..."
	@$(GOBUILD) -o $(ADMINSERVERBINARYPATH)/$(ADMINSERVERBINARYNAME) $(ADMINSERVERSOURCECODE)
	@echo "Done."

compile_ui:
	@echo "compiling binary for ui..."
	@$(GOBUILD) -o $(UIBINARYPATH)/$(UIBINARYNAME) $(UISOURCECODE)
	@echo "Done."
	
compile_jobservice:
	@echo "compiling binary for jobservice..."
	@$(GOBUILD) -o $(JOBSERVICEBINARYPATH)/$(JOBSERVICEBINARYNAME) $(JOBSERVICESOURCECODE)
	@echo "Done."
	
compile_clarity:
	@echo "compiling binary for clarity ui..."
	@if [ "$(HTTPPROXY)" != "" ] ; then \
		$(DOCKERCMD) run --rm -v $(UIPATH)/static:$(CLARITYSEEDPATH)/dist -v $(UINGPATH)/src:$(CLARITYSEEDPATH)/src $(CLARITYIMAGE) $(SHELL) $(CLARITYBUILDSCRIPT) -p $(HTTPPROXY); \
	else \
		$(DOCKERCMD) run --rm -v $(UIPATH)/static:$(CLARITYSEEDPATH)/dist -v $(UINGPATH)/src:$(CLARITYSEEDPATH)/src $(CLARITYIMAGE) $(SHELL) $(CLARITYBUILDSCRIPT); \
	fi
	@echo "Done."
	
compile_normal: compile_clarity compile_adminserver compile_ui compile_jobservice

compile_golangimage: compile_clarity
	@echo "compiling binary for adminserver (golang image)..."
	@echo $(GOBASEPATH)
	@echo $(GOBUILDPATH)
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATH) -w $(GOBUILDPATH_ADMINSERVER) $(GOBUILDIMAGE) $(GOIMAGEBUILD) -v -o $(GOBUILDMAKEPATH_ADMINSERVER)/$(ADMINSERVERBINARYNAME)
	@echo "Done."

	@echo "compiling binary for ui (golang image)..."
	@echo $(GOBASEPATH)
	@echo $(GOBUILDPATH)
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATH) -w $(GOBUILDPATH_UI) $(GOBUILDIMAGE) $(GOIMAGEBUILD) -v -o $(GOBUILDMAKEPATH_UI)/$(UIBINARYNAME)
	@echo "Done."
	
	@echo "compiling binary for jobservice (golang image)..."
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATH) -w $(GOBUILDPATH_JOBSERVICE) $(GOBUILDIMAGE) $(GOIMAGEBUILD) -v -o $(GOBUILDMAKEPATH_JOBSERVICE)/$(JOBSERVICEBINARYNAME)
	@echo "Done."
	
compile:check_environment $(COMPILETAG)

prepare: 
	@echo "preparing..."
	@if [ "$(NOTARYFLAG)" = "true" ] ; then \
		$(MAKEPATH)/$(PREPARECMD) --conf $(CONFIGPATH)/$(CONFIGFILE) --with-notary; \
	else \
		$(MAKEPATH)/$(PREPARECMD) --conf $(CONFIGPATH)/$(CONFIGFILE) ; \
	fi	
	
build_common: version
	@echo "buildging db container for photon..."
	@cd $(DOCKERFILEPATH_DB) && $(DOCKERBUILD) -f $(DOCKERFILENAME_DB) -t $(DOCKERIMAGENAME_DB):$(VERSIONTAG) .
	@echo "Done."

build_photon: build_common
	make -f $(MAKEFILEPATH_PHOTON)/Makefile build -e DEVFLAG=$(DEVFLAG)
	
build: build_$(BASEIMAGE)
	
modify_composefile: 
	@echo "preparing docker-compose file..."
	@cp $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSETPLFILENAME) $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)
	@$(SEDCMD) -i 's/__version__/$(VERSIONTAG)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)

modify_sourcefiles:
	@echo "change mode of source files."
	@chmod 600 $(MAKEPATH)/common/templates/notary/notary-signer.key
	@chmod 600 $(MAKEPATH)/common/templates/notary/notary-signer.crt
	@chmod 600 $(MAKEPATH)/common/templates/notary/notary-signer-ca.crt
	@chmod 600 $(MAKEPATH)/common/templates/ui/private_key.pem
	@chmod 600 $(MAKEPATH)/common/templates/registry/root.crt
	
install: compile build modify_sourcefiles prepare modify_composefile start
	
package_online: modify_composefile
	@echo "packing online package ..."
	@cp -r make $(HARBORPKG)
	@if [ -n "$(REGISTRYSERVER)" ] ; then \
		$(SEDCMD) -i 's/image\: vmware/image\: $(REGISTRYSERVER)\/$(REGISTRYPROJECTNAME)/' \
		$(HARBORPKG)/docker-compose.yml ; \
	fi
	@cp LICENSE $(HARBORPKG)/LICENSE
	@cp NOTICE $(HARBORPKG)/NOTICE
	@cp tools/migration/migration_cfg/upgrade $(HARBORPKG)/upgrade
	@cp tools/migration/migration_cfg/harbor_1_1_0_template $(HARBORPKG)/harbor_1_1_0_template

	@if [ "$(NOTARYFLAG)" = "true" ] ; then \
		$(TARCMD) -zcvf harbor-online-installer-$(GITTAGVERSION).tgz \
		          $(HARBORPKG)/common/templates $(HARBORPKG)/prepare \
				  $(HARBORPKG)/LICENSE $(HARBORPKG)/NOTICE \
				  $(HARBORPKG)/upgrade $(HARBORPKG)/harbor_1_1_0_template \
				  $(HARBORPKG)/install.sh $(HARBORPKG)/$(DOCKERCOMPOSEFILENAME) \
				  $(HARBORPKG)/harbor.cfg $(HARBORPKG)/$(DOCKERCOMPOSENOTARYFILENAME); \
	else \
		$(TARCMD) -zcvf harbor-online-installer-$(GITTAGVERSION).tgz \
		          $(HARBORPKG)/common/templates $(HARBORPKG)/prepare \
				  $(HARBORPKG)/LICENSE $(HARBORPKG)/NOTICE \
				  $(HARBORPKG)/upgrade $(HARBORPKG)/harbor_1_1_0_template \
				  $(HARBORPKG)/install.sh $(HARBORPKG)/$(DOCKERCOMPOSEFILENAME) \
				  $(HARBORPKG)/harbor.cfg ; \
	fi
						
	@rm -rf $(HARBORPKG)
	@echo "Done."
		
package_offline: compile build modify_sourcefiles modify_composefile
	@echo "packing offline package ..."
	@cp -r make $(HARBORPKG)
	
	@cp LICENSE $(HARBORPKG)/LICENSE
	@cp NOTICE $(HARBORPKG)/NOTICE
	@cp tools/migration/migration_cfg/upgrade $(HARBORPKG)/upgrade
	@cp tools/migration/migration_cfg/harbor_1_1_0_template $(HARBORPKG)/harbor_1_1_0_template
			
	@echo "pulling nginx and registry..."
	@$(DOCKERPULL) vmware/registry:$(REGISTRYVERSION)
	@$(DOCKERPULL) vmware/nginx:$(NGINXVERSION)
	@if [ "$(NOTARYFLAG)" = "true" ] ; then \
		echo "pulling notary and harbor-notary-db..."; \
		$(DOCKERPULL) vmware/notary-photon:$(NOTARYVERSION); \
		$(DOCKERPULL) vmware/notary-photon:$(NOTARYSIGNERVERSION); \
		$(DOCKERPULL) vmware/harbor-notary-db:$(MARIADBVERSION); \
	fi	
	
	@echo "saving harbor docker image"
	@if [ "$(NOTARYFLAG)" = "true" ] ; then \
		$(DOCKERSAVE) $(DOCKERIMAGENAME_ADMINSERVER):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_UI):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_LOG):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_DB):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) \
		vmware/nginx:$(NGINXVERSION) vmware/registry:$(REGISTRYVERSION) photon:$(PHOTONVERSION) \
		vmware/notary-photon:$(NOTARYVERSION) vmware/notary-photon:$(NOTARYSIGNERVERSION) \
		vmware/harbor-notary-db:$(MARIADBVERSION) | gzip > $(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tar.gz; \
	else \
		$(DOCKERSAVE) $(DOCKERIMAGENAME_ADMINSERVER):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_UI):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_LOG):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_DB):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) \
		vmware/nginx:$(NGINXVERSION) vmware/registry:$(REGISTRYVERSION) \
		photon:$(PHOTONVERSION) | gzip > $(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tar.gz; \
	fi
	
	@if [ "$(NOTARYFLAG)" = "true" ] ; then \
		$(TARCMD) -zcvf harbor-offline-installer-$(GITTAGVERSION).tgz \
		          $(HARBORPKG)/common/templates $(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tar.gz \
				  $(HARBORPKG)/prepare $(HARBORPKG)/NOTICE \
				  $(HARBORPKG)/LICENSE $(HARBORPKG)/install.sh \
				  $(HARBORPKG)/upgrade $(HARBORPKG)/harbor_1_1_0_template \
				  $(HARBORPKG)/harbor.cfg $(HARBORPKG)/$(DOCKERCOMPOSEFILENAME) \
				  $(HARBORPKG)/$(DOCKERCOMPOSENOTARYFILENAME) ; \
	else \
		$(TARCMD) -zcvf harbor-offline-installer-$(GITTAGVERSION).tgz \
		          $(HARBORPKG)/common/templates $(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tar.gz \
				  $(HARBORPKG)/prepare $(HARBORPKG)/NOTICE \
				  $(HARBORPKG)/upgrade $(HARBORPKG)/harbor_1_1_0_template \
				  $(HARBORPKG)/LICENSE $(HARBORPKG)/install.sh \
				  $(HARBORPKG)/harbor.cfg $(HARBORPKG)/$(DOCKERCOMPOSEFILENAME) ; \
	fi

	@rm -rf $(HARBORPKG)
	@echo "Done."

pushimage:
	@echo "pushing harbor images ..."
	@$(DOCKERTAG) $(DOCKERIMAGENAME_ADMINSERVER):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKERIMAGENAME_ADMINSERVER):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKERIMAGENAME_ADMINSERVER):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKERIMAGENAME_ADMINSERVER):$(VERSIONTAG)

	@$(DOCKERTAG) $(DOCKERIMAGENAME_UI):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKERIMAGENAME_UI):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKERIMAGENAME_UI):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKERIMAGENAME_UI):$(VERSIONTAG)
	
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
	@if [ "$(NOTARYFLAG)" = "true" ] ; then \
		$(DOCKERCOMPOSECMD) -f $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME) -f $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSENOTARYFILENAME) up -d ; \
	else \
		$(DOCKERCOMPOSECMD) -f $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME) up -d ; \
	fi
	@echo "Start complete. You can visit harbor now."
	
down:
	@echo "Please make sure to set -e NOTARYFLAG=true if you are using Notary in Harbor, otherwise the Notary containers cannot be stop automaticlly."
	@while [ -z "$$CONTINUE" ]; do \
        read -r -p "Type anything but Y or y to exit. [Y/N]: " CONTINUE; \
    done ; \
    [ $$CONTINUE = "y" ] || [ $$CONTINUE = "Y" ] || (echo "Exiting."; exit 1;)
	@echo "stoping harbor instance..."
	@if [ "$(NOTARYFLAG)" = "true" ] ; then \
		$(DOCKERCOMPOSECMD) -f $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME) -f $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSENOTARYFILENAME) down -v ; \
	else \
		$(DOCKERCOMPOSECMD) -f $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME) down -v ; \
	fi	
	@echo "Done."

cleanbinary:
	@echo "cleaning binary..."
	@if [ -f $(ADMINSERVERBINARYPATH)/$(ADMINSERVERBINARYNAME) ] ; then rm $(ADMINSERVERBINARYPATH)/$(ADMINSERVERBINARYNAME) ; fi
	@if [ -f $(UIBINARYPATH)/$(UIBINARYNAME) ] ; then rm $(UIBINARYPATH)/$(UIBINARYNAME) ; fi
	@if [ -f $(JOBSERVICEBINARYPATH)/$(JOBSERVICEBINARYNAME) ] ; then rm $(JOBSERVICEBINARYPATH)/$(JOBSERVICEBINARYNAME) ; fi

cleanimage:
	@echo "cleaning image for photon..."
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_ADMINSERVER):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_UI):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_DB):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_LOG):$(VERSIONTAG)
#	- $(DOCKERRMIMAGE) -f registry:$(REGISTRYVERSION)
#	- $(DOCKERRMIMAGE) -f nginx:1.11.5

cleandockercomposefile:
	@echo "cleaning $(DOCKERCOMPOSEFILEPATH)/docker-compose.yml"
	@if [ -f $(DOCKERCOMPOSEFILEPATH)/docker-compose.yml ] ; then rm $(DOCKERCOMPOSEFILEPATH)/docker-compose.yml ; fi

cleanversiontag:
	@echo "cleaning version TAG"
	@rm -rf $(VERSIONFILEPATH)/$(VERSIONFILENAME)	
	
cleanpackage:
	@echo "cleaning harbor install package"
	@if [ -d $(BUILDPATH)/harbor ] ; then rm -rf $(BUILDPATH)/harbor ; fi
	@if [ -f $(BUILDPATH)/harbor-online-installer-$(GITTAGVERSION).tgz ] ; \
	then rm $(BUILDPATH)/harbor-online-installer-$(GITTAGVERSION).tgz ; fi
	@if [ -f $(BUILDPATH)/harbor-offline-installer-$(GITTAGVERSION).tgz ] ; \
	then rm $(BUILDPATH)/harbor-offline-installer-$(GITTAGVERSION).tgz ; fi	

.PHONY: cleanall
cleanall: cleanbinary cleanimage cleandockercomposefile cleanversiontag cleanpackage

clean: 
	@echo "  make cleanall:		remove binary, Harbor images, specific version docker-compose"
	@echo "		file, specific version tag, online and offline install package"
	@echo "  make cleanbinary:		remove ui and jobservice binary"
	@echo "  make cleanimage:		remove Harbor images"
	@echo "  make cleandockercomposefile:	remove specific version docker-compose"
	@echo "  make cleanversiontag:		cleanpackageremove specific version tag"
	@echo "  make cleanpackage:		remove online and offline install package"

all: install
