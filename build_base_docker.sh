#!/usr/bin/env bash

DOCKERCMD=docker
CICDHOST=cicd.harbor.bitsf.xin
DOCKERFILEPATH=make/photon

${DOCKERCMD} login ${CICDHOST} || exit 2
for name in chartserver clair core db jobservice log nginx portal prepare redis registry registryctl; do
	echo building $name base docker image
	$DOCKERCMD build -f $DOCKERFILEPATH/$name/Dockerfile-base -t $CICDHOST/harbor-depend/$name:base . && \
	$DOCKERCMD push $CICDHOST/harbor-depend/$name:base
	if [ "$?" != "0" ]; then exit 1; fi
done
