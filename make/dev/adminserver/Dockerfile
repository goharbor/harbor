FROM golang:1.9.2

MAINTAINER yinw@vmware.com

COPY . /go/src/github.com/goharbor/harbor

WORKDIR /go/src/github.com/goharbor/harbor/src/adminserver

RUN go build -a -o /go/bin/harbor_adminserver \
    && chmod u+x /go/bin/harbor_adminserver 

COPY ./make/migrations /go/bin/migrations
WORKDIR /go/bin/
ENTRYPOINT ["/go/bin/harbor_adminserver"]
