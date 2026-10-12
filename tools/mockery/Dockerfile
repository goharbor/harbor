ARG GOLANG
FROM ${GOLANG}

ARG MOCKERY_VERSION

# https://github.com/docker-library/golang/issues/225
ENV XDG_CACHE_HOME=/tmp

RUN git clone --depth=1 --branch ${MOCKERY_VERSION} https://github.com/vektra/mockery /tmp/mockery && \
    cd /tmp/mockery && \
    go get golang.org/x/tools@latest && \
    go build -ldflags "-X github.com/vektra/mockery/v2/pkg/logging.SemVer=${MOCKERY_VERSION}" -o /go/bin/mockery && \
    rm -rf /tmp/mockery
