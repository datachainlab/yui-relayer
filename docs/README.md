# Yui Relayer

## Load Test

Run the following commands from project root directory.

```bash
# Build relayer, docker images of tendermint
setup-load

# Run tendermint node
run-tm

# Run scenario, Count means number of message, NAME means log file name
make scenario COUNT=100 NAME=tx-100

# Tally log file
make tally-log RELAY=relay-tx-100.log ACKS=acks-tx-100.log
```

