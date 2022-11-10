DOCKER_REPO     ?= sosedoff/wireguard
DOCKER_TAG      ?= $(shell date -u +"%Y%m%d")
DOCKER_IMAGE    ?= ${DOCKER_REPO}:${DOCKER_TAG}
DOCKER_PLATFORM ?= linux/$(shell uname -m)

.PHONY: build
build:
	docker build --platform=${DOCKER_PLATFORM} -t ${DOCKER_REPO}:latest -t ${DOCKER_IMAGE} .

.PHONY: release
release: DOCKER_PLATFORM = "linux/amd64"
release: build
	docker push ${DOCKER_IMAGE}
	docker push ${DOCKER_REPO}:latest

.PHONY: shell
shell:
	docker run \
		--cap-add=NET_ADMIN \
		-v $(shell pwd)/test:/etc/wireguard \
		-p 51820:51820/udp \
		-it --rm \
		${DOCKER_IMAGE} bash

.PHONY: package
package:
	helm package chart -d repo
	helm repo index repo
