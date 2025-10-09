#!/bin/bash
set -x

set -e

sudo dnf install -y podman podman-docker buildah podman-compose
sudo sed -i 's/unqualified-search-registries = \["registry.access.redhat.com", "registry.redhat.io", "docker.io"\]/unqualified-search-registries = ["docker.io"]/g' /etc/containers/registries.conf
cd /data
sudo docker compose down -v
cd -
sudo rm -rf /data
sudo mkdir -p /data
sudo firewall-cmd --add-port=443/tcp
sudo firewall-cmd --add-port=443/tcp --permanent

sudo mkdir -p /data/cert
openssl req -newkey rsa:4096 -nodes -x509 -days 30 \
    -subj "/C=AU/ST=Victoria/L=Melbourne/O=deamen/CN=$(hostname)" \
    -addext "subjectAltName=IP:127.0.0.1" \
    -keyout /data/cert/harbor.key \
    -out /data/cert/harbor.crt

sudo cp harbor.yml.tmpl harbor.yml
sudo sed -i "s|^hostname:.*|hostname: $(hostname)|" harbor.yml
sudo sed -i "s|^harbor_admin_password:.*|harbor_admin_password: \"$(openssl rand -base64 30)\"|" harbor.yml
sudo sed -i "/^database:/ { n; n; s|^  password:.*|  password: \"$(openssl rand -base64 30)\"| }" harbor.yml
sudo sed -i "s|^  certificate:.*|  certificate: /data/cert/harbor.crt|" harbor.yml
sudo sed -i "s|^  private_key:.*|  private_key: /data/cert/harbor.key|" harbor.yml
sudo cp -r ./* /data/
# sudo chown -R 1000:1000 /data
sudo semanage fcontext -a -t container_file_t "/data(/.*)?"
sudo restorecon -R /data/
cd ../
make -f make/photon/Makefile _build_prepare -e BUILD_BASE=true BASEIMAGETAG=dev VERSIONTAG=dev
cd -
if ! grep -qF 'localhost/goharbor/prepare:dev' prepare; then
    sudo sed -i 's|goharbor/prepare:dev|localhost/goharbor/prepare:dev|g' prepare
fi
cd /data
sudo ./install.sh --with-trivy --with-podman
