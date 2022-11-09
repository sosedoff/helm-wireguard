# Wireguard Helm Chart

TODO: Writeme

## Values

```yaml
network: 10.10.0.0/24
privateKey: <WG_PRIVATE_KEY>
publicKey: <WG_PUBLIC_KEY>
loadbalancer:
  annotations:
    service.beta.kubernetes.io/do-loadbalancer-name: "k8s-wireguard"
    service.beta.kubernetes.io/do-loadbalancer-size-unit: "1"
    service.beta.kubernetes.io/do-loadbalancer-healthcheck-port: "8080"
    service.beta.kubernetes.io/do-loadbalancer-healthcheck-protocol: "http"
    service.beta.kubernetes.io/do-loadbalancer-healthcheck-path: "/health"
peers:
  user1:
    privateKey: <...>
    publicKey: <...>
    src: 10.10.0.1/32
  user1:
    privateKey: <...>
    publicKey: <...>
    src: 10.10.0.2/32
```
