FROM photon:2.0

RUN mkdir /harbor/ \
    && tdnf install sudo -y >> /dev/null\
    && tdnf clean all \
    && groupadd -r -g 10000 harbor && useradd --no-log-init -r -g 10000 -u 10000 harbor 

COPY ./make/photon/jobservice/start.sh ./make/photon/jobservice/harbor_jobservice /harbor/

RUN chmod u+x /harbor/harbor_jobservice /harbor/start.sh
RUN mkdir -p /var/log/jobs
WORKDIR /harbor/
ENTRYPOINT ["/harbor/start.sh"]
