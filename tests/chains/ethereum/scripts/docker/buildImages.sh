#!/usr/bin/env bash

set -eu

DOCKER_BUILD="docker build --rm --no-cache --pull"

DOCKER_REPO=$1
DOCKER_TAG=$2
GETH_VERSION=$3
NETWORK_ID0=$4
NETWORK_ID1=$5

## chain0
${DOCKER_BUILD} -f ./Dockerfiles/geth/Dockerfile \
    --build-arg GETH_VERSION=${GETH_VERSION} \
    --build-arg NETWORKID=${NETWORK_ID0} \
    --build-arg LEDGER_BACKUP_PATH=backup/chain0/ledger \
		--tag ${DOCKER_REPO}ethereum-chain0:${DOCKER_TAG} .

## chain1
${DOCKER_BUILD} -f ./Dockerfiles/geth/Dockerfile \
    --build-arg GETH_VERSION=${GETH_VERSION} \
    --build-arg NETWORKID=${NETWORK_ID1} \
    --build-arg LEDGER_BACKUP_PATH=backup/chain1/ledger \
		--tag ${DOCKER_REPO}ethereum-chain1:${DOCKER_TAG} .
