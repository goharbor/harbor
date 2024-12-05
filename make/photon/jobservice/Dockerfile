ARG GOBUILDIMAGE
ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${GOBUILDIMAGE} AS builder
WORKDIR /harbor
COPY ./ /harbor
ENV CGO_ENABLED=0
RUN apt update \
    && apt install -y make \
    && cd /harbor \
    && make jobservice



FROM ${harbor_base_namespace}/harbor-jobservice-base:${harbor_base_image_version}
COPY --from=builder /harbor/make/photon/common/install_cert.sh /harbor/
COPY --from=builder /harbor/make/photon/jobservice/entrypoint.sh /harbor/
COPY --from=builder /harbor/make/photon/jobservice/harbor_jobservice /harbor/


RUN chown -R harbor:harbor /harbor/ \
    && chmod u+x /harbor/entrypoint.sh \
    && chmod u+x /harbor/install_cert.sh \
    && chmod u+x /harbor/harbor_jobservice

WORKDIR /harbor/

USER harbor

VOLUME ["/var/log/jobs/"]

HEALTHCHECK CMD curl --fail -s http://localhost:8080/api/v1/stats || curl -sk --fail --key /etc/harbor/ssl/job_service.key --cert /etc/harbor/ssl/job_service.crt https://localhost:8443/api/v1/stats || exit 1

ENTRYPOINT ["/harbor/entrypoint.sh"]
