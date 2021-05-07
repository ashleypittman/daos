#!/bin/sh

docker images -a --filter label=DAOS=true

IMAGE=$(docker images -a --filter label=DAOS=true --quiet | tail -n 1)

# 7Gb VM, 2g for docker, 2g for tmpfs.
docker run --name build-post --memory=2g --mount type=tmpfs,destination=/mnt/daos,tmpfs-mode=1777,tmpfs-size=2097152 $IMAGE ./utils/run_in_ga.sh
