#!/bin/bash

set -e

echo ::group::Build type debug.
scons --jobs 10 PREFIX=/opt/daos COMPILER=clang TARGET_TYPE=release BUILD_TYPE=debug
echo ::endgroup::

cat daos.conf

echo ::group::Install debug
scons install
echo ::endgroup::

echo ::group::Setting up daos_admin
. utils/sl/setup_local.sh
./utils/setup_daos_admin.sh
echo ::endgroup::

echo ::group::Container copy test
export LD_LIBRARY_PATH=/opt/daos/lib:/opt/daos/lib64:/opt/daos/bin/../lib64/daos_srv/../../prereq/release/spdk/lib/:$LD_LIBRARY_PATH
set +e
./utils/node_local_test.py --no-root --test cont_copy
dmesg
set -e
echo ::endgroup::

#echo ::group::Key Value test
#./utils/node_local_test.py --no-root kv
#echo ::endgroup::

# Allow time for sockets to settle down.
sleep 5

echo ::group::Fault injection test
./utils/node_local_test.py --no-root fi
echo ::endgroup::


echo ::group::Build type dev.
scons -c
scons -c install
scons --jobs 10 PREFIX=/opt/daos COMPILER=clang TARGET_TYPE=release BUILD_TYPE=dev
echo ::endgroup::

echo ::group::Install dev
scons install
echo ::endgroup::
