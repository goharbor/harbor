#export MYQL_ROOT_PASSWORD=root123
docker run --name harbor_mysql -d -e MYSQL_ROOT_PASSWORD=root123 -p 3306:3306 -v /devdata/database:/var/lib/mysql harbor/mysql:dev 
