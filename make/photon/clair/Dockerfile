FROM library/photon:1.0

RUN tdnf install -y git bzr rpm xz \
	&& mkdir /clair2.0.1/
RUN curl -L https://github.com/Yelp/dumb-init/releases/download/v1.2.0/dumb-init_1.2.0_amd64 > /usr/local/bin/dumb-init \
	&& chmod +x /usr/local/bin/dumb-init


COPY clair /clair2.0.1/

VOLUME /config
EXPOSE 6060 6061

RUN chmod u+x /clair2.0.1/clair

ENTRYPOINT ["/usr/local/bin/dumb-init", "--", "/clair2.0.1/clair"]
