#!/bin/bash
IP=$1
HARBOR_VERSION=$2

robot -v ip:$IP  -v ip1: -v HARBOR_PASSWORD:Harbor12345 /drone/tests/robot-cases/Group1-Nightly/Setup.robot
cd /drone/tests/robot-cases/Group3-Upgrade
python ./prepare.py -e $IP -v $HARBOR_VERSION -l /drone/tests/apitests/python/
