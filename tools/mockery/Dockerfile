ARG GOLANG
FROM ${GOLANG}

ARG MOCKERY_VERSION
ARG XTOOLS_VERSION=v0.51.0

# Use an x/tools version that can read Go 1.27.3 export data.
RUN mkdir /tmp/mockery-build && cd /tmp/mockery-build \
    && go mod init mockery-build \
    && go get github.com/vektra/mockery/v2@${MOCKERY_VERSION} golang.org/x/tools@${XTOOLS_VERSION} \
    && go install github.com/vektra/mockery/v2 \
    && rm -rf /tmp/mockery-build
