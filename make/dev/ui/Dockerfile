FROM golang:1.7.3

MAINTAINER jiangd@vmware.com

COPY src/. /go/src/github.com/vmware/harbor/src
WORKDIR /go/src/github.com/vmware/harbor/src/ui

RUN go build -v -a -o /go/bin/harbor_ui

COPY src/ui/views /go/bin/views
COPY src/ui/static /go/bin/static
COPY src/favicon.ico /go/bin/favicon.ico

RUN mkdir /go/bin/harbor/
COPY VERSION /go/bin/harbor/VERSION

RUN chmod u+x /go/bin/harbor_ui

WORKDIR /go/bin/
ENTRYPOINT ["/go/bin/harbor_ui"]

EXPOSE 80

