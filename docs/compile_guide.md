## Introduction

This guide shows how to compile binary, build images and install Harbor instance from source code via make commands.

## Step 1: Prepare Your System for Building Harbor

Harbor is deployed as several Docker containers and most of the code compiled by go language. The target host requires Python, Docker, Docker Compose and golang develop environment to be installed. 

Requirement:

Software              | Required Version
----------------------|--------------------------
docker                | 1.10.0 +
docker-compose        | 1.7.1 +
python                | 2.7 +
git                   | 1.9.1 +
make                  | 3.81 +
golang*               | 1.6.0 +
 *optional


## Step 2: Getting the Source Code

   ```sh
      $ git clone https://github.com/vmware/harbor
   ```

## Step 3: Resolving Dependencies
Compile Harbor source code by local golang environment needs LDAP develop package and you'll have to do it manually. If you want to compile source code by golang image, can skip this section. 

For PhotonOS:

   ```sh
      $ tdnf install -y sed apr-util-ldap
   ```

For Ubuntu:

   ```sh
      $ apt-get update && apt-get install -y libldap2-dev
   ```

Other platforms please consult the relevant documentation for LDAP package installation. 

## Step 4: Build and Install

### Configuration

Edit the file **make/harbor.cfg**, make necessary configuration changes such as hostname, admin password and mail server. Refer to [Installation and Configuration Guide](installation_guide.md#configuring-harbor) for more info. 

   ```sh
      $ cd harbor
      $ vi make/harbor.cfg
   ```
   
### Compile, Build and Install

Support 3 code compile method: golang image compile, local golang compile and developer manual compile.

#### I. Compile Code with Golang Image, then Automation Build and Install 

* Build Compile Golang Image

   ```sh
      $ make compile_buildgolangimage -e GOBUILDIMAGE=harborgo:1.6.2
   ```

*  Automation Build and Install

   ```sh
      $ make install -e GOBUILDIMAGE=harborgo:1.6.2 COMPILETAG=compile_golangimage
   ```

#### II. Compile Code with Local Golang, then Automation Build and Install 

* Move Code to $GOPATH

   ```sh
      $ mkdir $GOPATH/src/github.com/vmware/
      $ cd ..
      $ mv harbor $GOPATH/src/github.com/vmware/.
   ```

*  Automation Build and Install

   ```sh
      $ cd $GOPATH/src/github.com/vmware/harbor
      $ make install
   ```
   
#### III. Manual Build and Install (Compatible with Prior Versions)

   ```sh
      $ cd make
   
      $ ./prepare
      Generated configuration file: ./config/ui/env
      Generated configuration file: ./config/ui/app.conf
      Generated configuration file: ./config/registry/config.yml
      Generated configuration file: ./config/db/env
      ...
   
      $ cd dev
      
      $ docker-compose up -d
   ```

### Install Success

You can get this message from shell after successful complete Harbor installs.

   ```sh
      ...
      ✔ ----Harbor has been installed and started successfully.----

      Now you should be able to visit the admin portal at http://$YOURIP. 
      For more details, please visit https://github.com/vmware/harbor .
   ```

Refer to [Installation and Configuration Guide](installation_guide.md#managing-harbors-lifecycle) for more info.   

## Attachments
* Using the Makefile

Makefile is a special format file that together with the make utility will help developer to automagically build and manage Harbor projects.
At the top of the makefile, there are several user-configurable parameters designed to enable the Makefile to be easily portable.

Variable           | Description
-------------------|-------------
BASEIMAGE          | Container base image, default: photon
DEVFLAG            | Build model flag, default: dev
COMPILETAG         | Compile model flag, default: compile_normal (local golang build)
REGISTRYSERVER     | Remote registry server address
REGISTRYUSER       | Remote registry server user name
REGISTRYPASSWORD   | Remote registry server user password
REGISTRYPROJECTNAME| Project name on remote registry server

There are also a variety of rules that help with project management and debugging...

Rule                | Description
--------------------|-------------
all                 | prepare env, compile binarys, build images and install images 
prepare             | prepare env
compile             | compile ui and jobservice code
compile_golangimage | compile local golang image
compile_ui          | compile ui binary
compile_jobservice  | compile jobservice binary
build               | build Harbor docker images (defuault  |   build_photon)
build_photon        | build Harbor docker images from photon bsaeimage
build_ubuntu        | build Harbor docker images from ubuntu baseimage
install             | compile binarys, build images, prepare specific version composefile and startup Harbor instance
start               | startup Harbor instance 
down                | shutdown Harbor instance
package_online      | prepare online install package
package_offline     | prepare offline install package
pushimage           | push Harbor images to specific registry server
clean all           | remove binary, Harbor images, specific version docker-compose file, specific version tag and online/offline install package
cleanbinary         | remove ui and jobservice binary
cleanimage          | remove Harbor images 
cleandockercomposefile  | remove specific version docker-compose 
cleanversiontag     | remove specific version tag
cleanpackage        | remove online/offline install package

#### EXAMPLE:

#### compile from golang image: 

   ```sh
      $ make compile_golangimage -e GOBUILDIMAGE= [$YOURIMAGE]

   ```

#### build Harbor docker images form ubuntu

   ```sh
      $ make build -e BASEIMAGE=ubuntu

   ```

#### push Harbor images to specific registry server

   ```sh
      $ make pushimage -e DEVFLAG=false REGISTRYSERVER=[$SERVERADDRESS] REGISTRYUSER=[$USERNAME] REGISTRYPASSWORD=[$PASSWORD] REGISTRYPROJECTNAME=[$PROJECTNAME]

   ```

   note**: need add "/" on end of REGISTRYSERVER. If not setting REGISTRYSERVER will push images directly to dockerhub.


   ```sh
      $ make pushimage -e DEVFLAG=false REGISTRYUSER=[$USERNAME] REGISTRYPASSWORD=[$PASSWORD] REGISTRYPROJECTNAME=[$PROJECTNAME]

   ```

#### clean specific version binarys and images

   ```sh
      $ make clean -e VERSIONTAG=[TAG]

   ```
   note**: If commit new code to github, the git commit TAG will change. Better use this command clean previous images and files for specific TAG. 

#### By default DEVFLAG=true, if you want to release new version of Harbor, should setting the flag to false.

   ```sh
      $ make XXXX -e DEVFLAG=false

   ```
   
## Links

## Comments

