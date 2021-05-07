#!/bin/sh

docker images -a --filter label=DAOS=true

IMAGE=$(docker images -a --filter label=DAOS=true --quiet | tail -n 1)

docker run --name build-post --mount type=tmpfs,destination=/mnt/daos,tmpfs-mode=1777 $IMAGE ./utils/run_in_ga.sh
