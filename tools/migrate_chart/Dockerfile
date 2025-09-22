########## Stage 1: fetch correct Helm by arch ##########
FROM alpine:3.20 AS helmfetch
ARG TARGETARCH
ARG HELM_VERSION=v3.9.1
RUN apk add --no-cache curl tar ca-certificates \
 && curl -fsSL "https://get.helm.sh/helm-${HELM_VERSION}-linux-${TARGETARCH}.tar.gz" -o /tmp/helm.tgz \
 && tar -xzf /tmp/helm.tgz -C /tmp \
 && install -m 0755 "/tmp/linux-${TARGETARCH}/helm" /helm

########## Stage 2: base unchanged ##########
FROM python:3.8.5-slim

ENV HELM_EXPERIMENTAL_OCI=1 \
    REQUESTS_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt

COPY --from=helmfetch /helm /usr/local/bin/helm

COPY ./migrate_chart.py ./migrate_chart.sh /
RUN pip install --no-cache-dir click==7.1.2 requests==2.24.0 pyyaml \
 && chmod +x /migrate_chart.sh ./migrate_chart.py

ENTRYPOINT ["/migrate_chart.py"]