FROM photon:2.0

RUN tdnf install -y sudo >>/dev/null\
    && tdnf clean all \
    && mkdir /clair-adapter/ \
    && groupadd -r -g 10000 clair-adapter \
    && useradd --no-log-init -m -r -g 10000 -u 10000 clair-adapter