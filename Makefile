# Makefile for Harbor project
#	
# Targets:
#
# all:			prepare env, compile binarys, build images and install images 
# prepare: 		prepare env
# compile: 		compile ui and jobservice code
# compile_buildgolangimage:
#			compile local building golang image
#			forexample : make compile_buildgolangimage -e \
#							GOBUILDIMAGE=harborgo:1.6.2
# compile_golangimage:
#			compile from golang image
#			for example: make compile_golangimage -e GOBUILDIMAGE= \
#							harborgo:1.6.2
# compile_ui, compile_jobservice: compile specific binary
#
# build: 		build Harbor docker images (defuault: build_photon)
#			for example: make build -e BASEIMAGE=photon
# build_photon:	build Harbor docker images from photon bsaeimage
# build_ubuntu: build Harbor docker images from ubuntu baseimage
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
# cleanbinary:	remove ui and jobservice binary
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
GOBASEPATH=/go/src/github.com/vmware
CHECKENVCMD=checkenv.sh
BASEIMAGE=photon
COMPILETAG=compile_normal
REGISTRYSERVER=
REGISTRYPROJECTNAME=vmware
DEVFLAG=true

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
GOBUILDPATH_UI=$(GOBUILDPATH)/src/ui
GOBUILDPATH_JOBSERVICE=$(GOBUILDPATH)/src/jobservice
GOBUILDMAKEPATH=$(GOBUILDPATH)/make
GOBUILDMAKEPATH_UI=$(GOBUILDMAKEPATH)/dev/ui
GOBUILDMAKEPATH_JOBSERVICE=$(GOBUILDMAKEPATH)/dev/jobservice
GOLANGDOCKERFILENAME=Dockerfile.golang

# binary 
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
MAKEFILEPATH_UBUNTU=$(MAKEPATH)/ubuntu

# common dockerfile
DOCKERFILEPATH_COMMON=$(MAKEPATH)/common
DOCKERFILEPATH_DB=$(DOCKERFILEPATH_COMMON)/db
DOCKERFILENAME_DB=Dockerfile

# docker image name
DOCKERIMAGENAME_UI=vmware/harbor-ui
DOCKERIMAGENAME_JOBSERVICE=vmware/harbor-jobservice
DOCKERIMAGENAME_LOG=vmware/harbor-log
DOCKERIMAGENAME_DB=vmware/harbor-db


# docker-compose files
DOCKERCOMPOSEFILEPATH=$(MAKEPATH)
DOCKERCOMPOSETPLFILENAME=docker-compose.tpl
DOCKERCOMPOSEFILENAME=docker-compose.yml

# version prepare
VERSIONFILEPATH=$(SRCPATH)/ui/views/sections
VERSIONFILENAME=header-content.htm
GITCMD=$(shell which git)
GITTAG=$(GITCMD) describe --tags
ifeq ($(DEVFLAG), true)        
	VERSIONTAG=dev
else        
	VERSIONTAG=$(shell $(GITTAG))
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
	@if [ "$(DEVFLAG)" = "false" ] ; then \
		$(SEDCMD) -i 's/version=\"{{.Version}}\"/version=\"$(VERSIONTAG)\"/' -i $(VERSIONFILEPATH)/$(VERSIONFILENAME) ; \
	fi
	
check_environment:
	@$(MAKEPATH)/$(CHECKENVCMD)

compile_ui:
	@echo "compiling binary for ui..."
	@$(GOBUILD) -o $(UIBINARYPATH)/$(UIBINARYNAME) $(UISOURCECODE)
	@echo "Done."
	
compile_jobservice:
	@echo "compiling binary for jobservice..."
	@$(GOBUILD) -o $(JOBSERVICEBINARYPATH)/$(JOBSERVICEBINARYNAME) $(JOBSERVICESOURCECODE)
	@echo "Done."
	
compile_normal: compile_ui compile_jobservice

compile_buildgolangimage:
	@echo "compiling golang image for harbor ..."
	@$(DOCKERBUILD) -t $(GOBUILDIMAGE) -f $(TOOLSPATH)/$(GOLANGDOCKERFILENAME) .
	@echo "Done."

compile_golangimage:
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
	@$(MAKEPATH)/$(PREPARECMD) -conf $(CONFIGPATH)/$(CONFIGFILE)
	
build_common: version
	@echo "buildging db container for photon..."
	@cd $(DOCKERFILEPATH_DB) && $(DOCKERBUILD) -f $(DOCKERFILENAME_DB) -t $(DOCKERIMAGENAME_DB):$(VERSIONTAG) .
	@echo "Done."

build_photon: build_common
	make -f $(MAKEFILEPATH_PHOTON)/Makefile build -e DEVFLAG=$(DEVFLAG)
	
build_ubuntu: build_common
	make -f $(MAKEFILEPATH_UBUNTU)/Makefile build -e DEVFLAG=$(DEVFLAG)
	
build: build_$(BASEIMAGE)
	
modify_composefile: 
	@echo "preparing docker-compose file..."
	@cp $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSETPLFILENAME) $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)
	@$(SEDCMD) -i 's/image\: vmware.*/&:$(VERSIONTAG)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)
	
install: compile build prepare modify_composefile
	@echo "loading harbor images..."
	@$(DOCKERCOMPOSECMD) -f $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME) up -d
	@echo "Install complete. You can visit harbor now."
	
package_online: modify_composefile
	@echo "packing online package ..."
	@cp -r make $(HARBORPKG)
	@if [ -n "$(REGISTRYSERVER)" ] ; then \
		$(SEDCMD) -i 's/image\: vmware/image\: $(REGISTRYSERVER)\/$(REGISTRYPROJECTNAME)/' \
		$(HARBORPKG)/docker-compose.yml ; \
	fi
	@cp LICENSE $(HARBORPKG)/LICENSE
	@cp NOTICE $(HARBORPKG)/NOTICE
	@$(TARCMD) -zcvf harbor-online-installer-$(VERSIONTAG).tgz \
	          --exclude=$(HARBORPKG)/common/db --exclude=$(HARBORPKG)/common/config\
			  --exclude=$(HARBORPKG)/common/log --exclude=$(HARBORPKG)/ubuntu \
			  --exclude=$(HARBORPKG)/photon --exclude=$(HARBORPKG)/kubernetes \
			  --exclude=$(HARBORPKG)/dev --exclude=$(DOCKERCOMPOSETPLFILENAME) \
			  --exclude=$(HARBORPKG)/checkenv.sh \
			  --exclude=$(HARBORPKG)/jsminify.sh \
			  --exclude=$(HARBORPKG)/pushimage.sh \
			  $(HARBORPKG)
			
	@rm -rf $(HARBORPKG)
	@echo "Done."
	
package_offline: compile build modify_composefile
	@echo "packing offline package ..."
	@cp -r make $(HARBORPKG)
	
	@cp LICENSE $(HARBORPKG)/LICENSE
	@cp NOTICE $(HARBORPKG)/NOTICE
			
	@echo "pulling nginx and registry..."
	@$(DOCKERPULL) registry:2.5.0
	@$(DOCKERPULL) nginx:1.11.5
	
	@echo "saving harbor docker image"
	@$(DOCKERSAVE) -o $(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tgz \
		$(DOCKERIMAGENAME_UI):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_LOG):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_DB):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) \
		nginx:1.11.5 registry:2.5.0 photon:1.0

	@$(TARCMD) -zcvf harbor-offline-installer-$(VERSIONTAG).tgz \
	          --exclude=$(HARBORPKG)/common/db --exclude=$(HARBORPKG)/common/config\
			  --exclude=$(HARBORPKG)/common/log --exclude=$(HARBORPKG)/ubuntu \
			  --exclude=$(HARBORPKG)/photon --exclude=$(HARBORPKG)/kubernetes \
			  --exclude=$(HARBORPKG)/dev --exclude=$(DOCKERCOMPOSETPLFILENAME) \
			  --exclude=$(HARBORPKG)/checkenv.sh \
			  --exclude=$(HARBORPKG)/jsminify.sh \
			  --exclude=$(HARBORPKG)/pushimage.sh \
			  $(HARBORPKG)
	
	@rm -rf $(HARBORPKG)
	@echo "Done."

pushimage:
	@echo "pushing harbor images ..."
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
	@$(DOCKERCOMPOSECMD) -f $(DOCKERCOMPOSEFILEPATH)/docker-compose.yml up -d
	@echo "Start complete. You can visit harbor now."
	
down:
	@echo "stoping harbor instance..."
	@$(DOCKERCOMPOSECMD) -f $(DOCKERCOMPOSEFILEPATH)/docker-compose.yml down
	@echo "Done."

cleanbinary:
	@echo "cleaning binary..."
	@if [ -f $(UIBINARYPATH)/$(UIBINARYNAME) ] ; then rm $(UIBINARYPATH)/$(UIBINARYNAME) ; fi
	@if [ -f $(JOBSERVICEBINARYPATH)/$(JOBSERVICEBINARYNAME) ] ; then rm $(JOBSERVICEBINARYPATH)/$(JOBSERVICEBINARYNAME) ; fi

cleanimage:
	@echo "cleaning image for photon..."
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_UI):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_DB):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_LOG):$(VERSIONTAG)
#	- $(DOCKERRMIMAGE) -f registry:2.5.0
#	- $(DOCKERRMIMAGE) -f nginx:1.11.5

cleandockercomposefile:
	@echo "cleaning $(DOCKERCOMPOSEFILEPATH)/docker-compose.yml"
	@if [ -f $(DOCKERCOMPOSEFILEPATH)/docker-compose.yml ] ; then rm $(DOCKERCOMPOSEFILEPATH)/docker-compose.yml ; fi

cleanversiontag:
	@echo "cleaning version TAG"
	@$(SEDCMD) -i 's/version=\"$(VERSIONTAG)\"/version=\"{{.Version}}\"/' -i $(VERSIONFILEPATH)/$(VERSIONFILENAME)
	
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
