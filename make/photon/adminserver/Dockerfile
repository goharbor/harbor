FROM vmware/photon:1.0

RUN tdnf erase vim -y \
    && tdnf distro-sync -y \
    && tdnf install -y sudo >> /dev/null \
    && tdnf clean all \
    && groupadd -r -g 10000 harbor && useradd --no-log-init -r -g 10000 -u 10000 harbor \
    && mkdir /harbor/
COPY ./make/dev/adminserver/harbor_adminserver ./make/photon/adminserver/start.sh /harbor/
HEALTHCHECK CMD curl --fail -s http://127.0.0.1:8080/api/ping || exit 1 

RUN chmod u+x /harbor/harbor_adminserver /harbor/start.sh
WORKDIR /harbor/
ENTRYPOINT ["/harbor/start.sh"]
