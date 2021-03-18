#!/bin/bash

set -e
cd daos

echo ::group::Build
scons --jobs 8 PREFIX=/opt/daos COMPILER=clang TARGET_TYPE=release BUILD_TYPE=dev
echo ::endgroup::

cat daos.conf

echo ::group::Install
scons --jobs 10 install
echo ::endgroup::

echo ::group::Setting up daos_admin
. utils/sl/setup_local.sh
./utils/setup_daos_admin.sh
echo ::endgroup::

echo ::group::Key Value test
ldd /opt/daos/bin/daos_server
echo $LD_LIBRARY_PATH
ls /opt/daos/bin/../lib64/daos_srv/../../prereq/release/spdk/lib/
LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/opt/daos/bin/../lib64/daos_srv/../../prereq/release/spdk/lib/
ldd /opt/daos/bin/daos_server
ifconfig -a || true
./utils/node_local_test.py --no-root kv
echo ::endgroup::

echo ::group::Fault injection test
./utils/node_local_test.py --no-root fi
echo ::endgroup::
