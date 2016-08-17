FROM golang:1.6.2

MAINTAINER jiangd@vmware.com

RUN apt-get update \
    && apt-get install -y libldap2-dev \
    && rm -r /var/lib/apt/lists/*

COPY . /go/src/github.com/vmware/harbor

WORKDIR /go/src/github.com/vmware/harbor/jobservice

RUN go build -v -a -o /go/bin/harbor_jobservice \
    && chmod u+x /go/bin/harbor_jobservice 
WORKDIR /go/bin/
ENTRYPOINT ["/go/bin/harbor_jobservice"]
