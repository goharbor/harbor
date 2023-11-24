FROM photon:4.0

ARG GOVERSION

# install cgo-related dependencies && git
# official golang build from standard Debian, while photon os do not install git as default
RUN tdnf install build-essential git gpg -y

ENV PATH /usr/local/go/bin:$PATH

ENV GOLANG_VERSION ${GOVERSION}

# need to change sha256,  for every specific golang version from https://github.com/docker-library/golang
# or get checksum from https://go.dev/dl/ directly
RUN set eux; \
    url="https://dl.google.com/go/go${GOLANG_VERSION}.linux-amd64.tar.gz"; \
    sha256="73cac0215254d0c7d1241fa40837851f3b9a8a742d0b54714cbdfb3feaf8f0af"; \
    wget -O go.tgz.asc "$url.asc"; \
    wget -O go.tgz "$url"; \
    echo "$sha256 *go.tgz" | sha256sum -c -; \
    # https://github.com/golang/go/issues/14739#issuecomment-324767697
	GNUPGHOME="$(mktemp -d)"; export GNUPGHOME; \
    # https://www.google.com/linuxrepositories/
	gpg --batch --keyserver keyserver.ubuntu.com --recv-keys 'EB4C 1BFD 4F04 2F6D DDCC  EC91 7721 F63B D38B 4796'; \
    # let's also fetch the specific subkey of that key explicitly that we expect "go.tgz.asc" to be signed by, just to make sure we definitely have it
	gpg --batch --keyserver keyserver.ubuntu.com --recv-keys '2F52 8D36 D67B 69ED F998  D857 78BD 6547 3CB3 BD13'; \
	gpg --batch --verify go.tgz.asc go.tgz; \
	gpgconf --kill all; \
	rm -rf "$GNUPGHOME" go.tgz.asc; \
    tar -C /usr/local -xzf go.tgz; \
    rm go.tgz; \
    go version 

ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 1777 "$GOPATH"
WORKDIR $GOPATH