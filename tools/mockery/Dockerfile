ARG GOLANG
FROM ${GOLANG}

ARG MOCKERY_VERSION

# https://github.com/docker-library/golang/issues/225
ENV XDG_CACHE_HOME /tmp
ENV GO111MODULE auto

RUN ARCH= && dpkgArch="$(uname -m)" \
  && case "${dpkgArch}" in \
    x86_64) ARCH='x64'; mkdir -p /tmp/mockery-${MOCKERY_VERSION} && \
    curl -fsSL https://github.com/vektra/mockery/releases/download/${MOCKERY_VERSION}/mockery_${MOCKERY_VERSION#v}_Linux_x86_64.tar.gz | tar -xz -C /tmp/mockery-${MOCKERY_VERSION} && \
    mv /tmp/mockery-${MOCKERY_VERSION}/mockery /usr/local/bin && \
    chmod +x /usr/local/bin/mockery && \
    rm -rf /tmp/mockery-${MOCKERY_VERSION}  ;; \
    aarch64) ARCH='arm64'; mkdir -p /tmp/mockery-${MOCKERY_VERSION} && \
    curl -fsSL https://github.com/vektra/mockery/releases/download/${MOCKERY_VERSION}/mockery_${MOCKERY_VERSION#v}_Linux_arm64.tar.gz | tar -xz -C /tmp/mockery-${MOCKERY_VERSION} && \
    mv /tmp/mockery-${MOCKERY_VERSION}/mockery /usr/local/bin && \
    chmod +x /usr/local/bin/mockery && \
    rm -rf /tmp/mockery-${MOCKERY_VERSION}  ;; \
    ppc64el) ARCH='ppc64le'; ;; \
    s390x) ARCH='s390x'; ;; \
    *) echo "unsupported architecture"; exit 1 ;; \
    esac
