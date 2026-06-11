FROM photon:5.0

RUN tdnf install photon-repos -y \
    && sed -i 's/enabled\s*=.*/enabled=0/g' /etc/yum.repos.d/*.repo \
    && sed -i 's/enabled\s*=.*/enabled=1/g' /etc/yum.repos.d/photon-snapshot.repo \
    && tdnf clean all \
    && tdnf --disablerepo="*" --enablerepo="photon-snapshot" makecache \
    && tdnf --disablerepo="*" --enablerepo="photon-snapshot" update -y \
    && tdnf clean all
