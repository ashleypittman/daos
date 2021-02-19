#!/bin/bash

set -e

echo ::group::Build
scons --jobs 10 install PREFIX=/opt/daos COMPILER=clang TARGET_TYPE=debug BUILD_TYPE=dev
echo ::endgroup::

. utils/sl/setup_local.sh
./utils/setup_daos_admin.sh

echo ::group::Key Value test
./utils/node_local_test.py --no-root kv
echo ::endgroup::
#./utils/node_local_test.py --no-root fi

