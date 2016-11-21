FROM golang:1.6.2

MAINTAINER jiangd@vmware.com

RUN apt-get update \
    && apt-get install -y libldap2-dev \
    && rm -r /var/lib/apt/lists/*

COPY src/. /go/src/github.com/vmware/harbor/src
WORKDIR /go/src/github.com/vmware/harbor/src/ui

RUN go build -v -a -o /go/bin/harbor_ui

ENV MYSQL_USR root \
    MYSQL_PWD root \
    REGISTRY_URL localhost:5000

COPY src/ui/views /go/bin/views
COPY src/ui/static /go/bin/static
COPY src/favicon.ico /go/bin/favicon.ico
COPY make/jsminify.sh /tmp/jsminify.sh

RUN chmod u+x /go/bin/harbor_ui \
    && sed -i 's/TLS_CACERT/#TLS_CAERT/g' /etc/ldap/ldap.conf \
    && sed -i '$a\TLS_REQCERT allow' /etc/ldap/ldap.conf \
    && timestamp=`date '+%s'` \
    && /tmp/jsminify.sh /go/bin/views/sections/script-include.htm /go/bin/static/resources/js/harbor.app.min.$timestamp.js /go/bin/ \
    && sed -i "s/harbor\.app\.min\.js/harbor\.app\.min\.$timestamp\.js/g" /go/bin/views/sections/script-min-include.htm

WORKDIR /go/bin/
ENTRYPOINT ["/go/bin/harbor_ui"]

EXPOSE 80

