#!/bin/bash
set -e

workdir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $workdir/../harbor

echo "[Step 1]: preparing environment ..."
./prepare

echo "[Step 2]: starting Harbor ..."
docker-compose -f docker-compose*.yml up -d

protocol=http
hostname=reg.mydomain.com

if [[ $(cat ./harbor.cfg) =~ ui_url_protocol[[:blank:]]*=[[:blank:]]*(https?) ]]
then
protocol=${BASH_REMATCH[1]}
fi

if [[ $(grep 'hostname[[:blank:]]*=' ./harbor.cfg) =~ hostname[[:blank:]]*=[[:blank:]]*(.*) ]]
then
hostname=${BASH_REMATCH[1]}
fi

echo $"
----Harbor has been installed and started successfully.----

Now you should be able to visit the admin portal at ${protocol}://${hostname}. 
For more details, please visit https://github.com/vmware/harbor .
"
