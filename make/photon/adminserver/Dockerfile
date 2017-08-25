FROM library/photon:1.0

RUN mkdir /harbor/
COPY ./make/dev/adminserver/harbor_adminserver /harbor/

RUN chmod u+x /harbor/harbor_adminserver
WORKDIR /harbor/

HEALTHCHECK CMD curl http://localhost:80/v2 || exit 1
ENTRYPOINT ["/harbor/harbor_adminserver"]
