IP=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`
#echo $IP
docker pull hello-world
docker pull docker
#docker login -u admin -p Harbor12345 $IP

docker tag hello-world $IP/library/hello-world
docker push $IP/library/hello-world

docker tag docker $IP/library/docker
docker push $IP/library/docker
