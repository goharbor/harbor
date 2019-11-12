FROM photon:2.0

RUN tdnf install -y git shadow sudo rpm xz python-xml >>/dev/null\
    && tdnf clean all \
    && groupadd -r -g 10000 clair \
    && useradd --no-log-init -m -g 10000 -u 10000 clair