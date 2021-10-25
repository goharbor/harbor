#!/bin/bash
set -x

set -e

sudo make package_online GOBUILDTAGS="include_oss include_gcs" VERSIONTAG=dev-travis PKGVERSIONTAG=dev-travis UIVERSIONTAG=dev-travis GOBUILDIMAGE=golang:1.17.2 COMPILETAG=compile_golangimage BUILDBIN=true NOTARYFLAG=true CHARTFLAG=true TRIVYFLAG=true HTTPPROXY= PULL_BASE_FROM_DOCKERHUB=false
sudo make package_offline GOBUILDTAGS="include_oss include_gcs" VERSIONTAG=dev-travis PKGVERSIONTAG=dev-travis UIVERSIONTAG=dev-travis GOBUILDIMAGE=golang:1.17.2 COMPILETAG=compile_golangimage BUILDBIN=true NOTARYFLAG=true CHARTFLAG=true TRIVYFLAG=true HTTPPROXY= PULL_BASE_FROM_DOCKERHUB=false
