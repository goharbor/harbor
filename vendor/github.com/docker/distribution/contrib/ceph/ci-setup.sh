#! /bin/bash
#
# Ceph cluster setup in Circle CI
#

set -x
set -e
set -u

NODE=$(hostname)
CEPHDIR=/tmp/ceph

mkdir cluster
pushd cluster

# Install
retries=0
until [ $retries -ge 5 ]; do
  pip install ceph-deploy && break
  retries=$[$retries+1]
  sleep 30
done

retries=0
until [ $retries -ge 5 ]; do
    # apt-get can get stuck and hold the lock in some circumstances
    # so preemptively kill it
    kill `pgrep apt-get` || true
  ceph-deploy install --release hammer $NODE && break
  retries=$[$retries+1]
  sleep 30
done

retries=0
until [ $retries -ge 5 ]; do
  ceph-deploy pkg --install librados-dev $NODE && break
  retries=$[$retries+1]
  sleep 30
done

echo $(ip route get 1 | awk '{print $NF;exit}') $(hostname) >> /etc/hosts
ssh-keygen -t rsa -f ~/.ssh/id_rsa -q -N ""
cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys
ssh-keyscan $NODE >> ~/.ssh/known_hosts
ceph-deploy new $NODE

cat >> ceph.conf <<EOF
osd objectstore = memstore
memstore device bytes = 2147483648
osd data = $CEPHDIR
osd journal = $CEPHDIR/journal
osd crush chooseleaf type = 0
osd pool default size = 1
osd pool default min size = 1
osd scrub load threshold = 1000

debug_lockdep = 0/0
debug_context = 0/0
debug_crush = 0/0
debug_buffer = 0/0
debug_timer = 0/0
debug_filer = 0/0
debug_objecter = 0/0
debug_rados = 0/0
debug_rbd = 0/0
debug_journaler = 0/0
debug_objectcatcher = 0/0
debug_client = 0/0
debug_osd = 0/0
debug_optracker = 0/0
debug_objclass = 0/0
debug_filestore = 0/0
debug_journal = 0/0
debug_ms = 0/0
debug_monc = 0/0
debug_tp = 0/0
debug_auth = 0/0
debug_finisher = 0/0
debug_heartbeatmap = 0/0
debug_perfcounter = 0/0
debug_asok = 0/0
debug_throttle = 0/0
debug_mon = 0/0
debug_paxos = 0/0
debug_rgw = 0/0
osd_op_num_threads_per_shard = 1 //You may want to try with 1 as well
osd_op_num_shards = 5    //Depends on your cpu util
ms_nocrc = true
cephx_sign_messages = false
cephx_require_signatures = false
ms_dispatch_throttle_bytes = 0
throttler_perf_counter = false

[osd]
osd_client_message_size_cap = 0
osd_client_message_cap = 0
osd_enable_op_tracker = false
EOF

sed -i -r 's/mon_host =.*/mon_host = 127.0.0.1/' ceph.conf
sed -i -r 's/auth_cluster_required =.*/auth_cluster_required = none/' ceph.conf
sed -i -r 's/auth_service_required =.*/auth_service_required = none/' ceph.conf
sed -i -r 's/auth_client_required =.*/auth_client_required = none/' ceph.conf

# Setup monitor and keyrings
ceph-deploy mon create-initial
ceph-deploy admin $NODE
sudo chmod a+r /etc/ceph/ceph.client.admin.keyring

# Setup OSD
mkdir -p $CEPHDIR
OSD=$(ceph osd create)
ceph osd crush add osd.${OSD} 1 root=default host=$NODE
ceph-osd --id ${OSD} --mkjournal --mkfs
ceph-osd --id ${OSD}

# Status
ceph status
ceph health detail
ceph osd tree

popd
