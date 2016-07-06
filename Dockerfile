FROM golang:1.6.2


RUN apt-get update \
    && apt-get install -y libldap2-dev \
    && rm -r /var/lib/apt/lists/*
WORKDIR /go/src/github.com/vmware/harbor/ui
RUN go get -d github.com/docker/distribution \
    && go get -d github.com/docker/libtrust \
    && go get -d github.com/go-sql-driver/mysql 
