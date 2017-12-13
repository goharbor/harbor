FROM vmware/photon:1.0

RUN mkdir /harbor/ \
    && tdnf distro-sync -y \
    && tdnf install sudo -y \
    && tdnf clean all \
    && groupadd -r -g 10000 harbor && useradd --no-log-init -r -g 10000 -u 10000 harbor 
HEALTHCHECK CMD curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1:8080/api/jobs/replication/1/log|grep 401

COPY ./make/photon/jobservice/start.sh ./make/dev/jobservice/harbor_jobservice /harbor/

RUN chmod u+x /harbor/harbor_jobservice /harbor/start.sh
WORKDIR /harbor/
ENTRYPOINT ["/harbor/start.sh"]
