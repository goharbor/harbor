#!/bin/sh
# Copyright Project Harbor Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

set -e

# Only give socker support by default, use bash arguments to add dockerd parameters
# Use unix:///var/run/docker-local.sock to avoid collision with /var/run/docker.sock

# no arguments passed
# or first arg is `-f` or `--some-option`
IP=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`

if [ "$#" -eq 0 -o "${1#-}" != "$1" ]; then
	# add our default arguments
	set -- dockerd \
		--storage-driver=vfs \
		--insecure-registry=0.0.0.0/0 \
                --pidfile=/var/run/docker-local.pid \
		"$@"
fi

if [ "$1" = 'dockerd' ]; then
	# if we're running Docker, let's pipe through dind
	# (and we'll run dind explicitly with "sh" since its shebang is /bin/bash)
	set -- sh "$(which dind)" "$@" "--insecure-registry=0.0.0.0/0"
fi


echo "$@"
exec "$@"

