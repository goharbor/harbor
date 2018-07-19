#!/bin/bash
set -e
sudo -E -H -u \#10000 sh -c "/chartserver/chartm" #Parameters are set by ENV
set +e
