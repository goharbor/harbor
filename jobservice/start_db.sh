#export MYQL_ROOT_PASSWORD=root123
docker run --name harbor_mysql -d -e MYSQL_ROOT_PASSWORD=root123 -p 3306:3306 -v /devdata/database:/var/lib/mysql harbor/mysql:dev 

echo "sleep 10 seconds..."
sleep 10

mysql -h 127.0.0.1 -uroot -proot123 < ./populate.sql
