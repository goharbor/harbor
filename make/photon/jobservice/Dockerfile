ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-jobservice-base:${harbor_base_image_version}

COPY --link --chown=10000:10000 --chmod=755 ./make/photon/common/install_cert.sh /harbor/
COPY --link --chown=10000:10000 --chmod=744 ./make/photon/jobservice/entrypoint.sh /harbor/
COPY --link --chown=10000:10000 --chmod=755 ./make/photon/jobservice/harbor_jobservice /harbor/


WORKDIR /harbor/

USER harbor

VOLUME ["/var/log/jobs/"]

HEALTHCHECK CMD curl --fail -s http://localhost:8080/api/v1/stats || curl -sk --fail --key /etc/harbor/ssl/job_service.key --cert /etc/harbor/ssl/job_service.crt https://localhost:8443/api/v1/stats || exit 1

ENTRYPOINT ["/harbor/entrypoint.sh"]
