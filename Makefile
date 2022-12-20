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


#------------------------------------------------------------------------------
# scenario
#------------------------------------------------------------------------------
# e.g. make scenario COUNT=100 NAME=tx-100
.PHONY: scenario
scenario: clean
	date +%s
	#make -C tests/cases/tm2tm scenario
	make transfer COUNT=$(COUNT)
	# wait unrelay
	date +%s
	make -C tests/cases/tm2tm wait-unrelay MODE=unrelay TARGET=src COUNT=$(COUNT)
	date +%s
	make -C tests/cases/tm2tm tx-relay NAME=$(NAME)
	# wait unrelay-akcs
	date +%s
	make -C tests/cases/tm2tm wait-unrelay MODE=acks TARGET=dst COUNT=$(COUNT)
	date +%s
	make -C tests/cases/tm2tm tx-acks NAME=$(NAME)
	date +%s

.PHONY: scenario-all
scenario-all: clean
	#make run-tm-not-empty
	make scenario COUNT=1 NAME=tx-1
	make scenario COUNT=10 NAME=tx-10
	make scenario COUNT=20 NAME=tx-20
	make scenario COUNT=30 NAME=tx-30
	make scenario COUNT=40 NAME=tx-40
	make scenario COUNT=50 NAME=tx-50
	make scenario COUNT=60 NAME=tx-60
	make scenario COUNT=70 NAME=tx-70
	make scenario COUNT=80 NAME=tx-80
	make scenario COUNT=90 NAME=tx-90
	make scenario COUNT=100 NAME=tx-100
	make scenario COUNT=200 NAME=tx-200
	make scenario COUNT=300 NAME=tx-300
	make scenario COUNT=400 NAME=tx-400
	make scenario COUNT=500 NAME=tx-500

.PHONY: scenario-service
scenario-service:
	make -C tests/cases/tm2tm load-service

# e.g. make scenario-service-transfer MSG=1 INTERVAL=10
.PHONY: scenario-service-transfer
scenario-service-transfer:
	make -C tests/cases/tm2tm loop-transfer MSG=$(MSG) INTERVAL=$(INTERVAL)

#------------------------------------------------------------------------------
# tally log
#------------------------------------------------------------------------------
# e.g. make tally-log RELAY=relay-tx-100.log ACKS=acks-tx-100.log
.PHONY: tally-log
tally-log:
	make -C tests/cases/tm2tm sum-relay NAME=$(RELAY)
	make -C tests/cases/tm2tm sum-acks NAME=$(ACKS)

# e.g. make tally-relay-log RELAY=relay-tx-100.log
.PHONY: tally-relay-log
tally-relay-log:
	make -C tests/cases/tm2tm sum-relay NAME=$(RELAY)

# e.g. make tally-acks-log ACKS=acks-tx-100.log
.PHONY: tally-acks-log
tally-acks-log:
	make -C tests/cases/tm2tm sum-acks NAME=$(ACKS)

# e.g. make tally-all-log RELAY=relay-tx-100.log ACKS=acks-tx-100.log
.PHONY: tally-all-log
tally-all-log:
	make -C tests/cases/tm2tm sum-all-relay NAME=$(RELAY)
	make -C tests/cases/tm2tm sum-all-acks NAME=$(ACKS)

# e.g. make tally-relay-all-log RELAY=relay-tx-100.log
.PHONY: tally-relay-all-log
tally-relay-all-log:
	make -C tests/cases/tm2tm sum-all-relay NAME=$(RELAY)

# e.g. make tally-acks-all-log ACKS=acks-tx-100.log
.PHONY: tally-acks-all-log
tally-acks-all-log:
	make -C tests/cases/tm2tm sum-all-acks NAME=$(ACKS)

#------------------------------------------------------------------------------
# utils
#------------------------------------------------------------------------------
.PHONY: query-unrelay
query-unrelay:
	make -C tests/cases/tm2tm query-unrelay

.PHONY: clean
clean:
	make -C tests/cases/tm2tm clean
