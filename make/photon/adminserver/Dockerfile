FROM vmware/photon:1.0

RUN tdnf distro-sync -y \
    && tdnf erase vim -y \
    && tdnf clean all \
    && mkdir /harbor/
COPY ./make/dev/adminserver/harbor_adminserver /harbor/

RUN chmod u+x /harbor/harbor_adminserver
WORKDIR /harbor/
ENTRYPOINT ["/harbor/harbor_adminserver"]
