FROM photon:5.0

# 1. Copy your customized snapshot repo file into the container
COPY ./make/photon/common/photon-snapshot.repo /etc/yum.repos.d/photon-snapshot.repo

# 2. Use tdnf to disable all default repos, enable only the snapshot, and update/install
# The --disablerepo="*" flag turns off the mainline public repos
# The --enablerepo="photon-snapshot" flag explicitly targets your template
# TODO: update the package list when the package has CVE and need to be updated
RUN tdnf --disablerepo="*" --enablerepo="photon-snapshot" makecache \
    && tdnf --disablerepo="*" --enablerepo="photon-snapshot" distro-sync -y \
    && tdnf --disablerepo="*" --enablerepo="photon-snapshot" update -y