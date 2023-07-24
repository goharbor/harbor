FROM photon:4.0

ARG GOVERSION

ENV PATH /usr/local/go/bin:$PATH

ENV GOLANG_VERSION ${GOVERSION}

RUN set eux; \
    url="https://dl.google.com/go/go${GOLANG_VERSION}.linux-amd64.tar.gz"; \
    wget -O go.tgz "$url"; \
    tar -C /usr/local -xzf go.tgz; \
    rm go.tgz; \
    go version 

ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 1777 "$GOPATH"
WORKDIR $GOPATH