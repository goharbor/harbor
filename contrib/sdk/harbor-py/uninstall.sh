#!/bin/sh

set -x
set -e

python setup.py install --record files.txt
cat files.txt | xargs rm -rf
