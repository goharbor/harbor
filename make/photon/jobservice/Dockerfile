ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-jobservice-base:${harbor_base_image_version}

COPY ./make/photon/common/install_cert.sh /harbor/
COPY ./make/photon/jobservice/entrypoint.sh /harbor/
COPY ./make/photon/jobservice/harbor_jobservice /harbor/


RUN chown -R harbor:harbor /etc/pki/tls/certs \
    && chown harbor:harbor /harbor/entrypoint.sh && chmod u+x /harbor/entrypoint.sh \
    && chown harbor:harbor /harbor/install_cert.sh && chmod u+x /harbor/install_cert.sh \
    && chown harbor:harbor /harbor/harbor_jobservice && chmod u+x /harbor/harbor_jobservice

WORKDIR /harbor/

USER harbor

VOLUME ["/var/log/jobs/"]

HEALTHCHECK CMD curl --fail -s http://127.0.0.1:8080/api/v1/stats || curl -k --fail -s https://127.0.0.1:8443/api/v1/stats || exit 1

ENTRYPOINT ["/harbor/entrypoint.sh"]
