FROM python:2
MAINTAINER int32bit krystism@gmail.com

ADD . /opt/harborclient
RUN pip install -r /opt/harborclient/requirements.txt
RUN set -ex \
        && cd /opt/harborclient \
        && python setup.py install \
        && rm -rf /opt/harborclient
CMD ["harbor"]
