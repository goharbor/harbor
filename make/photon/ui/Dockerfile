FROM vmware/photon:1.0

RUN tdnf distro-sync -y \
    && tdnf erase vim -y \
    && tdnf install sudo -y \
    && tdnf clean all \
    && groupadd -r -g 10000 harbor && useradd --no-log-init -r -g 10000 -u 10000 harbor \
    && mkdir /harbor/

HEALTHCHECK CMD curl -s -o /dev/null -w "%{http_code}" 127.0.0.1:8080/api/systeminfo|grep 200
COPY ./make/dev/ui/harbor_ui ./src/favicon.ico ./make/photon/ui/start.sh ./VERSION /harbor/
COPY ./src/ui/views /harbor/views
COPY ./src/ui/static /harbor/static

RUN chmod u+x /harbor/start.sh /harbor/harbor_ui
WORKDIR /harbor/
	
ENTRYPOINT ["/harbor/start.sh"]
