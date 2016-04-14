#!/bin/bash

templates_path=`pwd`/templates/library
local_path=`pwd`/templates/static

find $templates_path -name "*.png" -exec cp {} $local_path/app_catalog_icons \;

date +%Y%m%d%H%M%S > $local_path/updatetime.html
export RSYNC_PASSWORD=E4aa42AQLh8B
rsync -aurvz --ignore-errors --delete --timeout=540 $local_path/ srystatic@120.132.52.58::srystatic/
