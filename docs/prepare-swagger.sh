#!/bin/bash
SCHEME=http
SERVER_IP=reg.mydomain.com
set -e
echo "Doing some clean up..."
rm -f *.tar.gz
echo "Downloading Swagger UI release package..."
wget https://github.com/swagger-api/swagger-ui/archive/v2.1.4.tar.gz -O swagger.tar.gz
echo "Untarring Swagger UI package to the static file path..."
tar -C ../static/vendors -zxf swagger.tar.gz swagger-ui-2.1.4/dist
echo "Executing some processes..."
sed -i 's/http:\/\/petstore\.swagger\.io\/v2\/swagger\.json/'$SCHEME':\/\/'$SERVER_IP'\/static\/resources\/yaml\/swagger\.yaml/g' \
 ../static/vendors/swagger-ui-2.1.4/dist/index.html
mkdir -p ../static/resources/yaml
cp swagger.yaml ../static/resources/yaml
sed -i 's/host: localhost/host: '$SERVER_IP'/g' ../static/resources/yaml/swagger.yaml
sed -i 's/  \- http$/  \- '$SCHEME'/g' ../static/resources/yaml/swagger.yaml
echo "Finish preparation for the Swagger UI."
