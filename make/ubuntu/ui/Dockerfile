FROM golang:1.6.2

MAINTAINER jiangd@vmware.com

RUN apt-get update && apt-get install -y libldap2-dev \
    && rm -r /var/lib/apt/lists/*

ENV MYSQL_USR root \
    MYSQL_PWD root \
    REGISTRY_URL localhost:5000

RUN mkdir /harbor/
COPY ./make/dev/ui/harbor_ui /harbor/

COPY ./src/ui/views /harbor/views
COPY ./src/ui/static /harbor/static
COPY ./src/favicon.ico /harbor/favicon.ico
COPY ./make/jsminify.sh /tmp/jsminify.sh

RUN chmod u+x /harbor/harbor_ui \
    && sed -i 's/TLS_CACERT/#TLS_CAERT/g' /etc/ldap/ldap.conf \
    && sed -i '$a\TLS_REQCERT allow' /etc/ldap/ldap.conf \
    && timestamp=`date '+%s'` \
    && /tmp/jsminify.sh /harbor/views/sections/script-include.htm /harbor/static/resources/js/harbor.app.min.$timestamp.js /harbor/ \
    && sed -i "s/harbor\.app\.min\.js/harbor\.app\.min\.$timestamp\.js/g" /harbor/views/sections/script-min-include.htm

WORKDIR /harbor/
ENTRYPOINT ["/harbor/harbor_ui"]

EXPOSE 80

