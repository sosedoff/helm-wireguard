# ------------------------------------------------------------------------------
# Builder Stage
# ------------------------------------------------------------------------------
FROM golang:1.19-bullseye AS build

WORKDIR /build
ADD util/* /build

RUN go build -o wg-http

# ------------------------------------------------------------------------------
# Release Stage
# ------------------------------------------------------------------------------
FROM debian:bullseye-slim

RUN apt-get update

RUN apt install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    software-properties-common \
    net-tools \
    wireguard \
    dnsutils \
    iptables \
    iproute2 \
    procps

RUN \
  apt-get clean autoclean && \
  apt-get autoremove --yes && \
  rm -rf /var/lib/{apt,dpkg,cache,log}/

COPY --from=build /build/wg-http /usr/bin/wg-http

CMD ["wg-http"]
