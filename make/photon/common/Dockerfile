FROM photon:5.0

# 1. Install photon-repos package
RUN tdnf install photon-repos -y

# 2. Enable photon-snapshot repo
RUN sed -i 's/enabled\s*=.*/enabled=0/g' /etc/yum.repos.d/*.repo
RUN sed -i 's/enabled\s*=.*/enabled=1/g' /etc/yum.repos.d/photon-snapshot.repo && tdnf clean all

# 2. Use tdnf to disable all default repos, enable only the snapshot, and update/install
# The --disablerepo="*" flag turns off the mainline public repos
# The --enablerepo="photon-snapshot" flag explicitly targets your template
RUN tdnf --disablerepo="*" --enablerepo="photon-snapshot" makecache \
    && tdnf --disablerepo="*" --enablerepo="photon-snapshot" update -y
