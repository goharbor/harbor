FROM mysql:5.6

MAINTAINER bhe@vmware.com

RUN sed -i -e 's/us.archive.ubuntu.com/archive.ubuntu.com/g' /etc/apt/sources.list

RUN apt-get update

RUN apt-get install -y curl python python-pip git python-mysqldb

RUN pip install alembic

RUN mkdir -p /harbor-migration

WORKDIR /harbor-migration

COPY ./ ./

RUN ./prepare.sh

ENTRYPOINT ["./run.sh"] 
