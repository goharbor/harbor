#!/bin/bash
set -x

set -e

sudo make package_online VERSIONTAG=dev-travis PKGVERSIONTAG=dev-travis UIVERSIONTAG=dev-travis GOBUILDIMAGE=golang:1.13.8 COMPILETAG=compile_golangimage NOTARYFLAG=true CLAIRFLAG=true CHARTFLAG=true TRIVYFLAG=true HTTPPROXY=
sudo make package_offline VERSIONTAG=dev-travis PKGVERSIONTAG=dev-travis UIVERSIONTAG=dev-travis GOBUILDIMAGE=golang:1.13.8 COMPILETAG=compile_golangimage NOTARYFLAG=true CLAIRFLAG=true CHARTFLAG=true TRIVYFLAG=true HTTPPROXY=
