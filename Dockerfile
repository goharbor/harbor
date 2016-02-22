FROM golang:1.5.1

MAINTAINER jiangd@vmware.com

RUN apt-get update \
    && apt-get install -y libldap2-dev \
    && rm -r /var/lib/apt/lists/*

COPY . /go/src/github.com/vmware/harbor
#golang.org is blocked in China
COPY ./vendor/golang.org /go/src/golang.org 
WORKDIR /go/src/github.com/vmware/harbor

ENV GO15VENDOREXPERIMENT 1
RUN go get -d github.com/docker/distribution \
    && go get -d github.com/docker/libtrust \
    && go get -d github.com/go-sql-driver/mysql
RUN go install -v -a 

ENV MYSQL_USR root \
    MYSQL_PWD root \
    MYSQL_PORT_3306_TCP_ADDR localhost \
    MYSQL_PORT_3306_TCP_PORT 3306 \
    REGISTRY_URL localhost:5000

COPY conf /go/bin/conf
COPY views /go/bin/views
COPY static /go/bin/static

RUN chmod u+x /go/bin/harbor \
    && sed -i 's/TLS_CACERT/#TLS_CAERT/g' /etc/ldap/ldap.conf \
    && sed -i '$a\TLS_REQCERT allow' /etc/ldap/ldap.conf

WORKDIR /go/bin/
ENTRYPOINT ["/go/bin/harbor"]

EXPOSE 80

