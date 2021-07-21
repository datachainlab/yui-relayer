#!/bin/sh

/usr/local/bin/geth \
  --config /root/geth-config.toml \
  --mine --miner.threads 1 \
  --password /root/geth.password \
  --unlock "0" \
  --allow-insecure-unlock \
  $@
