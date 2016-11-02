#!/usr/bin/env bash

tdnf install -y docker
systemctl enable docker.service

mkdir -p /var/log/harbor

echo "Downloading harbor..."
wget -O /ova.tar.gz http://10.117.5.62/ISV/appliancePackages/ova.tar.gz

echo "Downloading notice file..."
wget -O /NOTICE_Harbor_0.4.1_Beta.txt http://10.117.5.62/ISV/appliancePackages/NOTICE_Harbor_0.4.1_Beta.txt

echo "Downloading license file..."
wget -O /LICENSE_Harbor_0.4.1_Beta_100216.txt http://10.117.5.62/ISV/appliancePackages/LICENSE_Harbor_0.4.1_Beta_100216.txt