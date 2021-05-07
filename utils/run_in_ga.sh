#!/bin/bash

set -e

SCONS=scons

if [ ! -e /usr/bin/scons ]
then
    SCONS=scons-3
fi

echo ::group::Rebuild spdk
rm -rf /opt/daos/prereq/release/spdk
$SCONS PREFIX=/opt/daos --build-deps=yes --deps-only
echo ::endgroup::

echo ::group::Build type debug.
$SCONS --jobs 10 PREFIX=/opt/daos COMPILER=clang TARGET_TYPE=release BUILD_TYPE=debug
echo ::endgroup::

cat daos.conf

echo ::group::Install debug
$SCONS install
echo ::endgroup::

echo ::group::Setting up daos_admin
. utils/sl/setup_local.sh
./utils/setup_daos_admin.sh
echo ::endgroup::

echo ::group::Container copy test
export LD_LIBRARY_PATH=/opt/daos/lib:/opt/daos/lib64:/opt/daos/bin/../lib64/daos_srv/../../prereq/release/spdk/lib/:$LD_LIBRARY_PATH
./utils/node_local_test.py --no-root --test cont_copy
echo ::endgroup::

#echo ::group::Key Value test
#./utils/node_local_test.py --no-root kv
#echo ::endgroup::

# Allow time for sockets to settle down.
sleep 5

echo ::group::Fault injection test
./utils/node_local_test.py --no-root fi-core
echo ::endgroup::


echo ::group::Build type dev.
$SCONS -c
$SCONS -c install
$SCONS --jobs 10 PREFIX=/opt/daos COMPILER=clang TARGET_TYPE=release BUILD_TYPE=dev
echo ::endgroup::

echo ::group::Install dev
$SCONS install
echo ::endgroup::
