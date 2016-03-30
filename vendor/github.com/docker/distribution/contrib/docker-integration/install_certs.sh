#!/bin/sh
set -e

hostname=$1
if [ "$hostname" = "" ]; then
	hostname="localhost"
fi

mkdir -p /etc/docker/certs.d/$hostname:5011
cp ./nginx/ssl/registry-ca+ca.pem /etc/docker/certs.d/$hostname:5011/ca.crt

mkdir -p /etc/docker/certs.d/$hostname:5440
cp ./nginx/ssl/registry-ca+ca.pem /etc/docker/certs.d/$hostname:5440/ca.crt

mkdir -p /etc/docker/certs.d/$hostname:5441
cp ./nginx/ssl/registry-ca+ca.pem /etc/docker/certs.d/$hostname:5441/ca.crt

mkdir -p /etc/docker/certs.d/$hostname:5442
cp ./nginx/ssl/registry-ca+ca.pem /etc/docker/certs.d/$hostname:5442/ca.crt
cp ./nginx/ssl/registry-ca+client-cert.pem /etc/docker/certs.d/$hostname:5442/client.cert
cp ./nginx/ssl/registry-ca+client-key.pem /etc/docker/certs.d/$hostname:5442/client.key

mkdir -p /etc/docker/certs.d/$hostname:5443
cp ./nginx/ssl/registry-ca+ca.pem /etc/docker/certs.d/$hostname:5443/ca.crt
cp ./nginx/ssl/registry-noca+client-cert.pem /etc/docker/certs.d/$hostname:5443/client.cert
cp ./nginx/ssl/registry-noca+client-key.pem /etc/docker/certs.d/$hostname:5443/client.key

mkdir -p /etc/docker/certs.d/$hostname:5444
cp ./nginx/ssl/registry-ca+ca.pem /etc/docker/certs.d/$hostname:5444/ca.crt
cp ./nginx/ssl/registry-ca+client-cert.pem /etc/docker/certs.d/$hostname:5444/client.cert
cp ./nginx/ssl/registry-ca+client-key.pem /etc/docker/certs.d/$hostname:5444/client.key

mkdir -p /etc/docker/certs.d/$hostname:5447
cp ./nginx/ssl/registry-ca+client-cert.pem /etc/docker/certs.d/$hostname:5447/client.cert
cp ./nginx/ssl/registry-ca+client-key.pem /etc/docker/certs.d/$hostname:5447/client.key

mkdir -p /etc/docker/certs.d/$hostname:5448
cp ./nginx/ssl/registry-ca+ca.pem /etc/docker/certs.d/$hostname:5448/ca.crt
