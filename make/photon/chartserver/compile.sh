#!/bin/bash
set +e

usage(){
  echo "Usage: compile.sh <code path> <code tag> <main.go path> <binary name>"
  echo "e.g: compile.sh github.com/helm/chartmuseum v0.5.1 cmd/chartmuseum chartm"
  exit 1
}

if [ $# != 4 ]; then
  usage
fi

CODE_PATH="$1"
VERSION="$2"
MAIN_GO_PATH="$3"
BIN_NAME="$4"

#Get the source code of chartmusem
go get $CODE_PATH 

set -e

#Checkout the released tag branch
cd /go/src/$CODE_PATH
git checkout tags/$VERSION -b $VERSION 

#Install the go dep tool to restore the package dependencies
curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
dep ensure

#Compile
cd /go/src/$CODE_PATH/$MAIN_GO_PATH && go build -a -o $BIN_NAME
mv $BIN_NAME /go/bin/
