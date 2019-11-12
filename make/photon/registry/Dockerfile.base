FROM photon:2.0

RUN tdnf install sudo -y >> /dev/null\
    && tdnf clean all \
    && mkdir -p /etc/registry \
    && groupadd -r -g 10000 harbor && useradd --no-log-init -m -g 10000 -u 10000 harbor