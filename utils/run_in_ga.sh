#!/bin/bash

set -x

scons --help

scons --jobs 10 install PREFIX=/opt/daos COMPILER=clang TARGET_TYPE=debug BUILD_TYPE=dev

. utils/sl/setup_local.sh
./utils/setup_daos_admin.sh

./utils/node_local_test.py --no-root kv
./utils/node_local_test.py --no-root fi

