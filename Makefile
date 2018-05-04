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

# parameters
REGISTRYSERVER=
REGISTRYPROJECTNAME=vmware
DEVFLAG=true
NOTARYFLAG=false
CLAIRFLAG=false
HTTPPROXY=
REBUILDCLARITYFLAG=false
NEWCLARITYVERSION=
BUILDBIN=false
MIGRATORFLAG=false

# version prepare
# for docker image tag
VERSIONTAG=dev
# for harbor package name
PKGVERSIONTAG=dev
# for harbor about dialog
UIVERSIONTAG=dev
VERSIONFILEPATH=$(CURDIR)
VERSIONFILENAME=UIVERSION

#versions
REGISTRYVERSION=v2.6.2
NGINXVERSION=$(VERSIONTAG)
PHOTONVERSION=1.0
NOTARYVERSION=v0.5.1
MARIADBVERSION=$(VERSIONTAG)
CLAIRVERSION=v2.0.1
CLAIRDBVERSION=$(VERSIONTAG)
MIGRATORVERSION=v1.5.0
REDISVERSION=$(VERSIONTAG)

#clarity parameters
CLARITYIMAGE=vmware/harbor-clarity-ui-builder[:tag]
CLARITYSEEDPATH=/harbor_src
CLARITYUTPATH=${CLARITYSEEDPATH}/ui_ng/lib
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
ADMINSERVERBINARYPATH=$(MAKEDEVPATH)/adminserver
ADMINSERVERBINARYNAME=harbor_adminserver
UIBINARYPATH=$(MAKEDEVPATH)/ui
UIBINARYNAME=harbor_ui
JOBSERVICEBINARYPATH=$(MAKEDEVPATH)/jobservice
JOBSERVICEBINARYNAME=harbor_jobservice

# configfile
CONFIGPATH=$(MAKEPATH)
CONFIGFILE=harbor.cfg

# prepare parameters
PREPAREPATH=$(TOOLSPATH)
PREPARECMD=prepare
PREPARECMD_PARA=--conf $(CONFIGPATH)/$(CONFIGFILE)
ifeq ($(NOTARYFLAG), true)
	PREPARECMD_PARA+= --with-notary
endif
ifeq ($(CLAIRFLAG), true)
	PREPARECMD_PARA+= --with-clair
endif

# makefile
MAKEFILEPATH_PHOTON=$(MAKEPATH)/photon

# common dockerfile
DOCKERFILEPATH_COMMON=$(MAKEPATH)/common
DOCKERFILE_CLARITY=$(MAKEPATH)/dev/nodeclarity/Dockerfile

# docker image name
DOCKERIMAGENAME_ADMINSERVER=vmware/harbor-adminserver
DOCKERIMAGENAME_UI=vmware/harbor-ui
DOCKERIMAGENAME_JOBSERVICE=vmware/harbor-jobservice
DOCKERIMAGENAME_LOG=vmware/harbor-log
DOCKERIMAGENAME_DB=vmware/harbor-db
DOCKERIMAGENAME_CLARITY=vmware/harbor-clarity-ui-builder

# docker-compose files
DOCKERCOMPOSEFILEPATH=$(MAKEPATH)
DOCKERCOMPOSETPLFILENAME=docker-compose.tpl
DOCKERCOMPOSEFILENAME=docker-compose.yml
DOCKERCOMPOSENOTARYTPLFILENAME=docker-compose.notary.tpl
DOCKERCOMPOSENOTARYFILENAME=docker-compose.notary.yml
DOCKERCOMPOSECLAIRTPLFILENAME=docker-compose.clair.tpl
DOCKERCOMPOSECLAIRFILENAME=docker-compose.clair.yml

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
DOCKERSAVE_PARA=$(DOCKERIMAGENAME_ADMINSERVER):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_UI):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_LOG):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_DB):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) \
		vmware/redis-photon:$(REDISVERSION) \
		vmware/nginx-photon:$(NGINXVERSION) vmware/registry-photon:$(REGISTRYVERSION)-$(VERSIONTAG) \
		vmware/photon:$(PHOTONVERSION)
PACKAGE_OFFLINE_PARA=-zcvf harbor-offline-installer-$(PKGVERSIONTAG).tgz \
		          $(HARBORPKG)/common/templates $(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tar.gz \
				  $(HARBORPKG)/prepare $(HARBORPKG)/NOTICE \
				  $(HARBORPKG)/LICENSE $(HARBORPKG)/install.sh \
				  $(HARBORPKG)/harbor.cfg $(HARBORPKG)/$(DOCKERCOMPOSEFILENAME) \
				  $(HARBORPKG)/ha
PACKAGE_ONLINE_PARA=-zcvf harbor-online-installer-$(PKGVERSIONTAG).tgz \
		          $(HARBORPKG)/common/templates $(HARBORPKG)/prepare \
				  $(HARBORPKG)/LICENSE $(HARBORPKG)/NOTICE \
				  $(HARBORPKG)/install.sh $(HARBORPKG)/$(DOCKERCOMPOSEFILENAME) \
				  $(HARBORPKG)/harbor.cfg $(HARBORPKG)/ha
DOCKERCOMPOSE_LIST=-f $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)

ifeq ($(NOTARYFLAG), true)
	DOCKERSAVE_PARA+= vmware/notary-server-photon:$(NOTARYVERSION)-$(VERSIONTAG) vmware/notary-signer-photon:$(NOTARYVERSION)-$(VERSIONTAG)
	PACKAGE_OFFLINE_PARA+= $(HARBORPKG)/$(DOCKERCOMPOSENOTARYFILENAME)
	PACKAGE_ONLINE_PARA+= $(HARBORPKG)/$(DOCKERCOMPOSENOTARYFILENAME)
	DOCKERCOMPOSE_LIST+= -f $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSENOTARYFILENAME)
endif
ifeq ($(CLAIRFLAG), true)
	DOCKERSAVE_PARA+= vmware/clair-photon:$(CLAIRVERSION)-$(VERSIONTAG)
	PACKAGE_OFFLINE_PARA+= $(HARBORPKG)/$(DOCKERCOMPOSECLAIRFILENAME)
	PACKAGE_ONLINE_PARA+= $(HARBORPKG)/$(DOCKERCOMPOSECLAIRFILENAME)
	DOCKERCOMPOSE_LIST+= -f $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSECLAIRFILENAME)
endif
ifeq ($(MIGRATORFLAG), true)
	DOCKERSAVE_PARA+= vmware/harbor-migrator:$(MIGRATORVERSION)
endif

version:
	@printf $(UIVERSIONTAG) > $(VERSIONFILEPATH)/$(VERSIONFILENAME);

check_environment:
	@$(MAKEPATH)/$(CHECKENVCMD)

compile_clarity:
	@echo "compiling binary for clarity ui..."
	@if [ "$(HTTPPROXY)" != "" ] ; then \
		$(DOCKERCMD) run --rm -v $(BUILDPATH)/src:$(CLARITYSEEDPATH) $(CLARITYIMAGE) $(SHELL) $(CLARITYBUILDSCRIPT) -p $(HTTPPROXY); \
	else \
		$(DOCKERCMD) run --rm -v $(BUILDPATH)/src:$(CLARITYSEEDPATH) $(CLARITYIMAGE) $(SHELL) $(CLARITYBUILDSCRIPT); \
	fi
	@echo "Done."

compile_golangimage: compile_clarity
	@echo "compiling binary for adminserver (golang image)..."
	@echo $(GOBASEPATH)
	@echo $(GOBUILDPATH)
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATH) -w $(GOBUILDPATH_ADMINSERVER) $(GOBUILDIMAGE) $(GOIMAGEBUILD) -o $(GOBUILDMAKEPATH_ADMINSERVER)/$(ADMINSERVERBINARYNAME)
	@echo "Done."

	@echo "compiling binary for ui (golang image)..."
	@echo $(GOBASEPATH)
	@echo $(GOBUILDPATH)
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATH) -w $(GOBUILDPATH_UI) $(GOBUILDIMAGE) $(GOIMAGEBUILD) -o $(GOBUILDMAKEPATH_UI)/$(UIBINARYNAME)
	@echo "Done."

	@echo "compiling binary for jobservice (golang image)..."
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATH) -w $(GOBUILDPATH_JOBSERVICE) $(GOBUILDIMAGE) $(GOIMAGEBUILD) -o $(GOBUILDMAKEPATH_JOBSERVICE)/$(JOBSERVICEBINARYNAME)
	@echo "Done."

compile:check_environment compile_golangimage
	
prepare:
	@echo "preparing..."
	@$(MAKEPATH)/$(PREPARECMD) $(PREPARECMD_PARA)

build:
	make -f $(MAKEFILEPATH_PHOTON)/Makefile build -e DEVFLAG=$(DEVFLAG) -e MARIADBVERSION=$(MARIADBVERSION) \
	 -e REGISTRYVERSION=$(REGISTRYVERSION) -e NGINXVERSION=$(NGINXVERSION) -e NOTARYVERSION=$(NOTARYVERSION) \
	 -e CLAIRVERSION=$(CLAIRVERSION) -e CLAIRDBVERSION=$(CLAIRDBVERSION) -e VERSIONTAG=$(VERSIONTAG) \
	 -e BUILDBIN=$(BUILDBIN) -e REDISVERSION=$(REDISVERSION)

modify_composefile: modify_composefile_notary modify_composefile_clair
	@echo "preparing docker-compose file..."
	@cp $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSETPLFILENAME) $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)
	@cp $(DOCKERCOMPOSEFILEPATH)/ha/$(DOCKERCOMPOSETPLFILENAME) $(DOCKERCOMPOSEFILEPATH)/ha/$(DOCKERCOMPOSEFILENAME)
	@$(SEDCMD) -i 's/__version__/$(VERSIONTAG)/g' $(DOCKERCOMPOSEFILEPATH)/ha/$(DOCKERCOMPOSEFILENAME)
	@$(SEDCMD) -i 's/__version__/$(VERSIONTAG)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)
	@$(SEDCMD) -i 's/__postgresql_version__/$(VERSIONTAG)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)
	@$(SEDCMD) -i 's/__reg_version__/$(REGISTRYVERSION)-$(VERSIONTAG)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)
	@$(SEDCMD) -i 's/__reg_version__/$(REGISTRYVERSION)-$(VERSIONTAG)/g' $(DOCKERCOMPOSEFILEPATH)/ha/$(DOCKERCOMPOSEFILENAME)
	@$(SEDCMD) -i 's/__nginx_version__/$(NGINXVERSION)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)
	@$(SEDCMD) -i 's/__nginx_version__/$(NGINXVERSION)/g' $(DOCKERCOMPOSEFILEPATH)/ha/$(DOCKERCOMPOSEFILENAME)
	@$(SEDCMD) -i 's/__redis_version__/$(REDISVERSION)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)

modify_composefile_notary:
	@echo "preparing docker-compose notary file..."
	@cp $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSENOTARYTPLFILENAME) $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSENOTARYFILENAME)
	@$(SEDCMD) -i 's/__notary_version__/$(NOTARYVERSION)-$(VERSIONTAG)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSENOTARYFILENAME)
	@$(SEDCMD) -i 's/__mariadb_version__/$(MARIADBVERSION)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSENOTARYFILENAME)

modify_composefile_clair:
	@echo "preparing docker-compose clair file..."
	@cp $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSECLAIRTPLFILENAME) $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSECLAIRFILENAME)
	@$(SEDCMD) -i 's/__postgresql_version__/$(CLAIRDBVERSION)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSECLAIRFILENAME)
	@$(SEDCMD) -i 's/__clair_version__/$(CLAIRVERSION)-$(VERSIONTAG)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSECLAIRFILENAME)
	@cp $(DOCKERCOMPOSEFILEPATH)/ha/$(DOCKERCOMPOSECLAIRTPLFILENAME) $(DOCKERCOMPOSEFILEPATH)/ha/$(DOCKERCOMPOSECLAIRFILENAME)
	@$(SEDCMD) -i 's/__clair_version__/$(CLAIRVERSION)-$(VERSIONTAG)/g' $(DOCKERCOMPOSEFILEPATH)/ha/$(DOCKERCOMPOSECLAIRFILENAME)

modify_sourcefiles:
	@echo "change mode of source files."
	@chmod 600 $(MAKEPATH)/common/templates/notary/notary-signer.key
	@chmod 600 $(MAKEPATH)/common/templates/notary/notary-signer.crt
	@chmod 600 $(MAKEPATH)/common/templates/notary/notary-signer-ca.crt
	@chmod 600 $(MAKEPATH)/common/templates/ui/private_key.pem
	@chmod 600 $(MAKEPATH)/common/templates/registry/root.crt

install: compile version build modify_sourcefiles prepare modify_composefile start

package_online: modify_composefile
	@echo "packing online package ..."
	@cp -r make $(HARBORPKG)
	@if [ -n "$(REGISTRYSERVER)" ] ; then \
		$(SEDCMD) -i 's/image\: vmware/image\: $(REGISTRYSERVER)\/$(REGISTRYPROJECTNAME)/' \
		$(HARBORPKG)/docker-compose.yml ; \
		$(SEDCMD) -i 's/image\: vmware/image\: $(REGISTRYSERVER)\/$(REGISTRYPROJECTNAME)/' \
		$(HARBORPKG)/ha/docker-compose.yml ; \
	fi
	@cp LICENSE $(HARBORPKG)/LICENSE
	@cp NOTICE $(HARBORPKG)/NOTICE
	
	@$(TARCMD) $(PACKAGE_ONLINE_PARA)
	@rm -rf $(HARBORPKG)
	@echo "Done."
	
package_offline: compile version build modify_sourcefiles modify_composefile
	@echo "packing offline package ..."
	@cp -r make $(HARBORPKG)
	@cp LICENSE $(HARBORPKG)/LICENSE
	@cp NOTICE $(HARBORPKG)/NOTICE
	@cp $(HARBORPKG)/photon/db/registry.sql $(HARBORPKG)/ha/
	
	@if [ "$(MIGRATORFLAG)" = "true" ] ; then \
		echo "pulling Harbor migrator..."; \
		$(DOCKERPULL) vmware/harbor-migrator:$(MIGRATORVERSION); \
	fi

	@echo "saving harbor docker image"
	@$(DOCKERSAVE) $(DOCKERSAVE_PARA) > $(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tar
	@gzip $(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tar

	@$(TARCMD) $(PACKAGE_OFFLINE_PARA)
	@rm -rf $(HARBORPKG)
	@echo "Done."

refresh_clarity_builder:
	@if [ "$(REBUILDCLIATRYFLAG)" = "true" ] ; then \
		echo "set http proxy.."; \
		if [ "$(HTTPPROXY)" != "" ] ; then \
			$(SEDCMD) -i 's/__proxy__/--proxy $(HTTPPROXY)/g' $(DOCKERFILE_CLARITY) ; \
		else \
			$(SEDCMD) -i 's/__proxy__/ /g' $(DOCKERFILE_CLARITY) ; \
		fi ; \
		echo "build new clarity image.."; \
		$(DOCKERBUILD) -f $(DOCKERFILE_CLARITY) -t $(DOCKERIMAGENAME_CLARITY):$(NEWCLARITYVERSION) . ; \
		echo "push clarity image.."; \
		$(DOCKERTAG) $(DOCKERIMAGENAME_CLARITY):$(NEWCLARITYVERSION) $(DOCKERIMAGENAME_CLARITY):$(NEWCLARITYVERSION); \
		$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKERIMAGENAME_CLARITY):$(NEWCLARITYVERSION) \
			$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER); \
		echo "remove local clarity image.."; \
		$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKERIMAGENAME_ADMINSERVER):$(NEWCLARITYVERSION); \
	fi

run_clarity_ut:
	@echo "run clarity ut ..."
	@$(DOCKERCMD) run --rm -v $(UINGPATH):$(CLARITYSEEDPATH) -v $(BUILDPATH)/tests:$(CLARITYSEEDPATH)/tests $(CLARITYIMAGE) $(SHELL) $(CLARITYSEEDPATH)/tests/run-clarity-ut.sh

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
	@$(DOCKERCOMPOSECMD) $(DOCKERCOMPOSE_LIST) up -d
	@echo "Start complete. You can visit harbor now."

down:
	@echo "Please make sure to set -e NOTARYFLAG=true/CLAIRFLAG=true if you are using Notary/CLAIR in Harbor, otherwise the Notary/CLAIR containers cannot be stop automaticlly."
	@while [ -z "$$CONTINUE" ]; do \
        read -r -p "Type anything but Y or y to exit. [Y/N]: " CONTINUE; \
    done ; \
    [ $$CONTINUE = "y" ] || [ $$CONTINUE = "Y" ] || (echo "Exiting."; exit 1;)
	@echo "stoping harbor instance..."
	@$(DOCKERCOMPOSECMD) $(DOCKERCOMPOSE_LIST) down -v
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

cleandockercomposefile:
	@echo "cleaning $(DOCKERCOMPOSEFILEPATH)/docker-compose.yml"
	@if [ -f $(DOCKERCOMPOSEFILEPATH)/docker-compose.yml ] ; then rm $(DOCKERCOMPOSEFILEPATH)/docker-compose.yml ; fi

cleanversiontag:
	@echo "cleaning version TAG"
	@rm -rf $(VERSIONFILEPATH)/$(VERSIONFILENAME)

cleanpackage:
	@echo "cleaning harbor install package"
	@if [ -d $(BUILDPATH)/harbor ] ; then rm -rf $(BUILDPATH)/harbor ; fi
	@if [ -f $(BUILDPATH)/harbor-online-installer-$(VERSIONTAG).tgz ] ; \
	then rm $(BUILDPATH)/harbor-online-installer-$(VERSIONTAG).tgz ; fi
	@if [ -f $(BUILDPATH)/harbor-offline-installer-$(VERSIONTAG).tgz ] ; \
	then rm $(BUILDPATH)/harbor-offline-installer-$(VERSIONTAG).tgz ; fi

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
