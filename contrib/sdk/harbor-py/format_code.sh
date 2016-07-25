#!/bin/bash

set -x

find ./ -name "*.py" |xargs sudo yapf -i
