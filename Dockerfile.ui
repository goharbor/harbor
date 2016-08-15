FROM golang:1.6.2

MAINTAINER jiangd@vmware.com

RUN apt-get update \
    && apt-get install -y libldap2-dev \
    && rm -r /var/lib/apt/lists/*

COPY . /go/src/github.com/vmware/harbor
WORKDIR /go/src/github.com/vmware/harbor/ui

RUN go build -v -a -o /go/bin/harbor_ui

ENV MYSQL_USR root \
    MYSQL_PWD root \
    REGISTRY_URL localhost:5000

COPY views /go/bin/views
COPY static /go/bin/static
COPY favicon.ico /go/bin/favicon.ico
COPY Deploy/jsminify.sh /tmp/jsminify.sh

RUN chmod u+x /go/bin/harbor_ui \
    && sed -i 's/TLS_CACERT/#TLS_CAERT/g' /etc/ldap/ldap.conf \
    && sed -i '$a\TLS_REQCERT allow' /etc/ldap/ldap.conf \
    && /tmp/jsminify.sh /go/bin/views/sections/script-include.htm /go/bin/static/resources/js/harbor.app.min.js

WORKDIR /go/bin/
ENTRYPOINT ["/go/bin/harbor_ui"]

EXPOSE 80

