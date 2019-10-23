#!/bin/bash
IP=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`

#echo $IP
sudo sed "s/reg.mydomain.com/$IP/" -i make/harbor.yml

sed "s|/your/certificate/path|/data/cert/server.crt|g" -i make/harbor.yml
sed "s|/your/private/key/path|/data/cert/server.key|g" -i make/harbor.yml
