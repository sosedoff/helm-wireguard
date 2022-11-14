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
ARG COREDNS_VERSION=1.10.0

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

RUN curl -o coredns.tgz -L https://github.com/coredns/coredns/releases/download/v${COREDNS_VERSION}/coredns_${COREDNS_VERSION}_linux_amd64.tgz && \
    tar -zxf coredns.tgz && \
    chmod +x coredns && \
    mv coredns /usr/bin/coredns && \
    rm coredns*

RUN \
  apt-get clean autoclean && \
  apt-get autoremove --yes && \
  rm -rf /var/lib/{apt,dpkg,cache,log}/

COPY --from=build /build/wg-http /usr/bin/wg-http

ADD entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]
CMD ["wg-http"]
