FROM golang:1.9.2

MAINTAINER wangyan@vmware.com

COPY . /go/src/github.com/vmware/harbor

WORKDIR /go/src/github.com/vmware/harbor/src/registryctl

RUN go build -a -o /go/bin/harbor_registryctl \
    && chmod u+x /go/bin/harbor_registryctl 
WORKDIR /go/bin/
ENTRYPOINT ["/go/bin/harbor_registryctl"]
