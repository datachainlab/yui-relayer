include ../../docker.mk

.PHONY: network
network:
	TAG=${DOCKER_TAG} $(DOCKER_COMPOSE) \
		-f ../docker-compose-test.yaml \
		up -d \
		tendermint-chain0 tendermint-chain1

.PHONY: network-notempty
network-notempty:
	TAG=${DOCKER_NOTEMPTY_TAG} $(DOCKER_COMPOSE) \
		-f ../docker-compose-test.yaml \
		up -d \
		tendermint-chain0 tendermint-chain1

.PHONY: network-down
network-down:
	TAG=${DOCKER_TAG} $(DOCKER_COMPOSE) \
		-f ../docker-compose-test.yaml \
		down --volume --remove-orphans

.PHONY: network-notempty-down
network-notempty-down:
	TAG=${DOCKER_NOTEMPTY_TAG} $(DOCKER_COMPOSE) \
		-f ../docker-compose-test.yaml \
		down --volumes --remove-orphans

.PHONY: init-test
init-test:
	./scripts/fixture
	./scripts/init-rly
	./scripts/handshake

.PHONY: test
test: init-test
	./scripts/test-tx

#------------------------------------------------------------------------------
# Load test
#------------------------------------------------------------------------------

# e.g. `make transfer COUNT=10`
.PHONY: transfer
transfer:
	./scripts/transfer $(COUNT)

.PHONY: load-service
load-service:
	./scripts/load-service

# e.g. make loop-transfer MSG=1 INTERVAL=10
.PHONY: loop-transfer
loop-transfer:
	./scripts/loop-transfer $(MSG) $(INTERVAL)

.PHONY: load-shot
load-shot:
	./scripts/load-shot

.PHONY: tx-relay
tx-relay:
	./scripts/tx-relay $(NAME)

.PHONY: tx-acks
tx-acks:
	./scripts/tx-acks $(NAME)

.PHONY: query-unrelay
query-unrelay:
	./scripts/query-unrelay

# e.g. make wait-unrelay MODE=unrelay TARGET=src COUNT=100
.PHONY: wait-unrelay
wait-unrelay:
	./scripts/wait-unrelay $(MODE) $(TARGET) $(COUNT)

# e.g. make sum-relay NAME=relay-tx-100.log
.PHONY: sum-relay
sum-relay:
	#./scripts/sum relay-tx-100.log RelayPackets
	#./scripts/sum relay-tx-100.log Prover.relayPackets
	./scripts/sum $(NAME) RelayMsgs.Send
	./scripts/sum $(NAME) syncHeaders.GetHeader
	./scripts/sum $(NAME) tendermint.GetAddress
	./scripts/sum $(NAME) tendermint.QueryPacket
	./scripts/sum $(NAME) syncHeaders.GetProvableHeight
	./scripts/sum $(NAME) Prover.QueryPacketCommitmentWithProof
	./scripts/sum $(NAME) tendermint.QueryTxs
	./scripts/sum $(NAME) FindPacketFromEventsBySequence
	./scripts/sum $(NAME) Client.TxSearch

# e.g. make sum-acks NAME=acks-tx-100.log
.PHONY: sum-acks
sum-acks:
	./scripts/sum $(NAME) RelayAcknowledgements
	./scripts/sum $(NAME) RelayMsgs.Send
	./scripts/sum $(NAME) syncHeaders.GetHeader
	./scripts/sum $(NAME) tendermint.GetAddress
	./scripts/sum $(NAME) Prover.QueryPacketAcknowledgementCommitmentWithProof
	./scripts/sum $(NAME) syncHeaders.GetProvableHeight
	./scripts/sum $(NAME) tendermint.QueryTxs
	./scripts/sum $(NAME) Client.TxSearch
	./scripts/sum $(NAME) FindPacketAcknowledgementFromEventsBySequence
	./scripts/sum $(NAME) "tendermint.QueryPacket()"
	./scripts/sum $(NAME) tendermint.QueryPacketAcknowledgement
	./scripts/sum $(NAME) FindPacketFromEventsBySequence
	./scripts/sum $(NAME) UnrelayedAcknowledgements

.PHONY: clean
clean:
	rm -rf *.log