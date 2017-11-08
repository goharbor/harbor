FROM vmware/photon:1.0

RUN tdnf distro-sync -y \
    && tdnf erase vim -y \
    && tdnf clean all \
    && mkdir /harbor/
COPY ./make/dev/ui/harbor_ui /harbor/

COPY ./src/ui/views /harbor/views
COPY ./src/ui/static /harbor/static
COPY ./src/favicon.ico /harbor/favicon.ico
COPY ./VERSION /harbor/VERSION

RUN chmod u+x /harbor/harbor_ui
WORKDIR /harbor/

HEALTHCHECK CMD curl localhost:80 || exit 1
ENTRYPOINT ["/harbor/harbor_ui"]
