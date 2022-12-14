DOCKER         ?= docker
DOCKER_COMPOSE ?= docker-compose
DOCKER_REPO    ?= ""
DOCKER_TAG     ?= latest
DOCKER_NOTEMPTY_TAG ?= not-empty
DOCKER_BUILD   ?= $(DOCKER) build --rm --no-cache --pull

MAKEFILE_DIR:=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))

.PHONY: wait-for-launch
wait-for-launch:
	$(MAKEFILE_DIR)scripts/wait-for-launch $(ATTEMPT) $(CONTAINER)
