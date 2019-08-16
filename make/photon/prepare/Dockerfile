FROM cicd.harbor.bitsf.xin/harbor-depend/prepare:base

ENV LANG en_US.UTF-8

WORKDIR /usr/src/app

RUN mkdir -p /harbor_make

COPY make/photon/prepare /usr/src/app
RUN set -ex && pipenv install --deploy --system

ENTRYPOINT [ "python3", "main.py" ]

VOLUME ["/harbor_make"]

COPY make/harbor.yml /input/harbor.yml
