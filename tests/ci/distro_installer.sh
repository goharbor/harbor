#!/bin/bash
set -x

set -e

sudo docker system prune -af

sudo make package_online GOBUILDTAGS="include_oss include_gcs" VERSIONTAG=dev-gitaction PKGVERSIONTAG=dev-gitaction UIVERSIONTAG=dev-gitaction GOBUILDIMAGE=goharbor/golang:1.19.11 COMPILETAG=compile_golangimage NOTARYFLAG=true CHARTFLAG=true TRIVYFLAG=true HTTPPROXY= PULL_BASE_FROM_DOCKERHUB=false
sudo make package_offline GOBUILDTAGS="include_oss include_gcs" VERSIONTAG=dev-gitaction PKGVERSIONTAG=dev-gitaction UIVERSIONTAG=dev-gitaction GOBUILDIMAGE=goharbor/golang:1.19.11 COMPILETAG=compile_golangimage NOTARYFLAG=true CHARTFLAG=true TRIVYFLAG=true HTTPPROXY= PULL_BASE_FROM_DOCKERHUB=false
