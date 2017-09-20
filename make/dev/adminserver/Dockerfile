FROM golang:1.7.3

MAINTAINER yinw@vmware.com

COPY . /go/src/github.com/vmware/harbor

WORKDIR /go/src/github.com/vmware/harbor/src/adminserver

RUN go build -v -a -o /go/bin/harbor_adminserver \
    && chmod u+x /go/bin/harbor_adminserver

RUN useradd -m -d /home/adminserver -s /bin/bash adminserver \
    && chown adminserver.adminserver /go/bin/harbor_adminserver
    && chmod u+x /go/bin/harbor_adminserver

USER adminserver

WORKDIR /go/bin/
ENTRYPOINT ["/go/bin/harbor_adminserver"]
