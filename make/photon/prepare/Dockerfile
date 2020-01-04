FROM photon:3.0

ENV LANG en_US.UTF-8

WORKDIR /usr/src/app

RUN mkdir -p /harbor_make

RUN tdnf install -y python3 \
    && tdnf install -y python3-pip
RUN pip3 install pipenv==2018.11.26

COPY make/photon/prepare /usr/src/app
RUN set -ex && pipenv install --deploy --system

ENTRYPOINT [ "python3", "main.py" ]

VOLUME ["/harbor_make"]