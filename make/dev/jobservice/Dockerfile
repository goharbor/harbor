FROM golang:1.9.2

MAINTAINER jiangd@vmware.com

COPY . /go/src/github.com/vmware/harbor

WORKDIR /go/src/github.com/vmware/harbor/src/jobservice

RUN go build -a -o /go/bin/harbor_jobservice \
    && chmod u+x /go/bin/harbor_jobservice 
WORKDIR /go/bin/
ENTRYPOINT ["/go/bin/harbor_jobservice"]
