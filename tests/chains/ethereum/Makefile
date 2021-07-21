include ../../docker.mk

SCRIPT_DIR ?= $(CURDIR)/scripts

GETH_VERSION    ?= v1.10.4
GANACHE_VERSION ?= v6.12.2

NETWORK_ID0 ?= 12018
NETWORK_ID1 ?= 12019

export GETH_VERSION
export NETWORK_ID0
export NETWORK_ID1

.PHONY: clean
clean:
	rm -rf $(CURDIR)/backup
	rm -rf $(CURDIR)/dag

.PHONY: snapshot
snapshot:
	make up-scaffold
	make deploy-contract
	make pause-scaffold
	make backup
	make save-contract-address
	make down-scaffold

.PHONY: dag
dag:
	$(SCRIPT_DIR)/create_dag.sh $(CURDIR)/dag

.PHONY: deploy-contract
deploy-contract:
	cd contract && npm run migrate

.PHONY: up-scaffold
up-scaffold:
	$(DOCKER_COMPOSE) up --build -d

.PHONY: down-scaffold
down-scaffold:
	$(DOCKER_COMPOSE) down --volumes --remove-orphans

.PHONY: pause-scaffold
pause-scaffold:
	$(DOCKER_COMPOSE) pause

.PHONY: unpause-scaffold
unpause-scaffold:
	$(DOCKER_COMPOSE) unpause

.PHONY: save-contract-address
save-contract-address:
	$(SCRIPT_DIR)/docker/saveContractAddress.sh $(NETWORK_ID0)
	$(SCRIPT_DIR)/docker/saveContractAddress.sh $(NETWORK_ID1)

.PHONY: backup
backup:
	$(SCRIPT_DIR)/docker/backupVolume.sh ethereum-chain0-scaffold $(NETWORK_ID0)
	$(SCRIPT_DIR)/docker/backupVolume.sh ethereum-chain1-scaffold $(NETWORK_ID1)

.PHONY: docker-images
docker-images:
	$(SCRIPT_DIR)/docker/buildImages.sh  $(GETH_VERSION) $(DOCKER_REPO) $(DOCKER_TAG) ethereum-chain0 $(NETWORK_ID0)
	$(SCRIPT_DIR)/docker/buildImages.sh  $(GETH_VERSION) $(DOCKER_REPO) $(DOCKER_TAG) ethereum-chain1 $(NETWORK_ID1)

.PHONY: ganache
ganache:
	docker run -it --rm --name ganache \
	-e NODE_OPTIONS="--max-old-space-size=4096" \
	-p 8545:8545 \
	trufflesuite/ganache-cli:$(GANACHE_VERSION) \
	--networkId 5777 --defaultBalanceEther 10000 \
	--mnemonic "math razor capable expose worth grape metal sunset metal sudden usage scheme"
