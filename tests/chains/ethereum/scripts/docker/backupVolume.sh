#!/usr/bin/env bash

set -eu

# https://docs.docker.com/storage/volumes/#backup-restore-or-migrate-data-volumes
# https://stackoverflow.com/questions/46349071/commit-content-of-mounted-volumes-as-well

docker run \
--rm \
--volumes-from ethereum-chain0-scaffold \
-v $(pwd)/backup/chain0/ledger:/root/backup/ledger \
ubuntu tar cvf /root/backup/ledger/backup.tar -C /root .ethereum/

docker run \
--rm \
--volumes-from ethereum-chain1-scaffold \
-v $(pwd)/backup/chain1/ledger:/root/backup/ledger \
ubuntu tar cvf /root/backup/ledger/backup.tar -C /root .ethereum/

