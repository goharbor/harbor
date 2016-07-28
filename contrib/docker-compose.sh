#/bin/eash
read -p "Please input the platform name you want to pull images, for daocloud.io,enter 1; for docker hub, enter 2, otherwise enter the name of the platform, the default is 1:" choice 
cp docker-compose.yml.daocloud ../Deploy
cd ../Deploy
choice=${choice:-1}
if [ $choice == '2' ]
then
  sed -i -- 's/daocloud.io/docker.io/g' docker-compose.yml.daocloud
elif [ $choice != '1' ]
then
  sed -i -- 's/daocloud.io/'$choice'/g' docker-compose.yml.daocloud
fi
mv docker-compose.yml docker-compose.yml.bak
mv docker-compose.yml.daocloud docker-compose.yml
echo "succeeded! "
