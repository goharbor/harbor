FROM photon:5.0

# Build Harbor's base images from the rolling photon + photon-updates repos
# instead of the pinned photon-snapshot repo.
#
# Photon's photon-snapshot repo pins a point-in-time package set that is
# inconsistent across architectures: on x86_64 it carries postgresql 16/17/18
# but not 15, while on aarch64 it carries postgresql 13-17 but not 18. Relying
# on it makes the multi-arch base image build fail for one architecture or the
# other (goharbor/harbor#23324, #23309).
#
# The photon (base) and photon-updates repos share the same per-arch baseurl as
# photon-snapshot but are not version-pinned, so together they expose the full,
# consistent package set (e.g. postgresql 13-18, including both the 15 needed
# for pg_upgrade and the 18 runtime) on both amd64 and arm64. Enabling the base
# photon repo as well guarantees no package that used to come from the snapshot
# goes missing.
RUN tdnf install photon-repos -y \
    && sed -i 's/enabled\s*=.*/enabled=0/g' /etc/yum.repos.d/*.repo \
    && sed -i 's/enabled\s*=.*/enabled=1/g' /etc/yum.repos.d/photon.repo \
    && sed -i 's/enabled\s*=.*/enabled=1/g' /etc/yum.repos.d/photon-updates.repo \
    && tdnf clean all \
    && tdnf --disablerepo="*" --enablerepo="photon" --enablerepo="photon-updates" makecache \
    && tdnf --disablerepo="*" --enablerepo="photon" --enablerepo="photon-updates" update -y \
    && tdnf clean all
