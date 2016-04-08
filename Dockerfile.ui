FROM golang:1.5.1

MAINTAINER jiangd@vmware.com

RUN apt-get update \
    && apt-get install -y libldap2-dev \
    && rm -r /var/lib/apt/lists/*

COPY . /go/src/github.com/vmware/harbor
#golang.org is blocked in China
COPY ./vendor/golang.org /go/src/golang.org 
WORKDIR /go/src/github.com/vmware/harbor/ui

ENV GO15VENDOREXPERIMENT 1
RUN go get -d github.com/docker/distribution \
    && go get -d github.com/docker/libtrust \
    && go get -d github.com/go-sql-driver/mysql \
    && go build -v -a -o /go/bin/harbor_ui

ENV MYSQL_USR root \
    MYSQL_PWD root \
    REGISTRY_URL localhost:5000

COPY views /go/bin/views
COPY static /go/bin/static
COPY favicon.ico /go/bin/favicon.ico

RUN chmod u+x /go/bin/harbor_ui \
    && sed -i 's/TLS_CACERT/#TLS_CAERT/g' /etc/ldap/ldap.conf \
    && sed -i '$a\TLS_REQCERT allow' /etc/ldap/ldap.conf

WORKDIR /go/bin/
ENTRYPOINT ["/go/bin/harbor_ui"]

EXPOSE 80

