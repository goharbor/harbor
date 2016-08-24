#/bin/bash

echo " "
echo "Please enter the registry service you want to pull the pre-built images from."
echo "Enter 1 for Docker Hub."
echo "Enter 2 for Daocloud.io (recommended for Chinese users)."
echo "or enter other registry URL such as https://my_registry/harbor/ ."
read -p "The default is 1 (Docker Hub): " choice

cd ../../Deploy
template_file="docker-compose.yml.template"
yml_file='docker-compose.yml'
if test -e $template_file 
then
  cp $template_file $yml_file
else
  cp $yml_file $template_file
fi
platform=''
choice=${choice:-1} 
if [ $choice == '1' ]
then
  platform='prjharbor/'
elif [ $choice == '2' ]
then
  platform='daocloud.io/harbor/'
else
  platform=$choice
fi
version='0.3.0'
log='deploy_log:'
db='deploy_mysql:'
job_service='deploy_jobservice:'
ui='deploy_ui:'
sed -i -- '/build: .\/log\//c\    image: '$platform$log$version'' $yml_file
sed -i -- '/build: .\/db\//c\    image: '$platform$db$version'' $yml_file
sed -i -- '/ui:/{n;N;N;d}' $yml_file && sed -i -- '/ui:/a\\    image: '$platform$ui$version'' $yml_file
sed -i -- '/jobservice:/{n;N;N;d}' $yml_file && sed -i -- '/jobservice:/a\\    image: '$platform$job_service$version'' $yml_file
echo "Succeeded! "
echo "Please follow the normal installation process to install Harbor."

