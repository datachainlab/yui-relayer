include ../../docker.mk

SCRIPT_DIR ?= $(CURDIR)/scripts
DAG_DIR ?= $(CURDIR)/dag

GETH_VERSION ?= v1.10.4

NETWORK_ID0 ?= 12018
NETWORK_ID1 ?= 12019

.PHONY: clean
clean:
	rm -rf ./backup
	rm -rf ./dag

.PHONY: dag
dag:
	$(SCRIPT_DIR)/create_dag.sh $(DAG_DIR) $(GETH_VERSION)

.PHONY: deploy-contract
deploy-contract:
	cd contract && npm run migrate

.PHONY: up-scaffold
up-scaffold:
	GETH_VERSION=$(GETH_VERSION) NETWORK_ID0=$(NETWORK_ID0) NETWORK_ID1=$(NETWORK_ID1) \
	$(DOCKER_COMPOSE) up --build -d

.PHONY: pause-scaffold
pause-scaffold:
	$(DOCKER_COMPOSE) pause

.PHONY: unpause-scaffold
unpause-scaffold:
	$(DOCKER_COMPOSE) unpause

.PHONY: down-scaffold
down-scaffold:
	$(DOCKER_COMPOSE) down --volumes --remove-orphans

.PHONY: backup
backup:
	rm -rf $(CURDIR)/backup
	$(SCRIPT_DIR)/docker/backupVolume.sh

.PHONY: snapshot
snapshot:
	make dag
	make up-scaffold
	make deploy-contract
	make pause-scaffold
	make backup
	make unpause-scaffold
	make down-scaffold

.PHONY: docker-images
docker-images:
	$(SCRIPT_DIR)/docker/buildImages.sh $(DOCKER_REPO) $(DOCKER_TAG) $(GETH_VERSION) $(NETWORK_ID0) $(NETWORK_ID1)
