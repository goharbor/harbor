FROM photon:2.0

Label maintainer="wangyan@vmware.com"

RUN tdnf install sudo -y >> /dev/null \
    && tdnf clean all \
    && groupadd -r -g 10000 harbor && useradd --no-log-init -m -g 10000 -u 10000 harbor \
    && mkdir -p /etc/registry
