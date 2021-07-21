#!/usr/bin/env bash

set -eu

# https://docs.docker.com/storage/volumes/#backup-restore-or-migrate-data-volumes
# https://stackoverflow.com/questions/46349071/commit-content-of-mounted-volumes-as-well

DOCKER_CONTAINER=$1
NETWORK_ID=$2

docker run \
--rm \
--volumes-from ${DOCKER_CONTAINER} \
-v $(pwd)/backup/${NETWORK_ID}/ledger:/root/backup/ledger \
ubuntu tar cvf /root/backup/ledger/backup.tar -C /root .ethereum/
