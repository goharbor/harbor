FROM cicd.harbor.bitsf.xin/harbor-depend/registryctl:base

COPY ./make/photon/common/install_cert.sh /harbor
COPY ./make/photon/registry/binary/registry /usr/bin
COPY ./make/photon/registryctl/start.sh /harbor/
COPY ./make/photon/registryctl/harbor_registryctl /harbor/

RUN chmod u+x /harbor/harbor_registryctl \
    && chmod u+x /usr/bin/registry \
    && chmod u+x /harbor/start.sh

HEALTHCHECK CMD curl --fail -s http://127.0.0.1:8080/api/health || exit 1 

VOLUME ["/var/lib/registry"]
WORKDIR /harbor/

ENTRYPOINT ["/harbor/start.sh"]
