FROM photon:2.0

RUN tdnf install sudo -y >> /dev/null\
    && tdnf clean all \
    && groupadd -r -g 10000 harbor && useradd --no-log-init -r -g 10000 -u 10000 harbor

COPY ./make/photon/jobservice/harbor_jobservice /harbor/

RUN chmod u+x /harbor/harbor_jobservice

WORKDIR /harbor/

USER harbor

VOLUME ["/var/log/jobs/"]

ENTRYPOINT ["/harbor/harbor_jobservice", "-c", "/etc/jobservice/config.yml"]
