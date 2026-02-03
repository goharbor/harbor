#!/bin/bash
set -x

set -e

sudo make package_online GOBUILDTAGS="include_oss include_gcs" VERSIONTAG=dev-gitaction PKGVERSIONTAG=dev-gitaction UIVERSIONTAG=dev-gitaction GOBUILDIMAGE=golang:1.25.6 COMPILETAG=compile_golangimage TRIVYFLAG=true EXPORTERFLAG=true HTTPPROXY= PULL_BASE_FROM_DOCKERHUB=false
sudo make package_offline GOBUILDTAGS="include_oss include_gcs" VERSIONTAG=dev-gitaction PKGVERSIONTAG=dev-gitaction UIVERSIONTAG=dev-gitaction GOBUILDIMAGE=golang:1.25.6 COMPILETAG=compile_golangimage TRIVYFLAG=true EXPORTERFLAG=true HTTPPROXY= PULL_BASE_FROM_DOCKERHUB=false
