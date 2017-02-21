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
TOOLSPATH=$(BUILDPATH)/tools
CHECKENVCMD=checkenv.sh
DEVFLAG=true

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

# version prepare
VERSIONFILEPATH=$(SRCPATH)/views/sections
VERSIONFILENAME=header-content.htm
GITCMD=$(shell which git)
GITTAG=$(GITCMD) describe --tags
ifeq ($(DEVFLAG), true)        
	VERSIONTAG=dev
else        
	VERSIONTAG=$(shell $(GITTAG))
endif

check_environment:
	@$(MAKEPATH)/$(CHECKENVCMD)

build: 
	@echo "building adminserver container for photon..."
	$(DOCKERBUILD) -f $(DOCKERFILEPATH_ADMINSERVER)/$(DOCKERFILENAME_ADMINSERVER) -t $(DOCKERIMAGENAME_ADMINSERVER):$(VERSIONTAG) .
	@echo "Done."

	@echo "building ui container for photon..."
	$(DOCKERBUILD) -f $(DOCKERFILEPATH_UI)/$(DOCKERFILENAME_UI) -t $(DOCKERIMAGENAME_UI):$(VERSIONTAG) .
	@echo "Done."
	
	@echo "building jobservice container for photon..."
	$(DOCKERBUILD) -f $(DOCKERFILEPATH_JOBSERVICE)/$(DOCKERFILENAME_JOBSERVICE) -t $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) .
	@echo "Done."
	
	@echo "building log container for photon..."
	$(DOCKERBUILD) -f $(DOCKERFILEPATH_LOG)/$(DOCKERFILENAME_LOG) -t $(DOCKERIMAGENAME_LOG):$(VERSIONTAG) .
	@echo "Done."
	
cleanimage:
	@echo "cleaning image for photon..."
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_ADMINSERVER):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_UI):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_LOG):$(VERSIONTAG)
	
.PHONY: clean
clean: cleanimage

