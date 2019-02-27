#!/bin/bash
IP=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`
PROTOCOL='https'

#echo $IP
sudo sed "s/reg.mydomain.com/$IP/" -i make/harbor.cfg
sudo sed "s/^ui_url_protocol = .*/ui_url_protocol = $PROTOCOL/g" -i make/harbor.cfg
