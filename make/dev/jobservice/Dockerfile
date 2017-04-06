FROM golang:1.7.3

MAINTAINER jiangd@vmware.com

COPY . /go/src/github.com/vmware/harbor

WORKDIR /go/src/github.com/vmware/harbor/src/jobservice

RUN go build -v -a -o /go/bin/harbor_jobservice \
    && chmod u+x /go/bin/harbor_jobservice 
WORKDIR /go/bin/
ENTRYPOINT ["/go/bin/harbor_jobservice"]
