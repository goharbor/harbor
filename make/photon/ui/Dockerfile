FROM library/photon:1.0

RUN mkdir /harbor/
RUN tdnf install -y sed apr-util-ldap

COPY ./make/dev/ui/harbor_ui /harbor/

COPY ./src/ui/views /harbor/views
COPY ./src/ui/static /harbor/static
COPY ./src/favicon.ico /harbor/favicon.ico
COPY ./make/jsminify.sh /tmp/jsminify.sh

RUN chmod u+x /harbor/harbor_ui \
    && timestamp=`date '+%s'` \
    && /tmp/jsminify.sh /harbor/views/sections/script-include.htm /harbor/static/resources/js/harbor.app.min.$timestamp.js /harbor/ \
    && sed -i "s/harbor\.app\.min\.js/harbor\.app\.min\.$timestamp\.js/g" /harbor/views/sections/script-min-include.htm \
    && echo "TLS_REQCERT allow" >> /etc/openldap/ldap.conf
	
WORKDIR /harbor/
ENTRYPOINT ["/harbor/harbor_ui"]
