# Wireguard Helm Chart

Install Wireguard in your K8s cluster with Helm.

_This is experimental software, things might not work_

What does it do:

- Creates a new wireguard service Deployment (or DaemonSet) in your K8s cluster.
- Runs a http server for health checking and exposing prometheus metrics.
- Runs a DNS server (CoreDNS) so that client peers can resolve internal hostnames.
- Defines LoadBalancer resource so the setup could be scaled up (depends on CloudProvider).

What it does not do:

- Manage peer configuration for you. You have to generate and maintain peer keys on your own.
- Handle peer-to-peer communication when multiple Pods are deployed. This has to do with how wireguard system works.

## Installation

Add helm chart repo:

```bash
helm repo add wireguard https://raw.githubusercontent.com/sosedoff/helm-wireguard/main/repo/
helm repo update
```

Create `values.yaml` file (use example below) and run install chart:

```bash
helm upgrade wireguard wireguard/wireguard \
  --install \
  --atomic \
  --create-namespace \
  --namespace wireguard \
  --values ./values.yaml
```

Once you installation is complete, you should be able to obtain LoadBalancer external
IP address and use it your client configuration.

Get services in the `wireguard` namespace:

```
kubectl get services -n wireguard
NAME           TYPE           CLUSTER-IP       EXTERNAL-IP       PORT(S)                              AGE
wireguard      ClusterIP      10.xxx.xxx.xxx   <none>            8080/TCP,51820/UDP,9090/TCP,53/UDP   10d
wireguard-lb   LoadBalancer   10.xxx.xxx.xxx   174.xxx.xxx.xxx   8080:30553/TCP,51820:30115/UDP       10d
```

Where `174.xxx.xxx.xxx` is your wireguard endpoint.

## Values

```yaml
network: 10.10.0.1/24 # <- your wireguard server peer address
privateKey: <WG_PRIVATE_KEY>
publicKey: <WG_PUBLIC_KEY>

# By default we will use DaemonSet type of deployment to run wireguard server peer
# on each K8s node. To switch to regular deployment uncomment:
# replicas: 1
# deployKind: Deployment

loadbalancer:
  annotations:
    # These annotations are specific to DigitalOcean but should be similar in other cloud providers.
    service.beta.kubernetes.io/do-loadbalancer-name: "k8s-wireguard"
    service.beta.kubernetes.io/do-loadbalancer-size-unit: "1"
    service.beta.kubernetes.io/do-loadbalancer-healthcheck-port: "8080"
    service.beta.kubernetes.io/do-loadbalancer-healthcheck-protocol: "http"
    service.beta.kubernetes.io/do-loadbalancer-healthcheck-path: "/health"

peers:
  user1:
    privateKey: <...>
    publicKey: <...>
    src: 10.10.0.2/32
  user2:
    privateKey: <...>
    publicKey: <...>
    src: 10.10.0.3/32
```

You would need to install `wireguard-tools` to create new wireguard private/public keys.

## Running

A few things to consider when running this setup:

- All wireguard features will work just find when running in a single server peer node
- P2P communication between clients will not work if they're connected to different server peers. Server pods will run using exact same configuration, however wireguard connections require handshake prior to use and will not work if peer is bounced to another server peer.
- Connection timeouts will occur when client peers are bounced between server pods via loadbalancer.

## Metrics

To see which metrics are available, query `http://localhost:9090/metrics` in the pod,
or if you're connected using wireguard, query the service endpoint `http://wireguard.wireguard.svc.cluster.local:9090/metrics`.

Example:

```prometheus
# HELP wireguard_bytes_received_total Total number of bytes received on the interface
# TYPE wireguard_bytes_received_total gauge
wireguard_bytes_received_total{interface="wg0"} 1.0852592e+07
# HELP wireguard_bytes_sent_total Total number of bytes sent on the interface
# TYPE wireguard_bytes_sent_total gauge
wireguard_bytes_sent_total{interface="wg0"} 1.26899624e+08
# HELP wireguard_peer_bytes Total number of bytes transmitted to this peer
# TYPE wireguard_peer_bytes gauge
wireguard_peer_bytes{interface="wg0",peer="RzMpiw1CbCIA1EOPVJpFv/mJJ+VFxSZTgLY3Fc1M4Vo="} 1.26899624e+08
wireguard_peer_bytes{interface="wg0",peer="yEwuNeSqRZYZ8kjD9okgJP6YJaFKxNXD2dGqxVhLlGk="} 0
# HELP wireguard_peer_last_handshake_seconds Number of seconds since last handshake from this peer
# TYPE wireguard_peer_last_handshake_seconds gauge
wireguard_peer_last_handshake_seconds{interface="wg0",peer="RzMpiw1CbCIA1EOPVJpFv/mJJ+VFxSZTgLY3Fc1M4Vo="} 80.428219818
# HELP wireguard_peer_last_receive_seconds Number of seconds since data received from this peer
# TYPE wireguard_peer_last_receive_seconds gauge
wireguard_peer_last_receive_seconds{interface="wg0",peer="RzMpiw1CbCIA1EOPVJpFv/mJJ+VFxSZTgLY3Fc1M4Vo="} 0
# HELP wireguard_peer_last_trasmit_seconds Number of seconds since data trasmitted to this peer
# TYPE wireguard_peer_last_trasmit_seconds gauge
wireguard_peer_last_trasmit_seconds{interface="wg0",peer="RzMpiw1CbCIA1EOPVJpFv/mJJ+VFxSZTgLY3Fc1M4Vo="} 0
# HELP wireguard_peer_rx_bytes Total number of bytes received from this peer
# TYPE wireguard_peer_rx_bytes gauge
wireguard_peer_rx_bytes{interface="wg0",peer="RzMpiw1CbCIA1EOPVJpFv/mJJ+VFxSZTgLY3Fc1M4Vo="} 1.0852592e+07
wireguard_peer_rx_bytes{interface="wg0",peer="yEwuNeSqRZYZ8kjD9okgJP6YJaFKxNXD2dGqxVhLlGk="} 0
# HELP wireguard_peers_active_count Total number of active peers
# TYPE wireguard_peers_active_count gauge
wireguard_peers_active_count{interface="wg0"} 1
# HELP wireguard_peers_count Total number of peers in the interface configuration
# TYPE wireguard_peers_count gauge
wireguard_peers_count{interface="wg0"} 2
```
