FROM photon:5.0

RUN tdnf install -y python3 python3-pip python3-PyYAML python3-jinja2 && tdnf clean all
RUN pip3 install pipenv==2022.1.8

#To install only htpasswd binary from photon package httpd
RUN tdnf install -y rpm cpio apr-util
RUN tdnf -y --downloadonly --downloaddir=/tmp/ install httpd && rpm2cpio /tmp/httpd-*.rpm | cpio -ivdm ./usr/bin/htpasswd && rm -f /tmp/*

ENV LANG en_US.UTF-8

WORKDIR /usr/src/app

RUN mkdir -p /harbor_make

COPY Pipfile.lock Pipfile /usr/src/app/
RUN set -ex && pipenv install --deploy --system
COPY . /usr/src/app

ENTRYPOINT [ "python3", "main.py" ]

VOLUME ["/harbor_make"]