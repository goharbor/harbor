#!/bin/bash

set -e

sudo make package_online VERSIONTAG=dev-travis PKGVERSIONTAG=dev-travis UIVERSIONTAG=dev-travis GOBUILDIMAGE=golang:1.12.12 COMPILETAG=compile_golangimage NOTARYFLAG=true CLAIRFLAG=true MIGRATORFLAG=false CHARTFLAG=true HTTPPROXY=
sudo make package_offline VERSIONTAG=dev-travis PKGVERSIONTAG=dev-travis UIVERSIONTAG=dev-travis GOBUILDIMAGE=golang:1.12.12 COMPILETAG=compile_golangimage NOTARYFLAG=true CLAIRFLAG=true MIGRATORFLAG=false CHARTFLAG=true HTTPPROXY=
