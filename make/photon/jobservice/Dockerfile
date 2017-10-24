FROM vmware/photon:1.0-20170928

RUN mkdir /harbor/ \
    && tdnf distro-sync -y \
    && tdnf clean all
COPY ./make/dev/jobservice/harbor_jobservice /harbor/

RUN chmod u+x /harbor/harbor_jobservice
WORKDIR /harbor/
ENTRYPOINT ["/harbor/harbor_jobservice"]
