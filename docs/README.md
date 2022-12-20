# Yui Relayer

## Load Test

Run the following commands from project root directory.

```bash
# Build relayer, docker images of tendermint
make setup-load

# Run tendermint node
make run-tm

# Run scenario, Count means number of message, NAME means log file name
make scenario COUNT=100 NAME=tx-100

# Tally log file
make tally-log RELAY=relay-tx-100.log ACKS=acks-tx-100.log

# Tally log file for query/tx/internal cost
make tally-all-log RELAY=relay-tx-100.log ACKS=acks-tx-100.log
```

## Load Test using service
- terminal 1
```bash
# Build relayer, docker images of tendermint
make setup-load

# Run tendermint node
make run-tm

# Run service
make load-service
```

- terminal 2
```bash
# Keep transferring msgs and seeing unrelayed packets 
make scenario-service-transfer MSG=1 INTERVAL=10
```