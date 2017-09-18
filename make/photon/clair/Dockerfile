FROM library/photon:1.0

RUN tdnf install -y git bzr rpm xz \
	&& mkdir /clair2.0.1/ 

COPY clair /clair2.0.1/

VOLUME /config
EXPOSE 6060 6061	

RUN chmod u+x /clair2.0.1/clair

ENTRYPOINT ["/clair2.0.1/clair"]
