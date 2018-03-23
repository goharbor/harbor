#!/bin/bash

outputdir=/tmp
outputfolder=harbor_logs
dir=$outputdir/$outputfolder
mkdir -p $dir

echo "Version" >> $dir/docker
docker version >> $dir/docker
printf "\n\nInfo\n" >> $dir/docker
docker info >> $dir/docker
printf "\n\nImages\n" >> $dir/docker
docker images >> $dir/docker
printf "\n\nRunning containers\n" >> $dir/docker
docker ps >> $dir/docker

docker-compose version >> $dir/docker-compose

base_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cp -r $base_dir/../script $dir/
cp $base_dir/../harbor/harbor.cfg $dir/
cp -r /var/log/harbor $dir/

properties=( 
	email_server 
	email_server_port 
	email_username 
	email_password 
	email_from 
	harbor_admin_password 
	ldap_url 
	ldap_searchdn 
	ldap_search_pwd 
	ldap_basedn 
	db_password 
	)
	
for property in "${properties[@]}"
do
	sed -i -r "s%#?$property\s*=\s*.*%$property = %" $dir/harbor.cfg
done

tar --remove-files -zcf $outputfolder.tar.gz -C $outputdir $outputfolder

echo "$outputfolder.tar.gz is generated in current directory."