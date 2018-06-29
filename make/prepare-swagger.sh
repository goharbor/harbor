#!/bin/bash
# Note: Download prepare-swagger.sh and swagger.yaml to the directory which contains the docker-compose.yml 
SCHEME=http
SERVER_IP=reg.mydomain.com

if [ $# = 1 ]  && [ $1 = "-f" ]; then 
    SCHEME=$(grep "ui_url_protocol =" ./harbor.cfg  |  awk '{ print $3 }')
    SERVER_IP=$(grep "hostname =" ./harbor.cfg  |  awk '{ print $3 }')
fi
set -e
echo "Doing some clean up..."
rm -f *.tar.gz
echo "Downloading Swagger UI release package..."
wget https://github.com/swagger-api/swagger-ui/archive/v2.1.4.tar.gz -O swagger.tar.gz
echo "Untarring Swagger UI package to the static file path..."
mkdir -p ../src/ui/static/vendors
tar -C ../src/ui/static/vendors -zxf swagger.tar.gz swagger-ui-2.1.4/dist
echo "Executing some processes..."
sed -i.bak 's/http:\/\/petstore\.swagger\.io\/v2\/swagger\.json/'$SCHEME':\/\/'$SERVER_IP'\/static\/resources\/yaml\/swagger\.yaml/g' \
../src/ui/static/vendors/swagger-ui-2.1.4/dist/index.html
sed -i.bak '/jsonEditor: false,/a\        validatorUrl: null,' ../src/ui/static/vendors/swagger-ui-2.1.4/dist/index.html
mkdir -p ../src/ui/static/resources/yaml
cp swagger.yaml ../src/ui/static/resources/yaml
sed -i.bak 's/host: localhost/host: '$SERVER_IP'/g' ../src/ui/static/resources/yaml/swagger.yaml
sed -i.bak 's/  \- http$/  \- '$SCHEME'/g' ../src/ui/static/resources/yaml/swagger.yaml


COMPOSE_FILE=./docker-compose.yml

if [ $# = 1 ]  && [ $1 = "-f" ];  then 
    	if grep -q "swagger" -F $COMPOSE_FILE ; then
		echo "Skip to enable swagger in docker-compose.yml"
	else
		sed -i.bak "/\/etc\/ui\/token\/\:z/a\      \- ../src/ui/static/vendors/swagger-ui-2.1.4/dist:/harbor/static/vendors/swagger\n\      \- ../src/ui/static/resources/yaml/swagger.yaml:/harbor/static/resources/yaml/swagger.yaml" ./docker-compose.yml

	fi
fi
echo "Finish preparation for the Swagger UI."
if [ $# = 1 ]  && [ $1 = "-f" ]; then 
    echo "Restarting harbor"
    docker-compose down -v 
    docker-compose up -d
    echo "Swagger UI is enabled, please visit $SCHEME://$SERVER_IP/static/vendors/swagger/index.html"
fi
