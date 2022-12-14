include ../../docker.mk

.PHONY: build
build:
	go build -o ./build/simd ./simapp/simd

.PHONY: docker-images
docker-images:
	$(DOCKER_BUILD) \
		--build-arg CHAINID=ibc0 \
		--tag $(DOCKER_REPO)tendermint-chain0:$(DOCKER_TAG) .
	$(DOCKER_BUILD) \
		--build-arg CHAINID=ibc1 \
		--tag $(DOCKER_REPO)tendermint-chain1:$(DOCKER_TAG) .

.PHONY: docker-images-not-empty-block
docker-images-not-empty-block:
	$(DOCKER_BUILD) \
		--build-arg CHAINID=ibc0 \
		--build-arg EMPTY_BLOCK=false \
		--tag $(DOCKER_REPO)tendermint-chain0:$(DOCKER_NOTEMPTY_TAG) .
	$(DOCKER_BUILD) \
		--build-arg CHAINID=ibc1 \
		--build-arg EMPTY_BLOCK=false \
		--tag $(DOCKER_REPO)tendermint-chain1:$(DOCKER_NOTEMPTY_TAG) .
