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
		--cap-add=NET_BIND_SERVICE \
		-v $(shell pwd)/test:/etc/wireguard \
		-p 51820:51820/udp \
		-p 53:53/udp \
		-p 8080:8080 \
		-p 9090:9090 \
		-it --rm \
		${DOCKER_IMAGE} bash

.PHONY: run
run:
	docker run \
		--cap-add=NET_ADMIN \
		--cap-add=NET_BIND_SERVICE \
		-v $(shell pwd)/test:/etc/wireguard \
		-p 51820:51820/udp \
		-p 8080:8080 \
		-p 9090:9090 \
		-p 5053:53/udp \
		-e WG_ENABLE=1 \
		-e WG_PEER_MONITOR=1 \
		-e WG_PROM_METRICS=1 \
		-e WG_COREDNS=1 \
		-it --rm \
		${DOCKER_IMAGE} wg-http

.PHONY: package
package:
	helm package chart -d repo
	helm repo index repo
