#!/usr/bin/env bash

set -eu

DOCKER_BUILD="docker build --rm --no-cache --pull"

GETH_VERSION=$1
DOCKER_REPO=$2
DOCKER_TAG=$3
DOCKER_IMAGE=$4
NETWORK_ID=$5

${DOCKER_BUILD} -f ./Dockerfiles/geth/Dockerfile \
    --build-arg GETH_VERSION=${GETH_VERSION} \
    --build-arg NETWORKID=${NETWORK_ID} \
    --build-arg LEDGER_BACKUP_PATH=backup/${NETWORK_ID}/ledger \
    --build-arg CONTRACT_ADDRESS_DIR=contract/build/addresses/${NETWORK_ID} \
		--tag ${DOCKER_REPO}${DOCKER_IMAGE}:${DOCKER_TAG} .
