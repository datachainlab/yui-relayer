.PHONY: build
build:
	go build -o ./build/uly .

.PHONY: test
test:
	go test -v ./...

.PHONY: proto-gen
proto-gen:
	@echo "Generating Protobuf files"
	docker run -v $(CURDIR):/workspace --workdir /workspace tendermintdev/sdk-proto-gen sh ./scripts/protocgen.sh

#------------------------------------------------------------------------------
# For Load Test
#------------------------------------------------------------------------------
# setup
.PHONY: setup-load
setup-load: build
	make -C tests/chains/tendermint docker-images
	make -C tests/chains/tendermint docker-images-not-empty-block


# run network
.PHONY: run-tm
run-tm:
	make -C tests/cases/tm2tm network-down
	make -C tests/cases/tm2tm network
	make -C tests/cases/tm2tm init-test

.PHONY: run-tm-not-empty
run-tm-not-empty:
	make -C tests/cases/tm2tm network-notempty-down
	make -C tests/cases/tm2tm network-notempty
	make -C tests/cases/tm2tm init-test


# load
#e.g. make transfer COUNT=1000
.PHONY: transfer
transfer:
	make -C tests/cases/tm2tm transfer COUNT=$(COUNT)

.PHONY: transfer1
transfer1:
	make -C tests/cases/tm2tm transfer COUNT=1

.PHONY: transfer10
transfer10:
	make -C tests/cases/tm2tm transfer COUNT=10

.PHONY: transfer100
transfer100:
	make -C tests/cases/tm2tm transfer COUNT=100

.PHONY: transfer1000
transfer1000:
	make -C tests/cases/tm2tm transfer COUNT=1000

.PHONY: load-service
load-service:
	make -C tests/cases/tm2tm load-service

.PHONY: load-shot
load-shot:
	make -C tests/cases/tm2tm load-shot

.PHONY: tx-relay
tx-relay:
	make -C tests/cases/tm2tm tx-relay PARAM=name

.PHONY: tx-acks
tx-acks:
	make -C tests/cases/tm2tm tx-acks PARAM=name


# scenario
# e.g. make scenario COUNT=100 NAME=tx-100
.PHONY: scenario
scenario: clean
	#make -C tests/cases/tm2tm scenario
	make transfer COUNT=$(COUNT)
	# wait unrelay
	make -C tests/cases/tm2tm wait-unrelay MODE=unrelay TARGET=src
	make -C tests/cases/tm2tm tx-relay NAME=$(NAME)
	# wait unrelay-akcs
	make -C tests/cases/tm2tm wait-unrelay MODE=acks TARGET=dst
	make -C tests/cases/tm2tm tx-acks NAME=$(NAME)

# e.g. make scenario-service COUNT=10000 NAME=service-tx-100
.PHONY: scenario-service
scenario-service:
	make transfer COUNT=$(COUNT)
	make -C tests/cases/tm2tm load-service

# e.g. make tally-log RELAY=relay-tx-100.log ACKS=acks-tx-100.log
tally-log:
	make -C tests/cases/tm2tm sum-relay NAME=$(RELAY)
	make -C tests/cases/tm2tm sum-acks NAME=$(ACKS)

# utils
.PHONY: query-unrelay
query-unrelay:
	make -C tests/cases/tm2tm query-unrelay

.PHONY: clean
clean:
	make -C tests/cases/tm2tm clean
