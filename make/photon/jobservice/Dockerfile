FROM library/photon:1.0

RUN mkdir /harbor/
COPY ./make/dev/jobservice/harbor_jobservice /harbor/

RUN chmod u+x /harbor/harbor_jobservice
WORKDIR /harbor/
ENTRYPOINT ["/harbor/harbor_jobservice"]
