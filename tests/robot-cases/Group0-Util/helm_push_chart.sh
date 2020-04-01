#!/bin/bash

IP=$1
USER=$2
PWD=$3
CHART_FILE=$4
ARCHIVE=$5
PROJECT=$6
REPOSITORY=$7
VERSION=$8

echo $IP


export HELM_EXPERIMENTAL_OCI=1
wget $CHART_FILE
echo "========================"
echo ${CHART_FILE##*/}
echo "========================"
tar xvf ${CHART_FILE##*/}
helm3 registry login $IP -u $USER -p $PWD
helm3 chart save $ARCHIVE $IP/$PROJECT/$REPOSITORY
helm3 chart push $IP/$PROJECT/$REPOSITORY:$VERSION

