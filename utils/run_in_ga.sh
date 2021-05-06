#!/bin/bash

set -e

# BUILD_TYPE=debug does not work

echo ::group::Build
scons --jobs 10 PREFIX=/opt/daos COMPILER=clang TARGET_TYPE=release BUILD_TYPE=debug
echo ::endgroup::

cat daos.conf

echo ::group::Install
scons --jobs 10 install
echo ::endgroup::

echo ::group::Setting up daos_admin
. utils/sl/setup_local.sh
./utils/setup_daos_admin.sh
echo ::endgroup::

echo ::group::Container copy test
export LD_LIBRARY_PATH=/opt/daos/lib:/opt/daos/lib64:/opt/daos/bin/../lib64/daos_srv/../../prereq/release/spdk/lib/:$LD_LIBRARY_PATH
./utils/node_local_test.py --no-root --test cont_copy
echo ::endgroup::

echo ::group::Key Value test
./utils/node_local_test.py --no-root kv
echo ::endgroup::

echo ::group::Fault injection test
./utils/node_local_test.py --no-root fi
echo ::endgroup::
