ARG harbor_base_image_version
ARG harbor_base_namespace
ARG GOBUILDIMAGE
FROM ${GOBUILDIMAGE} AS builder
WORKDIR /harbor
ARG TRIVYADAPTERVERSION
ARG TRIVYVERSION
ENV CGO_ENABLED=0
RUN mkdir -p /go/src/github.com/aquasecurity
WORKDIR  /go/src/github.com/aquasecurity

RUN git clone -b $TRIVYADAPTERVERSION https://github.com/aquasecurity/harbor-scanner-trivy.git \
    && cd  harbor-scanner-trivy/ \
    && make build

RUN git clone -b $TRIVYVERSION https://github.com/aquasecurity/trivy.git \
    && cd  trivy/ \
    && go build -o trivy cmd/trivy/main.go


FROM ${harbor_base_namespace}/harbor-trivy-adapter-base:${harbor_base_image_version}

ARG trivy_version

COPY ./make/photon/common/install_cert.sh /home/scanner
COPY ./make/photon/trivy-adapter/entrypoint.sh /home/scanner
COPY --from=builder /go/src/github.com/aquasecurity/trivy/trivy /usr/local/bin/trivy
COPY --from=builder /go/src/github.com/aquasecurity/harbor-scanner-trivy/scanner-trivy /home/scanner/bin/scanner-trivy


RUN chown -R scanner:scanner /etc/pki/tls/certs \
    && chown scanner:scanner /home/scanner/entrypoint.sh && chmod u+x /home/scanner/entrypoint.sh \
    && chown scanner:scanner /usr/local/bin/trivy && chmod u+x /usr/local/bin/trivy \
    && chown scanner:scanner /home/scanner/bin/scanner-trivy && chmod u+x /home/scanner/bin/scanner-trivy \
    && chown scanner:scanner /home/scanner/install_cert.sh && chmod u+x /home/scanner/install_cert.sh

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl --fail -s http://localhost:8080/probe/healthy || curl -k --fail -s https://localhost:8443/probe/healthy || exit 1

ENV TRIVY_VERSION=${trivy_version}

USER scanner

ENTRYPOINT ["/home/scanner/entrypoint.sh"]
