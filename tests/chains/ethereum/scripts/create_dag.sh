#!/usr/bin/env bash

set -eu

DAG_DIR=$1
GETH_VERSION=$2

if [ -d $DAG_DIR ]; then
  if [ "$(ls -A $DAG_DIR)" ]; then
       echo "$DAG_DIR is not empty"
       exit 0
  fi
else
  echo "$DAG_DIR is not exists"
  mkdir -p $DAG_DIR
fi

docker run --rm \
-v $DAG_DIR:/root/dag \
ethereum/client-go:${GETH_VERSION} makedag 0 /root/dag

chains=(chain0 chain1)
for chain in ${chains[@]}; do
  mkdir -p $DAG_DIR/${chain}/
  cp $DAG_DIR/full-* $DAG_DIR/${chain}/
done

rm -rf $DAG_DIR/full-*

