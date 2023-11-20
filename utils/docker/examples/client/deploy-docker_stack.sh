#!/bin/bash

# set -x
set -e -o pipefail

CWD="$(realpath "$(dirname $0)")"

set -a
# shellcheck source-path=.
source "$CWD/.env"
set +a

docker stack up -c "$1" daos_stack
